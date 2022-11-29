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

	tunnel, err := utils.NewTunnel(&utils.TunnelConfig{
		Type: utils.ListenQuicListenUdp,
	})
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

	err := tunnel.Serve()
	if err != nil {
		logger.WithError(err).Error("Tunnel serve error")
	}

}
