package broker

import (
	"context"
	"net"
	"strconv"
	"time"

	"github.com/weilinfox/youmu-thlink/utils"

	"github.com/lucas-clemente/quic-go"
	"github.com/sirupsen/logrus"
)

var logger = logrus.WithField("broker", "internal")

var peers = make(map[int]int)

func Main(listenAddr string) {

	tcpAddr, _ := net.ResolveTCPAddr("tcp", listenAddr)

	// start udp command interface
	logger.Info("Start tcp command interface at " + tcpAddr.String())
	listener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		logger.WithError(err).Fatal("Adddress listen failed")
	}
	defer listener.Close()

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
		go func() {
			switch dataStream.Type() {
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

				switch dataStream.Data()[0] {
				case 't':
					logger.WithField("host", conn.RemoteAddr().String()).Info("New tcp tunnel")
					host, _, _ := net.SplitHostPort(conn.RemoteAddr().String())
					port1, port2, err = newTcpTunnel(host)
				case 'u':
					logger.WithField("host", conn.RemoteAddr().String()).Info("New udp tunnel")
					host, _, _ := net.SplitHostPort(conn.RemoteAddr().String())
					port1, port2, err = newUdpTunnel(host)
				default:
					logger.Warn("Invalid tunnel type")
				}

				if err != nil {
					logger.WithError(err).Error("Failed to build new tunnel")
				}

				_, err = conn.Write(utils.NewDataFrame(utils.TUNNEL, []byte{byte(port1 >> 8), byte(port1), byte(port2 >> 8), byte(port2)}))

				if err != nil {
					logger.WithError(err).Error("Send response failed")
				}

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
func newUdpTunnel(hostIP string) (int, int, error) {

	serveUdpAddr, err := net.ResolveUDPAddr("udp", "0.0.0.0:0")

	// quic tunnel between broker and client
	tlsConfig, err := utils.GenerateTLSConfig()
	if err != nil {
		return 0, 0, err
	}
	hostListener, err := quic.ListenAddr("0.0.0.0:0", tlsConfig, nil)
	if err != nil {
		return 0, 0, err
	}
	logger.Info("QUIC listen at ", hostListener.Addr().String())
	serveConn, err := net.ListenUDP("udp", serveUdpAddr)
	if err != nil {
		_ = hostListener.Close()
		return 0, 0, err
	}
	logger.Info("UDP listen at ", serveConn.LocalAddr().String())

	_, hostPort, _ := net.SplitHostPort(hostListener.Addr().String())
	_, servePort, _ := net.SplitHostPort(serveConn.LocalAddr().String())
	iHostPort, _ := strconv.ParseInt(hostPort, 10, 32)
	iServePort, _ := strconv.ParseInt(servePort, 10, 32)

	logger.Infof("New udp peer " + hostPort + "-" + servePort)
	peers[int(iHostPort)] = int(iServePort)
	go handleUdpTunnel(int(iHostPort), hostListener, serveConn)

	return int(iHostPort), int(iServePort), nil

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

func handleUdpTunnel(clientPort int, hostListener quic.Listener, serveConn *net.UDPConn) {

	defer func() {
		delete(peers, clientPort)
	}()
	defer logger.Infof("End udp peer %d-%d", clientPort, peers[clientPort])

	defer hostListener.Close()
	defer serveConn.Close()

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

	var remoteAddr *net.UDPAddr
	connected := false

	ch := make(chan int)

	// QUIC -> UDP
	go func() {
		defer func() {
			ch <- 1
		}()

		dataStream := utils.NewDataStream()
		buf := make([]byte, utils.TransBufSize)

		for {
			// read from quic
			n, err := qStream.Read(buf)
			// logger.Info("quic read ", n)
			if err != nil {
				logger.WithError(err).Warn("QUIC read error")
				break
			}

			dataStream.Append(buf[:n])
			for dataStream.Parse() {

				switch dataStream.Type() {
				case utils.DATA:
					if connected && dataStream.Len() > 0 {
						// logger.Info("udp write ", n)
						p, err := serveConn.WriteToUDP(dataStream.Data(), remoteAddr)

						if err != nil || p != dataStream.Len() {
							logger.WithError(err).Warn("UDP write error or write count not match")
							break
						}
					}

				case utils.PING:
					qStream.Write(utils.NewDataFrame(utils.PING, nil))
				}

			}
		}

	}()

	// UDP -> QUIC
	go func() {
		defer func() {
			ch <- 1
		}()

		var n int
		buf := make([]byte, utils.TransBufSize)

		for {
			n, remoteAddr, err = serveConn.ReadFromUDP(buf)
			// logger.Info("udp read ", n)
			if err != nil {
				logger.WithError(err).Warn("UDP read error")
				break
			}
			if !connected {
				connected = true
				logger.WithField("host", remoteAddr.String()).Info("Remote connected")
			}

			if n > 0 {
				// logger.Info("quic write ", n)
				gData := utils.NewDataFrame(utils.DATA, buf[:n])
				p, err := qStream.Write(gData)

				if err != nil || p != len(gData) {
					logger.WithError(err).WithField("count", len(gData)).WithField("sent", p).
						Warn("QUIC write error or write count not match")
					break
				}
			}
		}

	}()

	<-ch
}
