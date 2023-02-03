package broker

import (
	"context"
	"errors"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/weilinfox/youmu-thlink/utils"

	"github.com/quic-go/quic-go"
	"github.com/sirupsen/logrus"
)

var logger = logrus.WithField("broker", "internal")

var peers = make(map[int]int)

func Main(listenAddr string, upperAddr string) {

	var upperAddress string // upper
	var upperStatus = 0     // upper broker 0 health, >0 retry times
	var selfPort int        // self port
	var newBrokers sync.Map // 1 jump string=>time.Time
	var netBrokers sync.Map // >1 jump string=>time.Time

	_, slistenPort, err := net.SplitHostPort(listenAddr)
	if err != nil {
		logger.WithError(err).Fatal("Address split error")
	}
	listenPort64, err := strconv.ParseInt(slistenPort, 10, 32)
	selfPort = int(listenPort64)
	tcpAddr, err := net.ResolveTCPAddr("tcp", listenAddr)
	if err != nil {
		logger.WithError(err).Fatal("Address port parse error")
	}
	if err != nil {
		logger.WithError(err).Fatal("Address resolve error")
	}

	// start udp command interface
	logger.Info("Start tcp command interface at " + tcpAddr.String())
	listener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		logger.WithError(err).Fatal("Address listen failed")
	}
	defer listener.Close()

	// net data syncing
	// TODO: make it more efficient
	go func() {

		for {
			// tell upper broker
			if upperAddr != "" {

				tcpConn, err := net.DialTimeout("tcp", upperAddr, time.Second)

				if err != nil {

					if upperAddress == "" {
						logger.WithError(err).Fatal("Upper broker connect error")
					} else {
						upperStatus++
						logger.WithError(err).WithField("retry", upperStatus).Error("Upper broker connect error")
					}

				} else {

					// logger.Debug("Ping upper addr")
					_, err = tcpConn.Write(utils.NewDataFrame(utils.NET_INFO_UPDATE, []byte{byte(selfPort >> 8), byte(selfPort)}))
					if err != nil {
						if upperAddress == "" {
							logger.WithError(err).Fatal("Send NET_INFO_UPDATE to upper broker error")
						} else {
							upperStatus++
							logger.WithError(err).WithField("retry", upperStatus).Error("Send NET_INFO_UPDATE to upper broker error")
						}
					}

					_ = tcpConn.Close()

					// first time
					if upperAddress == "" {
						upperAddress = tcpConn.RemoteAddr().String()
						upperStatus = 0
						logger.Info("Upper broker connected ", upperAddress)

						// sync broker list
						tcpConn, err := net.DialTimeout("tcp", upperAddr, time.Second)

						if err != nil {

							logger.WithError(err).Fatal("Upper broker connect for sync error")

						} else {

							logger.Info("Sync brokers in thlink network")
							_, err = tcpConn.Write(utils.NewDataFrame(utils.NET_INFO, []byte{byte(selfPort >> 8), byte(selfPort)}))
							if err != nil {
								logger.WithError(err).Fatal("Send NET_INFO to upper broker error")
							}
							// read broker list
							buf := make([]byte, utils.TransBufSize)
							n, err := tcpConn.Read(buf)
							if err != nil {
								logger.WithError(err).Fatal("Read NET_INFO response error")
							}
							// parse broker list
							dataStream := utils.NewDataStream()
							dataStream.Append(buf[:n])
							if !dataStream.Parse() {
								logger.Fatal("Parse NET_INFO response error")
							}
							for i := 0; i < dataStream.Len(); {
								logger.Info("Sync broker: ", string(dataStream.Data()[i+1:i+1+int(dataStream.Data()[i])]))
								netBrokers.Store(string(dataStream.Data()[i+1:i+1+int(dataStream.Data()[i])]), time.Now())
								i += 1 + int(dataStream.Data()[i])
							}

							_ = tcpConn.Close()

						}

					}

				}

			}

			// find 10s timeout broker
			data := []byte{byte(selfPort >> 8), byte(selfPort)}
			newBrokers.Range(func(k, v interface{}) bool {
				if time.Now().Sub(v.(time.Time)).Seconds() > 10 {
					logger.Info("Timeout broker: ", k)
					newBrokers.Delete(k)
					data = append(data, byte(len(k.(string)))|0x80)
					data = append(data, []byte(k.(string))...)
				}
				return true
			})
			if len(data) > 2 {
				// tell 1 jump brokers
				newBrokers.Range(func(k, _ interface{}) bool {
					bkrConn, err := net.DialTimeout("tcp", k.(string), time.Second)
					if err != nil {
						logger.WithError(err).Warn("Send new broker 1 jump broker error")
						return true
					}
					logger.Debug("Send timeout broker data to ", k.(string))
					_, _ = bkrConn.Write(utils.NewDataFrame(utils.NET_INFO_UPDATE, data))
					return true
				})
				// tell upper broker
				if upperAddress != "" {
					bkrConn, err := net.DialTimeout("tcp", upperAddress, time.Second)
					if err != nil {
						logger.WithError(err).Warn("Send new broker to upper broker error")
					} else {
						logger.Debug("Send timeout broker data to upper broker ", upperAddress)
						_, _ = bkrConn.Write(utils.NewDataFrame(utils.NET_INFO_UPDATE, data))
					}
				}
			}

			time.Sleep(time.Second)
		}

	}()

	for {

		buf := make([]byte, utils.CmdBufSize)
		conn, err := listener.Accept()
		if err != nil {
			logger.WithError(err).Error("TCP listen error")
			continue
		}
		n, err := conn.Read(buf)
		if err != nil {
			logger.WithError(err).Error("TCP read failed")
			conn.Close()
			continue
		}

		if n >= utils.CmdBufSize {
			logger.Warn("RawData data too long!")
			conn.Close()
			continue
		}

		// handle commands
		dataStream := utils.NewDataStream()
		dataStream.Append(buf[:n])
		if !dataStream.Parse() {
			logger.Warn("Invalid command")
			continue
		}

		cmdData := dataStream.Data()
		cmdLen := dataStream.Len()
		cmdType := dataStream.Type()
		go func() {
			switch cmdType {
			case utils.PING:
				// ping
				_, err := conn.Write(utils.NewDataFrame(utils.PING, nil))

				if err != nil {
					logger.WithError(err).Error("Send response failed")
				}

			case utils.TUNNEL:
				// new tcp/udp tunnel
				// <type> t/u
				var port1, port2 int
				var err error

				if cmdLen > 1 {
					switch cmdData[0] {
					case 't':
						logger.WithField("host", conn.RemoteAddr().String()).Info("New tcp tunnel")
						host, _, _ := net.SplitHostPort(conn.RemoteAddr().String())
						port1, port2, err = newTcpTunnel(host)
					case 'u':
						logger.WithField("host", conn.RemoteAddr().String()).Info("New udp tunnel")
						host, _, _ := net.SplitHostPort(conn.RemoteAddr().String())
						port1, port2, err = newUdpTunnel(host, cmdData[1])
					default:
						logger.Warn("Invalid tunnel type")
					}

					if err != nil {
						logger.WithError(err).Error("Failed to build new tunnel")
					}
				}

				_, err = conn.Write(utils.NewDataFrame(utils.TUNNEL, []byte{byte(port1 >> 8), byte(port1), byte(port2 >> 8), byte(port2)}))

				if err != nil {
					logger.WithError(err).Error("Send response failed")
				}

			case utils.BROKER_INFO:
				// broker info
				_, err := conn.Write(utils.NewDataFrame(utils.BROKER_INFO, []byte{byte(len(peers) >> 56), byte(len(peers) >> 48), byte(len(peers) >> 40), byte(len(peers) >> 32),
					byte(len(peers) >> 24), byte(len(peers) >> 16), byte(len(peers) >> 8), byte(len(peers))}))

				if err != nil {
					logger.WithError(err).Error("Send response failed")
				}

			case utils.NET_INFO:
				// broker count BrokersCntMax max, broker No bigger than BrokersCntMax will not send
				// NET_INFO, self port 16bit (client should be 0)
				if cmdLen == 2 {

					var data []byte
					var count int
					rp := int(cmdData[0])<<8 + int(cmdData[1])
					hr, _, _ := net.SplitHostPort(conn.RemoteAddr().String())
					ar := net.JoinHostPort(hr, strconv.Itoa(rp))
					logger.Debug("Net info command from ", ar)
					newBrokers.Range(func(k, _ interface{}) bool {
						if count > utils.BrokersCntMax {
							return false
						}

						if k == ar {
							return true
						}

						data = append(data, byte(len(k.(string))))
						data = append(data, []byte(k.(string))...)

						count++
						return true
					})
					netBrokers.Range(func(k, _ interface{}) bool {
						if count > utils.BrokersCntMax {
							return false
						}

						data = append(data, byte(len(k.(string))))
						data = append(data, []byte(k.(string))...)

						count++
						return true
					})
					if count <= utils.BrokersCntMax && upperAddress != "" {
						data = append(data, byte(len(upperAddress)))
						data = append(data, []byte(upperAddress)...)
					}

					// all known broker address
					_, err := conn.Write(utils.NewDataFrame(utils.NET_INFO, data))
					if err != nil {
						logger.WithError(err).Error("Send response failed")
					}

				}

			case utils.NET_INFO_UPDATE:
				// broker HANDSHAKE
				// NET_INFO_UPDATE, self port 16bit, address len address string, address len address string...
				// response with router table
				if cmdLen >= 2 {
					remotePort := int(cmdData[0])<<8 + int(cmdData[1])
					addr, _, _ := net.SplitHostPort(conn.RemoteAddr().String())
					peerAddress := net.JoinHostPort(addr, strconv.Itoa(remotePort))

					// ping package
					if peerAddress != upperAddress {
						if _, ok := newBrokers.Load(peerAddress); ok {
							// old broker
							newBrokers.Store(peerAddress, time.Now())
						} else {

							// new broker, no response
							newBrokers.Store(peerAddress, time.Now())
							logger.Info("New broker connected: ", peerAddress)

							newData := []byte{byte(selfPort >> 8), byte(selfPort), byte(len(peerAddress))}
							newData = append(newData, []byte(peerAddress)...)
							// send to other 1 jump brokers
							newBrokers.Range(func(k, _ interface{}) bool {

								if k == peerAddress {
									return true
								}

								bkrConn, err := net.DialTimeout("tcp", k.(string), time.Second)
								if err != nil {
									logger.WithError(err).Warn("Send new broker 1 jump broker error")
									return true
								}
								logger.Debug("Send new broker to ", k.(string))
								_, _ = bkrConn.Write(utils.NewDataFrame(utils.NET_INFO_UPDATE, newData))

								return true
							})
							// send to upper broker
							if upperAddress != "" {
								bkrConn, err := net.DialTimeout("tcp", upperAddress, time.Second)
								if err != nil {
									logger.WithError(err).Warn("Send new broker to upper broker error")
								} else {
									logger.Debug("Send new broker to ", upperAddress)
									_, _ = bkrConn.Write(utils.NewDataFrame(utils.NET_INFO_UPDATE, newData))
								}
							}

						}
					}

					// not a broker ping package
					if cmdLen > 2 {
						routeData := cmdData[2:]
						for i := 0; i < len(routeData); i++ {
							// u > 0 delete; u == 0 update
							u := routeData[i] & 0x80
							l := int(routeData[i] & 0x7f)
							if u > 0 {
								logger.WithField("from", conn.RemoteAddr()).Info("Remove broker: ", string(routeData[i+1:i+1+l]))
								netBrokers.Delete(string(routeData[i+1 : i+1+l]))
							} else {
								logger.WithField("from", conn.RemoteAddr()).Info("New broker: ", string(routeData[i+1:i+1+l]))
								netBrokers.Store(string(routeData[i+1:i+1+l]), time.Now())
							}
							i += l
						}

						// send to other 1 jump brokers
						newBrokers.Range(func(k, _ interface{}) bool {

							if k == peerAddress {
								return true
							}

							bkrConn, err := net.DialTimeout("tcp", k.(string), time.Second)
							if err != nil {
								logger.WithError(err).Warn("Send broker update to 1 jump broker error")
								return true
							}
							logger.Debug("Send new broker to ", k.(string))
							_, _ = bkrConn.Write(utils.NewDataFrame(utils.NET_INFO_UPDATE, append([]byte{byte(selfPort >> 8), byte(selfPort)}, routeData...)))

							return true
						})
						// send to upper broker
						if upperAddress != "" && peerAddress != upperAddress {
							bkrConn, err := net.DialTimeout("tcp", upperAddress, time.Second)
							if err != nil {
								logger.WithError(err).Warn("Send broker update to upper broker error")
							} else {
								logger.Debug("Send new broker to ", upperAddress)
								_, _ = bkrConn.Write(utils.NewDataFrame(utils.NET_INFO_UPDATE, append([]byte{byte(selfPort >> 8), byte(selfPort)}, routeData...)))
							}
						}
					}
				}

			case utils.VERSION:
				// tunnel VERSION
				version := []byte{utils.TunnelVersion}
				version = append(version, []byte(utils.Version)...)
				_, err = conn.Write(utils.NewDataFrame(utils.VERSION, version))

			default:
				logger.Warn("RawData data invalid")
			}

			_ = conn.Close()

		}()

	}
}

