package main

import (
	"github.com/sirupsen/logrus"
	"net"
	"strconv"
	"time"
)

const (
	serverHost = "inuyasha.love"
	serverPort = 4646
	// 花 17723 则 10800
	localPort = 17723
)

var logger = logrus.WithField("client", "internal")

func main() {

	dileHost := serverHost + ":" + strconv.Itoa(serverPort)
	serverAddr, _ := net.ResolveTCPAddr("tcp4", dileHost)
	conn, err := net.DialTCP("tcp4", nil, serverAddr)

	if err != nil {
		logger.WithError(err).Fatal("Cannot connect to broker")
	}
	logger.Info("Connected to broker")

	buf := make([]byte, 512)

	// calculate delay ms
	var delay int64 = 0
	for i := 5; i >= 0; i-- {
		timeSend := time.Now()
		conn.Write([]byte{0x01})
		n, _ := conn.Read(buf)
		conn.Close()
		timeResp := time.Now()
		if buf[0] != 0x01 || n >= 512 {
			logger.Fatal("Invalid response from server")
		}
		delay += timeResp.Sub(timeSend).Milliseconds()
	}
	logger.Infof("Delay %.3fms", float64(delay)/5.0)

	// new udp tunnel
	conn, err = net.DialTCP("tcp4", nil, serverAddr)
	if err != nil {
		logger.WithError(err).Fatal("Broker connection lost")
	}

	logger.Info("Ask for new udp tunnel")
	conn.Write([]byte{0x02, 'u'})
	n, _ := conn.Read(buf)

	if buf[0] != 0x02 || n >= 512 {
		logger.Fatal("Invalid response from server")
	}
	conn.Close()

	var port1, port2 int
	port1 = int(buf[1])<<8 + int(buf[2])
	port2 = int(buf[3])<<8 + int(buf[4])

	// connect to broker
	serverAddr, _ = net.ResolveTCPAddr("tcp4", net.JoinHostPort(serverHost, strconv.Itoa(port1)))
	conn, err = net.DialTCP("tcp4", nil, serverAddr)
	if err != nil {
		logger.Fatal("Cannot connect to tunnel")
	}
	_ = conn.SetKeepAlive(true)
	defer conn.Close()

	logger.Infof("Tunnel established on remote "+serverAddr.IP.String()+":%d", port2)
	handleUdp(conn)
}

func handleUdp(serverConn *net.TCPConn) {

	var udpConn *net.UDPConn
	ch := make(chan int)
	go func() {
		defer func() {
			ch <- 1
		}()

		buf := make([]byte, 512)

		for {
			n, err := serverConn.Read(buf)
			if err != nil {
				break
			}

			p := 0
			for udpConn != nil && p < n {
				p, err = udpConn.Write(buf[p:n])
				if err != nil {
					logger.WithError(err).Warn("Send data to game failed")
					udpConn = nil
				}
			}
		}
	}()

	go func() {
		defer func() {
			ch <- 1
		}()

		localHost := "localhost:" + strconv.Itoa(localPort)
		udpAddr, _ := net.ResolveUDPAddr("udp4", localHost)

		for {
			for udpConn == nil {
				var err error
				udpConn, err = net.DialUDP("udp4", nil, udpAddr)
				if err != nil {
					time.Sleep(time.Millisecond * 100)
				}
			}

			buf := make([]byte, 512)
			for udpConn != nil {
				n, err := udpConn.Read(buf)
				if err != nil {
					udpConn.Close()
					udpConn = nil
					break
				}

				p := 0
				for p < n {
					p, err = serverConn.Write(buf[p:n])
					if err != nil {
						logger.WithError(err).Warn("Send data to server failed")
						break
					}
				}
			}

		}
	}()

	<-ch

}
