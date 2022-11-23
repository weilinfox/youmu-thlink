package main

import (
	"net"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	listenAddr = "0.0.0.0:4646"
)

var logger = logrus.WithField("broker", "internal")

var peers = make(map[int]int)

func main() {

	tcpAddr, _ := net.ResolveTCPAddr("tcp4", listenAddr)

	logger.Info("Start to listen at " + tcpAddr.String())
	listener, err := net.ListenTCP("tcp4", tcpAddr)
	if err != nil {
		logger.WithError(err).Fatal("Adddress listen failed")
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			logger.WithError(err).Error("Connection failed")
			continue
		}

		buf := make([]byte, 512)
		n, err := conn.Read(buf)
		if err != nil {
			logger.WithError(err).Error("Connection read failed")
			_ = conn.Close()
			continue
		}

		if n >= 512 {
			logger.Warn("Command data too long!")
			_ = conn.Close()
			continue
		}

		switch buf[0] {
		case 0x01:
			// ping
			_, err := conn.Write([]byte{0x01})

			if err != nil {
				logger.WithError(err).Error("Send response failed")
			}

		case 0x02:
			// new tunnel
			// 0x02 t/u
			var port1, port2 int
			var err error

			switch buf[1] {
			case 't':
				logger.WithField("host", conn.RemoteAddr().String()).Info("New tcp tunnel")
				port1, port2, err = newTcpTunnel(net.ParseIP(conn.RemoteAddr().String()).String())
			case 'u':
				logger.WithField("host", conn.RemoteAddr().String()).Info("New udp tunnel")
				port1, port2, err = newUdpTunnel(net.ParseIP(conn.RemoteAddr().String()).String())
			default:
				logger.Warn("Invalid tunnel type")
			}

			if err != nil {
				logger.WithError(err).Error("Failed to build new tunnel")
			} else if port1 != 0 && port2 != 0 {
				_, err = conn.Write([]byte{0x02, byte(port1 >> 8), byte(port1), byte(port2 >> 8), byte(port2)})

				if err != nil {
					logger.WithError(err).Error("Send response failed")
				}
			}

		default:
			logger.Warn("Command data invalid")
		}

		_ = conn.Close()

	}
}

func newTcpTunnel(hostIP string) (int, int, error) {

	hostTcpAddr, _ := net.ResolveTCPAddr("tcp4", hostIP+":0")
	serveTcpAddr, err := net.ResolveTCPAddr("tcp4", "0.0.0.0:0")

	hostListener, err := net.ListenTCP("tcp4", hostTcpAddr)
	if err != nil {
		return 0, 0, err
	}
	serveListener, err := net.ListenTCP("tcp4", serveTcpAddr)
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

func newUdpTunnel(hostIP string) (int, int, error) {

	hostTcpAddr, _ := net.ResolveTCPAddr("tcp4", hostIP+":0")
	serveUdpAddr, err := net.ResolveUDPAddr("udp4", "0.0.0.0:0")

	hostListener, err := net.ListenTCP("tcp4", hostTcpAddr)
	if err != nil {
		return 0, 0, err
	}
	serveConn, err := net.ListenUDP("udp4", serveUdpAddr)
	if err != nil {
		_ = hostListener.Close()
		return 0, 0, err
	}

	_, hostPort, _ := net.SplitHostPort(hostListener.Addr().String())
	_, servePort, _ := net.SplitHostPort(serveConn.LocalAddr().String())
	iHostPort, _ := strconv.ParseInt(hostPort, 10, 32)
	iServePort, _ := strconv.ParseInt(servePort, 10, 32)

	logger.Infof("New udp peer " + hostPort + "-" + servePort)
	peers[int(iHostPort)] = int(iServePort)
	go handleUdpTunnel(int(iHostPort), hostListener, serveConn)

	return int(iHostPort), int(iServePort), nil

}

func handleTcpTunnel(clientPort int, hostListener *net.TCPListener, serveListener *net.TCPListener) {

	defer func() {
		delete(peers, clientPort)
	}()
	defer logger.Infof("End tcp peer %d-%d", clientPort, peers[clientPort])

	_ = hostListener.SetDeadline(time.Now().Add(time.Second * 10))

	defer hostListener.Close()
	defer serveListener.Close()

	conn, err := hostListener.AcceptTCP()

	if err != nil {
		logger.WithError(err).Error("Get client connection failed")
		return
	}

	_ = hostListener.SetDeadline(time.Time{})
	_ = conn.SetKeepAlive(true)
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

		buf := make([]byte, 1024)

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

		buf := make([]byte, 1024)

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

func handleUdpTunnel(clientPort int, hostListener *net.TCPListener, serveConn *net.UDPConn) {

	defer func() {
		delete(peers, clientPort)
	}()
	defer logger.Infof("End udp peer %d-%d", clientPort, peers[clientPort])

	_ = hostListener.SetDeadline(time.Now().Add(time.Second * 10))

	defer hostListener.Close()
	defer serveConn.Close()

	conn, err := hostListener.AcceptTCP()

	if err != nil {
		logger.WithError(err).Error("Get client connection failed")
		return
	}

	_ = hostListener.SetDeadline(time.Time{})
	_ = conn.SetKeepAlive(true)
	defer conn.Close()

	var remoteAddr *net.UDPAddr
	connected := false

	ch := make(chan int)
	go func() {
		defer func() {
			ch <- 1
		}()

		buf := make([]byte, 1024)

		for {
			n, err := conn.Read(buf)

			if connected && n > 0 {
				p := 0
				for {
					p, err = serveConn.WriteToUDP(buf[p:n], remoteAddr)

					if err != nil || p >= n {
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

		var n int
		buf := make([]byte, 1024)

		for {
			n, remoteAddr, err = serveConn.ReadFromUDP(buf)
			if !connected {
				connected = true
				logger.WithField("host", remoteAddr.String()).Info("Remote connected")
			}

			if n > 0 {
				p := 0
				for {
					p, err = conn.Write(buf[p:n])

					if err != nil || p >= n {
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
