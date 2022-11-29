package client

import (
	"net"
	"strconv"
	"time"

	"github.com/weilinfox/youmu-thlink/utils"

	"github.com/sirupsen/logrus"
)

var logger = logrus.WithField("client", "internal")
var localPort int

func Main(locPort int, serverHost string, serverPort int) {

	localPort = locPort

	dileHost := serverHost + ":" + strconv.Itoa(serverPort)

	logger.Info("Will connect to local port ", localPort)
	logger.Info("Will connect to broker address ", dileHost)

	serverAddr, err := net.ResolveTCPAddr("tcp", dileHost)
	if err != nil {
		logger.WithError(err).Fatal("Cannot resolve broker address")
	}
	logger.Info("Connected to broker")

	buf := make([]byte, utils.CmdBufSize)

	// calculate delay ms
	var delay int64 = 0
	for i := 5; i >= 0; i-- {
		timeSend := time.Now()

		// send ping
		conn, err := net.DialTCP("tcp", nil, serverAddr)
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
		if !dataStream.Parse() || dataStream.Type() != utils.PING {
			logger.Fatal("Invalid PING response from server")
		}

		delay += timeResp.Sub(timeSend).Milliseconds()
	}
	logger.Infof("Delay %.3fms", float64(delay)/5.0)

	// new udp tunnel
	logger.Info("Ask for new udp tunnel")
	conn, err := net.DialTCP("tcp", nil, serverAddr)

	conn.Write(utils.NewDataFrame(utils.TUNNEL, []byte{'u'}))
	n, _ := conn.Read(buf)
	conn.Close()

	dataStream := utils.NewDataStream()
	dataStream.Append(buf[:n])
	if !dataStream.Parse() || dataStream.Type() != utils.TUNNEL {
		logger.Fatal("Invalid TUNNEL response from server")
	}

	var port1, port2 int
	port1 = int(dataStream.Data()[0])<<8 + int(dataStream.Data()[1])
	port2 = int(dataStream.Data()[2])<<8 + int(dataStream.Data()[3])
	if port1 <= 0 || port1 > 65535 || port2 <= 0 || port2 > 65535 {
		logger.Fatal("Invalid port peer ", port1, port2)
	}

	// New tunnel
	tunnel, err := utils.NewTunnel(&utils.TunnelConfig{
		Type:     utils.DialQuicDialUdp,
		Address0: serverHost + ":" + strconv.Itoa(port1),
		Address1: "localhost:" + strconv.Itoa(localPort),
	})
	if err != nil {
		logger.WithError(err).Fatal("New DialQuicDialUdp error")
	}
	defer tunnel.Close()

	logger.Infof("Tunnel established for remote "+serverAddr.IP.String()+":%d", port2)

	err = tunnel.Serve()
	if err != nil {
		logger.WithError(err).Fatal("Tunnel serve error")
	}

}
