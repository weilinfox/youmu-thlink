package client

import (
	"context"
	"crypto/tls"
	"net"
	"strconv"
	"sync"
	"time"

	broker "github.com/weilinfox/youmu-thlink/broker/lib"
	"github.com/weilinfox/youmu-thlink/utils"

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

	serverAddr, err := net.ResolveTCPAddr("tcp4", dileHost)
	if err != nil {
		logger.WithError(err).Fatal("Cannot resolve broker address")
	}
	logger.Info("Connected to broker")

	buf := make([]byte, broker.CmdBufSize)

	// calculate delay ms
	var delay int64 = 0
	for i := 5; i >= 0; i-- {
		timeSend := time.Now()

		// send ping
		conn, err := net.DialTCP("tcp4", nil, serverAddr)
		if err != nil {
			logger.WithError(err).Fatal("Cannot connect to broker")
		}
		_, err = conn.Write(utils.NewDataFrame(utils.PING, nil))
		if err != nil {
			conn.Close()
			logger.Fatal("Send ping failed")
		}
		n, err := conn.Read(buf)
		if err != nil {
			conn.Close()
			logger.Fatal("Get ping response failed")
		}
		conn.Close()
		timeResp := time.Now()

		// parse response
		dataStream := utils.NewDataStream()
		dataStream.Append(buf[:n])
		if !dataStream.Parse() || dataStream.Type != utils.PING {
			logger.Fatal("Invalid PING response from server")
		}

		delay += timeResp.Sub(timeSend).Milliseconds()
	}
	logger.Infof("Delay %.3fms", float64(delay)/5.0)

	// new udp tunnel
	logger.Info("Ask for new udp tunnel")
	conn, err := net.DialTCP("tcp4", nil, serverAddr)

	conn.Write(utils.NewDataFrame(utils.TUNNEL, []byte{'u'}))
	n, _ := conn.Read(buf)
	conn.Close()

	dataStream := utils.NewDataStream()
	dataStream.Append(buf[:n])
	if !dataStream.Parse() || dataStream.Type != utils.TUNNEL {
		logger.Fatal("Invalid TUNNEL response from server")
	}

	var port1, port2 int
	port1 = int(dataStream.RawData[0])<<8 + int(dataStream.RawData[1])
	port2 = int(dataStream.RawData[2])<<8 + int(dataStream.RawData[3])
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

	var mutex sync.Mutex
	ch := make(chan int)
	// PING
	go func() {
		defer func() {
			ch <- 1
		}()

		for {
			mutex.Lock()
			_, err := serverConn.Write(utils.NewDataFrame(utils.PING, nil))
			mutex.Unlock()
			if err != nil {
				logger.Error("Send PING package failed")
				break
			}

			time.Sleep(time.Second)
		}
	}()

	// QUIC -> UDP
	go func() {
		defer func() {
			ch <- 1
		}()

		dataStream := utils.NewDataStream()
		buf := make([]byte, broker.TransBufSize)

		for {

			// logger.Info("QUIC read")
			n, err := serverConn.Read(buf)
			// logger.Info("QUIC read finish")
			if err != nil {
				logger.WithError(err).Warn("Read data from QUIC stream error")
				break
			}

			dataStream.Append(buf[:n])
			for dataStream.Parse() {
				switch dataStream.Type {

				case utils.DATA:
					if udpConn != nil {
						// logger.Info("UDP write")
						p, err := udpConn.Write(dataStream.RawData)
						// logger.Info("UDP write finish")
						if err != nil || p != dataStream.Length {
							logger.WithError(err).WithField("count", dataStream.Length).WithField("sent", p).
								Warn("Send data to game error or send count not match ")
							udpConn.Close()
							udpConn = nil
							continue
						}
					}

				case utils.PING:
					// logger.Info("Get PING")
				}
			}
		}
	}()

	// UDP -> QUIC
	go func() {
		defer func() {
			ch <- 1
		}()

		buf := make([]byte, broker.TransBufSize)

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
				// logger.Info("UDP read")
				// udpConn.SetReadDeadline(time.Now().Add(time.Millisecond * 100))
				n, err := udpConn.Read(buf)
				// logger.Info("UDP read finish")
				if err != nil {
					logger.WithError(err).Warn("Read data from local game error")
					udpConn.Close()
					udpConn = nil
					continue
				}

				// logger.Info("QUIC write")
				mutex.Lock()
				p, err := serverConn.Write(utils.NewDataFrame(utils.DATA, buf[:n]))
				mutex.Unlock()
				// logger.Info("QUIC write finish")
				if err != nil || p != n+3 {
					logger.WithError(err).WithField("count", n+3).WithField("sent", p).
						Warn("Send data to QUIC stream error or send count not match")
					break
				}
			}

		}
	}()

	<-ch

}
