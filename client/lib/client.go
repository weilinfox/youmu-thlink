package client

import (
	"errors"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/weilinfox/youmu-thlink/utils"

	"github.com/sirupsen/logrus"
)

var logger = logrus.WithField("client", "internal")

const (
	DefaultLocalPort  = 10800
	DefaultServerHost = "thlink.inuyasha.love:4646"
	DefaultTunnelType = "tcp"
)

type Client struct {
	tunnel *utils.Tunnel

	localPort  int
	serverHost string
	tunnelType string
}

// New set up new client
func New(localPort int, serverHost string, tunnelType string) (*Client, error) {

	// check arguments
	if localPort <= 0 || localPort > 65535 {
		return nil, errors.New("Invalid port " + strconv.Itoa(localPort))
	}

	_, port, err := net.SplitHostPort(serverHost)
	if err != nil {
		return nil, errors.New("Invalid hostname " + serverHost)
	}
	port64, err := strconv.ParseInt(port, 10, 32)
	if port64 <= 0 || port64 > 65535 {
		return nil, errors.New("Invalid port " + strconv.FormatInt(port64, 10))
	}

	if strings.ToLower(tunnelType) != "tcp" && strings.ToLower(tunnelType) != "quic" {
		return nil, errors.New("Invalid tunnel type " + tunnelType)
	}

	return &Client{
		localPort:  localPort,
		serverHost: serverHost,
		tunnelType: tunnelType,
	}, nil
}

// NewWithDefault set up default client
func NewWithDefault() *Client {
	return &Client{
		localPort:  DefaultLocalPort,
		serverHost: DefaultServerHost,
		tunnelType: DefaultTunnelType,
	}
}

// Ping get client to broker delay
func (c *Client) Ping() time.Duration {

	buf := make([]byte, utils.CmdBufSize)

	// dial port
	conn, err := net.DialTimeout("tcp", c.serverHost, time.Millisecond*500)
	if err != nil {
		logger.WithError(err).Error("Cannot connect to broker")
		return time.Second
	}

	// send ping
	timeSend := time.Now()
	_ = conn.SetWriteDeadline(time.Now().Add(time.Millisecond * 500))
	_, err = conn.Write(utils.NewDataFrame(utils.PING, nil))
	_ = conn.SetWriteDeadline(time.Time{})
	if err != nil {
		_ = conn.Close()
		logger.WithError(err).Error("Send ping failed")
		return time.Second
	}
	_ = conn.SetReadDeadline(time.Now().Add(time.Millisecond * 500))
	n, err := conn.Read(buf)
	_ = conn.SetReadDeadline(time.Time{})
	if err != nil {
		_ = conn.Close()
		logger.WithError(err).Error("Get ping response failed")
		return time.Second
	}
	_ = conn.Close()
	timeResp := time.Now()

	// parse response
	dataStream := utils.NewDataStream()
	dataStream.Append(buf[:n])
	if !dataStream.Parse() || dataStream.Type() != utils.PING {
		logger.Error("Invalid PING response from server")
		return time.Second
	}

	delay := timeResp.Sub(timeSend)

	return delay

}

// Serve ask new tunnel and serve
func (c *Client) Serve() error {

	logger.Info("Will connect to local port ", c.localPort)
	logger.Info("Will connect to broker address ", c.serverHost)

	host, _, err := net.SplitHostPort(c.serverHost)
	if err != nil {
		return err
	}

	// connect to broker
	serverAddr, err := net.ResolveTCPAddr("tcp", c.serverHost)
	if err != nil {
		return err
	}
	logger.Info("Connected to broker")

	buf := make([]byte, utils.CmdBufSize)

	// new tunnel command
	logger.Info("Ask for new udp tunnel")
	conn, err := net.DialTCP("tcp", nil, serverAddr)
	if err != nil {
		return err
	}

	_, err = conn.Write(utils.NewDataFrame(utils.TUNNEL, []byte{'u', c.tunnelType[0]}))
	if err != nil {
		return err
	}
	defer conn.Close()
	n, err := conn.Read(buf)
	if err != nil {
		return err
	}

	// new tunnel command response
	dataStream := utils.NewDataStream()
	dataStream.Append(buf[:n])
	if !dataStream.Parse() || dataStream.Type() != utils.TUNNEL {
		return errors.New("invalid TUNNEL response from server")
	}

	var port1, port2 int
	port1 = int(dataStream.Data()[0])<<8 + int(dataStream.Data()[1])
	port2 = int(dataStream.Data()[2])<<8 + int(dataStream.Data()[3])
	if port1 <= 0 || port1 > 65535 || port2 <= 0 || port2 > 65535 {
		return errors.New("Invalid port peer " + strconv.Itoa(port1) + "-" + strconv.Itoa(port2))
	}

	// Set up tunnel
	config := utils.TunnelConfig{
		Address0: host + ":" + strconv.Itoa(port1),
		Address1: "localhost:" + strconv.Itoa(c.localPort),
	}
	switch c.tunnelType[0] {
	case 't' | 'T':
		config.Type = utils.DialTcpDialUdp
	case 'q' | 'Q':
		config.Type = utils.DialQuicDialUdp
	}

	c.tunnel, err = utils.NewTunnel(&config)
	if err != nil {
		return err
	}

	logger.Infof("Tunnel established for remote "+serverAddr.IP.String()+":%d", port2)

	return c.tunnel.Serve()
}

// Close stop this tunnel
func (c *Client) Close() {
	if c.tunnel != nil {
		c.tunnel.Close()
	}
}

// TunnelDelay ping delay between client and broker
func (c *Client) TunnelDelay() time.Duration {
	return c.tunnel.PingDelay()
}

// LocalPort get client config local port
func (c *Client) LocalPort() int {
	return c.localPort
}

// TunnelType get client config tunnel type tcp/quic
func (c *Client) TunnelType() string {
	return c.tunnelType
}

// ServerHost get client config server host
func (c *Client) ServerHost() string {
	return c.serverHost
}