// start new tcp tunnel
func newTcpTunnel(hostIP string) (int, int, error) {

	serveTcpAddr, err := net.ResolveTCPAddr("tcp", "0.0.0.0:0")

	// quic tunnel between broker and client
	tlsConfig, err := utils.GenerateTLSConfig()
	hostListener, err := quic.ListenAddr(hostIP+":0", tlsConfig, nil)
	if err != nil {
		return 0, 0, err
	}
	serveListener, err := net.ListenTCP("tcp", serveTcpAddr)
	if err != nil {
		_ = hostListener.Close()
		return 0, 0, err
	}

	_, hostPort, _ := net.SplitHostPort(hostListener.Addr().String())
	_, servePort, _ := net.SplitHostPort(serveListener.Addr().String())
	iHostPort, _ := strconv.ParseInt(hostPort, 10, 32)
	iServePort, _ := strconv.ParseInt(servePort, 10, 32)

	logger.Infof("New tcp peer " + hostPort + "-" + servePort)
	peers[int(iHostPort)] = int(iServePort)
	go handleTcpTunnel(int(iHostPort), hostListener, serveListener)

	return int(iHostPort), int(iServePort), nil

}

// start new udp tunnel
func newUdpTunnel(hostIP string, tunnelType byte) (int, int, error) {

	config := utils.TunnelConfig{}
	switch tunnelType {
	case 'q':
		config.Type = utils.ListenQuicListenUdp
	case 't':
		config.Type = utils.ListenTcpListenUdp
	default:
		return 0, 0, errors.New("no such tunnel type " + string(tunnelType))
	}

	tunnel, err := utils.NewTunnel(&config)
	if err != nil {
		return 0, 0, err
	}

	port1, port2 := tunnel.Ports()
	peers[port1] = port2
	logger.Infof("New udp peer " + strconv.Itoa(port1) + "-" + strconv.Itoa(port2))

	go handleUdpTunnel(tunnel)

	return port1, port2, nil

}

