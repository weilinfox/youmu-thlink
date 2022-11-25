package broker

import (
	"crypto/sha1"
	"errors"
	"net"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/xtaci/kcp-go/v5"
	"golang.org/x/crypto/pbkdf2"
)

const (
	CmdBufSize = 64   // command frame size
	KcpBufSize = 2048 // kcp frame size
)

var logger = logrus.WithField("broker", "internal")

var peers = make(map[int]int)

func Main(listenAddr string) {

	tcpAddr, _ := net.ResolveTCPAddr("tcp4", listenAddr)

	// start udp command interface
	logger.Info("Start tcp command interface at " + tcpAddr.String())
	listener, err := net.ListenTCP("tcp4", tcpAddr)
	if err != nil {
		logger.WithError(err).Fatal("Adddress listen failed")
	}
	defer listener.Close()

	for {

		buf := make([]byte, CmdBufSize)
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

		if n >= CmdBufSize {
			logger.Warn("Command data too long!")
			conn.Close()
			continue
		}

		// handle commands
		go func() {
			switch buf[0] {
			case 0x01:
				// ping
				_, err := conn.Write([]byte{0x01})

				if err != nil {
					logger.WithError(err).Error("Send response failed")
				}

			case 0x02:
				// new tcp/udp tunnel
				// 0x02 t/u
				var port1, port2 int
				var err error

				switch buf[1] {
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

				_, err = conn.Write([]byte{0x02, byte(port1 >> 8), byte(port1), byte(port2 >> 8), byte(port2)})

				if err != nil {
					logger.WithError(err).Error("Send response failed")
				}

			default:
				logger.Warn("Command data invalid")
			}

			_ = conn.Close()

		}()

	}
}

// start new tcp tunnel
func newTcpTunnel(hostIP string) (int, int, error) {

	serveTcpAddr, err := net.ResolveTCPAddr("tcp", "0.0.0.0:0")

	// kcp tunnel between broker and client
	hostListener, err := kcp.Listen(hostIP + ":0")
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

	serveUdpAddr, err := net.ResolveUDPAddr("udp4", "0.0.0.0:0")

	// kcp tunnel between broker and client
	key := pbkdf2.Key([]byte("myon-0406"), []byte("myon-salt"), 1024, 32, sha1.New)
	block, _ := kcp.NewAESBlockCrypt(key)
	hostListener, err := kcp.ListenWithOptions("0.0.0.0:0", block, 10, 3)
	if err != nil {
		return 0, 0, err
	}
	logger.Info("KCP listen at ", hostListener.Addr().String())
	serveConn, err := net.ListenUDP("udp4", serveUdpAddr)
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

func handleTcpTunnel(clientPort int, hostListener net.Listener, serveListener *net.TCPListener) {

	defer func() {
		delete(peers, clientPort)
	}()
	defer logger.Infof("End tcp peer %d-%d", clientPort, peers[clientPort])

	defer hostListener.Close()
	defer serveListener.Close()

	// client connect tunnel in 10s
	var waitMs int64 = 0
	var conn net.Conn
	var err error
	for {
		switch waitMs {
		case 0:
			go func() {
				conn, err = hostListener.Accept()
			}()
		default:
			if conn == nil && err == nil {
				time.Sleep(time.Millisecond)
			}
		}

		if conn != nil || err != nil {
			break
		}

		waitMs++
		if waitMs > 1000*10 {
			logger.WithError(err).Error("Get client connection failed")

			return
		}
	}

	defer conn.Close()

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

		buf := make([]byte, KcpBufSize)

		for {
			n, err := conn.Read(buf)

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

		buf := make([]byte, KcpBufSize)

		for {
			n, err := conn2.Read(buf)

			if n > 0 {
				p := 0
				for {
					p, err = conn.Write(buf[p:n])

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

func handleUdpTunnel(clientPort int, hostListener *kcp.Listener, serveConn *net.UDPConn) {

	defer func() {
		delete(peers, clientPort)
	}()
	defer logger.Infof("End udp peer %d-%d", clientPort, peers[clientPort])

	defer hostListener.Close()
	defer serveConn.Close()

	// client connect tunnel in 10s
	var waitMs int64 = 0
	var kConn net.Conn
	var err error
	for {
		switch waitMs {
		case 0:
			go func() {
				var n int
				buf := make([]byte, KcpBufSize)
				kConn, err = hostListener.AcceptKCP()
				n, err = kConn.Read(buf)
				if err == nil {
					if n != 1 || buf[0] != 0x01 {
						err = errors.New("invalid kcp connection, close it")
						kConn.Close()
						kConn = nil
					} else {
						logger.Info("Client KCP tunnel connected")
					}
				}
			}()
		default:
			if kConn == nil && err == nil {
				time.Sleep(time.Millisecond)
			}
		}

		if kConn != nil || err != nil {
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

	defer kConn.Close()

	var remoteAddr *net.UDPAddr
	connected := false

	ch := make(chan int)
	go func() {
		defer func() {
			ch <- 1
		}()

		buf := make([]byte, KcpBufSize)

		for {
			// read from kcp
			n, err := kConn.Read(buf)
			// logger.Info("kcp read ", n)
			if err != nil {
				logger.WithError(err).Error("KCP read error")
				break
			}

			if connected && n > 0 {
				// logger.Info("udp write ", n)
				p, err := serveConn.WriteToUDP(buf[:n], remoteAddr)

				if err != nil || p != n {
					logger.WithError(err).Error("UDP write error or write count not match")
					break
				}
			}
		}

	}()

	go func() {
		defer func() {
			ch <- 1
		}()

		var n int
		buf := make([]byte, KcpBufSize)

		for {
			n, remoteAddr, err = serveConn.ReadFromUDP(buf)
			// logger.Info("udp read ", n)
			if err != nil {
				logger.WithError(err).Error("UDP read error")
				break
			}
			if !connected {
				connected = true
				logger.WithField("host", remoteAddr.String()).Info("Remote connected")
			}

			if n > 0 {
				// logger.Info("kcp write ", n)
				p, err := kConn.Write(buf[:n])

				if err != nil || p != n {
					logger.WithError(err).Error("KCP write error or write count not match")
					break
				}
			}
		}

	}()

	<-ch
}
