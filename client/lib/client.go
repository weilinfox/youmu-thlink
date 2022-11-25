package client

import (
	"github.com/xtaci/kcp-go/v5"
	"net"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
	broker "github.com/weilinfox/youmu-thlink/broker/lib"
)

var logger = logrus.WithField("client", "internal")
var localPort int

func Main(locPort int, serverHost string, serverPort int) {

	localPort = locPort

	dileHost := serverHost + ":" + strconv.Itoa(serverPort)

	logger.Info("Will connect to local port ", localPort)
	logger.Info("Will connect to broker address ", dileHost)

	serverAddr, _ := net.ResolveTCPAddr("tcp", dileHost)
	conn, err := net.DialTCP("tcp", nil, serverAddr)

	if err != nil {
		logger.WithError(err).Fatal("Cannot connect to broker")
	}
	logger.Info("Connected to broker")

	buf := make([]byte, broker.CmdBufSize)

	// calculate delay ms
	var delay int64 = 0
	for i := 5; i >= 0; i-- {
		timeSend := time.Now()
		conn.Write([]byte{0x01})
		_, _ = conn.Read(buf)
		conn.Close()
		timeResp := time.Now()
		if buf[0] != 0x01 {
			logger.Fatal("Invalid ping response from server")
		}
		delay += timeResp.Sub(timeSend).Milliseconds()
	}
	logger.Infof("Delay %.3fms", float64(delay)/5.0)

	// new udp tunnel
	logger.Info("Ask for new udp tunnel")
	conn, err = net.DialTCP("tcp", nil, serverAddr)
	conn.Write([]byte{0x02, 'u'})
	n, _ := conn.Read(buf)

	if buf[0] != 0x02 || n != 5 {
		logger.Fatal("Invalid response from server")
	}
	conn.Close()

	var port1, port2 int
	port1 = int(buf[1])<<8 + int(buf[2])
	port2 = int(buf[3])<<8 + int(buf[4])
	if port1 <= 0 || port1 > 65535 || port2 <= 0 || port2 > 65535 {
		logger.Fatal("Invalid port peer ", port1, port2)
	}

	// connect to broker
	kConn, err := kcp.Dial(serverHost + ":" + strconv.Itoa(port1))
	if err != nil {
		logger.WithError(err).Fatal("Cannot dial tunnel")
	}
	defer kConn.Close()
	_, err = kConn.Write([]byte{0x01})
	if err != nil {
		logger.WithError(err).Fatal("Cannot connect to tunnel")
	}

	logger.Infof("Tunnel established for remote "+serverAddr.IP.String()+":%d", port2)
	handleUdp(kConn)

}

func handleUdp(serverConn net.Conn) {

	localHost := "localhost:" + strconv.Itoa(localPort)
	udpAddr, _ := net.ResolveUDPAddr("udp4", localHost)
	udpConn, err := net.DialUDP("udp4", nil, udpAddr)
	defer udpConn.Close()

	ch := make(chan int)
	go func() {
		defer func() {
			ch <- 1
		}()

		buf := make([]byte, broker.KcpBufSize)

		for {
			n, err := serverConn.Read(buf)
			if err != nil {
				logger.WithError(err).Warn("Read data from KCP tunnel error")
				break
			}

			p, err := udpConn.Write(buf[:n])
			if err != nil || p != n {
				// logger.WithError(err).Warn("Send data to game error or send count not match")
				continue
			}
		}
	}()

	go func() {
		defer func() {
			ch <- 1
		}()

		if err != nil {
			logger.Error("Cannot connect to local game")
			return
		}

		for {

			buf := make([]byte, broker.KcpBufSize)

			n, err := udpConn.Read(buf)
			if err != nil {
				// logger.WithError(err).Warn("Read data from local game error")
				continue
			}

			p, err := serverConn.Write(buf[:n])
			if err != nil || p != n {
				logger.WithError(err).Warn("Send data to KCP tunnel error or send count not match")
				break
			}

		}
	}()

	<-ch

}
