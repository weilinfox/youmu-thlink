package client

import (
	"crypto/sha1"
	"net"
	"strconv"
	"time"

	broker "github.com/weilinfox/youmu-thlink/broker/lib"

	"github.com/sirupsen/logrus"
	"github.com/xtaci/kcp-go/v5"
	"golang.org/x/crypto/pbkdf2"
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
	key := pbkdf2.Key([]byte("myon-0406"), []byte("myon-salt"), 1024, 32, sha1.New)
	block, _ := kcp.NewAESBlockCrypt(key)
	kSess, err := kcp.DialWithOptions(serverHost+":"+strconv.Itoa(port1), block, 10, 3)
	if err != nil {
		logger.WithError(err).Fatal("Cannot dial tunnel")
	}
	defer kSess.Close()
	_, err = kSess.Write([]byte{0x01})
	if err != nil {
		logger.WithError(err).Fatal("Cannot connect to tunnel")
	}

	logger.Info("KCP local address: ", kSess.LocalAddr())
	logger.Info("KCP remote address: ", kSess.RemoteAddr())
	logger.Infof("Tunnel established for remote "+serverAddr.IP.String()+":%d", port2)
	handleUdp(kSess)

}

func handleUdp(serverConn *kcp.UDPSession) {

	localHost := "localhost:" + strconv.Itoa(localPort)
	var udpConn *net.UDPConn
	defer func() {
		if udpConn != nil {
			udpConn.Close()
		}
	}()

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

			if udpConn != nil {
				p, err := udpConn.Write(buf[:n])
				if err != nil || p != n {
					logger.WithError(err).Warn("Send data to game error or send count not match")
					udpConn.Close()
					udpConn = nil
					continue
				}
			}
		}
	}()

	go func() {
		defer func() {
			ch <- 1
		}()

		buf := make([]byte, broker.KcpBufSize)

		for {

			var err error
			if udpConn == nil {
				logger.Info("Connect to local game")
				udpAddr, _ := net.ResolveUDPAddr("udp4", localHost)
				udpConn, err = net.DialUDP("udp4", nil, udpAddr)
				if err != nil {
					logger.WithError(err).Warn("Connect to local game error")
				}
			}

			if udpConn != nil {
				n, err := udpConn.Read(buf)
				if err != nil {
					logger.WithError(err).Warn("Read data from local game error")
					udpConn.Close()
					udpConn = nil
					continue
				}

				p, err := serverConn.Write(buf[:n])
				if err != nil || p != n {
					logger.WithError(err).Warn("Send data to KCP tunnel error or send count not match")
					break
				}
			}

		}
	}()

	<-ch

}