func handleTcpTunnel(clientPort int, hostListener quic.Listener, serveListener *net.TCPListener) {

	defer func() {
		delete(peers, clientPort)
	}()
	defer logger.Infof("End tcp peer %d-%d", clientPort, peers[clientPort])

	defer hostListener.Close()
	defer serveListener.Close()

	// client connect tunnel in 10s
	var waitMs int64 = 0
	var qConn quic.Connection
	var err error
	for {
		switch waitMs {
		case 0:
			go func() {
				qConn, err = hostListener.Accept(context.Background())
			}()

		default:
			if qConn == nil && err == nil {
				time.Sleep(time.Millisecond)
			}
		}

		if qConn != nil || err != nil {
			break
		}

		waitMs++
		if waitMs > 1000*10 {
			logger.WithError(err).Error("Get client connection timeout")

			return
		}
	}
	if err != nil {
		logger.WithError(err).Error("Get client connection failed")
		return
	}

	qStream, err := qConn.AcceptStream(context.Background())
	if err != nil {
		logger.WithError(err).Error("Get client stream failed")
		return
	}
	defer qStream.Close()

	conn2, err := serveListener.AcceptTCP()
	if err != nil {
		logger.WithError(err).Error("Get serve connection failed")
		return
	}

	_ = conn2.SetKeepAlive(true)
	defer conn2.Close()

	ch := make(chan int)
	go func() {
		defer func() {
			ch <- 1
		}()

		buf := make([]byte, utils.TransBufSize)

		for {
			n, err := qStream.Read(buf)

			if n > 0 {
				p := 0
				for {
					p, err = conn2.Write(buf[p:n])

					if err != nil || p == n {
						break
					}
				}
			}

			if err != nil {
				break
			}
		}
	}()

	go func() {
		defer func() {
			ch <- 1
		}()

		buf := make([]byte, utils.TransBufSize)

		for {
			n, err := conn2.Read(buf)

			if n > 0 {
				p := 0
				for {
					p, err = qStream.Write(buf[p:n])

					if err != nil || p == n {
						break
					}
				}
			}

			if err != nil {
				break
			}
		}
	}()

	<-ch
}

func handleUdpTunnel(tunnel *utils.Tunnel) {

	port1, port2 := tunnel.Ports()

	defer func() {
		delete(peers, port1)
	}()
	defer logger.Infof("End udp peer %d-%d", port1, port2)
	defer tunnel.Close()

	err := tunnel.Serve(nil, nil, nil, nil)
	if err != nil {
		logger.WithError(err).Error("Tunnel serve error")
	}

}
