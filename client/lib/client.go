package client

import (
	"context"
	"crypto/tls"
	"net"
	"strconv"
	"time"

	broker "github.com/weilinfox/youmu-thlink/broker/lib"

	"github.com/lucas-clemente/quic-go"
	"github.com/sirupsen/logrus"
)

var logger = logrus.WithField("client", "internal")
var localPort int

func Main(locPort int, serverHost string, serverPort int) {

	localPort = locPort

	dileHost := serverHost + ":" + strconv.Itoa(serverPort)

	logger.Info("Will connect to local port ", localPort)
	logger.Info("Will connect to broker address ", dileHost)

	serverAddr, _ := net.ResolveTCPAddr("tcp4", dileHost)
	conn, err := net.DialTCP("tcp4", nil, serverAddr)

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
	conn, err = net.DialTCP("tcp4", nil, serverAddr)
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

	// connect to broker via QUIC
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{"myonTHlink"},
	}
	if err != nil {
		logger.WithError(err).Fatal("Generate TLS Config error ", err)
	}
	qConn, err := quic.DialAddr(serverHost+":"+strconv.Itoa(port1), tlsConfig, nil)
	if err != nil {
		logger.WithError(err).Fatal("QUIC connection failed ", err)
	}
	logger.Info("QUIC local address: ", qConn.LocalAddr())
	logger.Info("QUIC remote address: ", qConn.RemoteAddr())

	qStream, err := qConn.OpenStreamSync(context.Background())
	if err != nil {
		logger.WithError(err).Fatal("QUIC stream open error", err)
	}
	defer qStream.Close()

	logger.Infof("Tunnel established for remote "+serverAddr.IP.String()+":%d", port2)
	handleUdp(qStream)

}

func handleUdp(serverConn quic.Stream) {

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

		buf := make([]byte, broker.TransBufSize)

		for {

			// logger.Info("QUIC read")
			n, err := serverConn.Read(buf)
			// logger.Info("QUIC read finish")
			if err != nil {
				logger.WithError(err).Warn("Read data from QUIC stream error")
				break
			}

			if udpConn != nil {
				// logger.Info("UDP write")
				p, err := udpConn.Write(buf[:n])
				// logger.Info("UDP write finish")
				if err != nil || p != n {
					logger.WithError(err).WithField("count", n).WithField("sent", p).
						Warn("Send data to game error or send count not match ")
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

		buf := make([]byte, broker.TransBufSize)

		for {

			var err error
			if udpConn == nil {
				// logger.Info("Connect to local game")
				udpAddr, _ := net.ResolveUDPAddr("udp4", localHost)
				udpConn, err = net.DialUDP("udp4", nil, udpAddr)
				if err != nil {
					logger.WithError(err).Warn("Connect to local game error")
				}
			}

			if udpConn != nil {
				// logger.Info("UDP read")
				udpConn.SetReadDeadline(time.Now().Add(time.Millisecond * 100))
				n, err := udpConn.Read(buf)
				// logger.Info("UDP read finish")
				if err != nil {
					// logger.WithError(err).Warn("Read data from local game error")
					udpConn.Close()
					udpConn = nil
					continue
				}

				// logger.Info("QUIC write")
				p, err := serverConn.Write(buf[:n])
				// logger.Info("QUIC write finish")
				if err != nil || p != n {
					logger.WithError(err).WithField("count", n).WithField("sent", p).
						Warn("Send data to QUIC stream error or send count not match")
					break
				}
			}

		}
	}()

	<-ch

}
