package utils

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/lucas-clemente/quic-go"
	"github.com/sirupsen/logrus"
)

// Tunnel just like a bidirectional pipe
type Tunnel struct {
	tunnelType TunnelType

	configPort0 int
	configPort1 int
	connection0 interface{}
	connection1 interface{}
}

// TunnelType type of tunnel Dial/Listen Address0 and Dial/Listen Address1.
// Warn that Address1 is for communicate between client and broker,
// that means DataStream and DataFrame(see NewDataFrame) will be used.
// Address2 is for generic connection
type TunnelType int

const (
	DialQuicDialUdp TunnelType = iota
	DialTcpDialUdp
	ListenQuicListenUdp
	ListenTcpListenUdp
)

// TunnelConfig default IP is 0.0.0.0:0
type TunnelConfig struct {
	Type     TunnelType
	Address0 string
	Address1 string
}

var loggerTunnel = logrus.WithField("utils", "tunnel")

// NewTunnel set up a new tunnel
func NewTunnel(config *TunnelConfig) (*Tunnel, error) {

	if len(strings.TrimSpace(config.Address0)) == 0 {
		config.Address0 = "0.0.0.0:0"
	}
	if len(strings.TrimSpace(config.Address1)) == 0 {
		config.Address1 = "0.0.0.0:0"
	}

	switch config.Type {
	case ListenQuicListenUdp:

		// listen quic port
		tlsConfig, err := GenerateTLSConfig()
		if err != nil {
			return nil, err
		}
		quicListener, err := quic.ListenAddr(config.Address0, tlsConfig, nil)
		if err != nil {
			return nil, err
		}
		loggerTunnel.Debug("QUIC listen at ", quicListener.Addr().String())

		// listen udp port
		udpAddr, err := net.ResolveUDPAddr("udp", config.Address1)
		if err != nil {
			_ = quicListener.Close()
			return nil, err
		}
		udpConn, err := net.ListenUDP("udp", udpAddr)
		if err != nil {
			_ = quicListener.Close()
			return nil, err
		}
		loggerTunnel.Debug("UDP listen at ", udpConn.LocalAddr().String())

		_, sport0, _ := net.SplitHostPort(quicListener.Addr().String())
		_, sport1, _ := net.SplitHostPort(udpConn.LocalAddr().String())
		port0, _ := strconv.ParseInt(sport0, 10, 32)
		port1, _ := strconv.ParseInt(sport1, 10, 32)

		return &Tunnel{
			tunnelType:  config.Type,
			configPort0: int(port0),
			connection0: quicListener,
			configPort1: int(port1),
			connection1: udpConn,
		}, nil

	case ListenTcpListenUdp:

		// listen tcp port
		tcpAddr, err := net.ResolveTCPAddr("tcp", config.Address0)
		if err != nil {
			return nil, err
		}
		tcpListener, err := net.ListenTCP("tcp", tcpAddr)
		if err != nil {
			return nil, err
		}
		loggerTunnel.Debug("TCP listen at ", tcpListener.Addr().String())

		// listen udp port
		udpAddr, err := net.ResolveUDPAddr("udp", config.Address1)
		if err != nil {
			_ = tcpListener.Close()
			return nil, err
		}
		udpConn, err := net.ListenUDP("udp", udpAddr)
		if err != nil {
			_ = tcpListener.Close()
			return nil, err
		}
		loggerTunnel.Debug("UDP listen at ", udpConn.LocalAddr().String())

		_, sport0, _ := net.SplitHostPort(tcpListener.Addr().String())
		_, sport1, _ := net.SplitHostPort(udpConn.LocalAddr().String())
		port0, _ := strconv.ParseInt(sport0, 10, 32)
		port1, _ := strconv.ParseInt(sport1, 10, 32)

		return &Tunnel{
			tunnelType:  config.Type,
			configPort0: int(port0),
			connection0: tcpListener,
			configPort1: int(port1),
			connection1: udpConn,
		}, nil

	case DialQuicDialUdp:

		// connect quic addr
		tlsConfig := &tls.Config{
			InsecureSkipVerify: true,
			NextProtos:         []string{nextProto},
		}
		quicConn, err := quic.DialAddr(config.Address0, tlsConfig, nil)
		if err != nil {
			return nil, err
		}
		quicStream, err := quicConn.OpenStreamSync(context.Background())
		if err != nil {
			return nil, err
		}
		loggerTunnel.Debug("QUIC dial ", quicConn.RemoteAddr())

		// connect udp addr
		udpAddr, err := net.ResolveUDPAddr("udp", config.Address1)
		if err != nil {
			_ = quicStream.Close()
			return nil, err
		}
		udpConn, err := net.DialUDP("udp", nil, udpAddr)
		if err != nil {
			_ = quicStream.Close()
			return nil, err
		}
		loggerTunnel.Debug("UDP dial ", config.Address1)

		_, sport0, _ := net.SplitHostPort(config.Address0)
		_, sport1, _ := net.SplitHostPort(config.Address1)
		port0, _ := strconv.ParseInt(sport0, 10, 32)
		port1, _ := strconv.ParseInt(sport1, 10, 32)

		return &Tunnel{
			tunnelType:  config.Type,
			configPort0: int(port0),
			connection0: quicStream,
			configPort1: int(port1),
			connection1: udpConn,
		}, nil

	case DialTcpDialUdp:

		// connect tcp addr
		tcpAddr, err := net.ResolveTCPAddr("tcp", config.Address0)
		if err != nil {
			return nil, err
		}
		tcpConn, err := net.DialTCP("tcp", nil, tcpAddr)
		if err != nil {
			return nil, err
		}
		loggerTunnel.Debug("TCP dial ", tcpConn.RemoteAddr())

		// connect udp addr
		udpAddr, err := net.ResolveUDPAddr("udp", config.Address1)
		if err != nil {
			_ = tcpConn.Close()
			return nil, err
		}
		udpConn, err := net.DialUDP("udp", nil, udpAddr)
		if err != nil {
			_ = tcpConn.Close()
			return nil, err
		}
		loggerTunnel.Debug("UDP dial ", config.Address1)

		_, sport0, _ := net.SplitHostPort(config.Address0)
		_, sport1, _ := net.SplitHostPort(config.Address1)
		port0, _ := strconv.ParseInt(sport0, 10, 32)
		port1, _ := strconv.ParseInt(sport1, 10, 32)

		return &Tunnel{
			tunnelType:  config.Type,
			configPort0: int(port0),
			connection0: tcpConn,
			configPort1: int(port1),
			connection1: udpConn,
		}, nil

	}

	return nil, errors.New("no such protocol")
}

// Close make sure all connection be closed after use
func (t *Tunnel) Close() {

	if t.connection0 != nil {
		switch tt := t.connection0.(type) {
		case quic.Listener:
			_ = tt.Close()
		case quic.Stream:
			_ = tt.Close()
		case *net.TCPListener:
			_ = tt.Close()
		case *net.TCPConn:
			_ = tt.Close()
		default:
			loggerTunnel.Errorf("I do not know how to close it: %T", tt)
		}
	}

	if t.connection1 != nil {
		switch tt := t.connection1.(type) {
		case quic.Listener:
			_ = tt.Close()
		case quic.Stream:
			_ = tt.Close()
		case *net.UDPConn:
			_ = tt.Close()
		case *net.TCPConn:
			_ = tt.Close()
		default:
			loggerTunnel.Errorf("I do not know how to close it: %T", tt)
		}
	}

}

// Serve wait for connection and sync data
func (t *Tunnel) Serve() error {

	switch t.tunnelType {
	case ListenQuicListenUdp:

		// accept quic stream
		quicConn, err := t.connection0.(quic.Listener).Accept(context.Background())
		if err != nil {
			return err
		}
		loggerTunnel.Debug("Accept quic connection from ", quicConn.RemoteAddr().String())

		quicStream, err := quicConn.AcceptStream(context.Background())
		if err != nil {
			return err
		}
		loggerTunnel.Debug("Accept quic stream from ", quicConn.RemoteAddr().String())

		defer quicStream.Close()

		t.syncUdp(quicStream, t.connection1.(*net.UDPConn), false, false)

	case ListenTcpListenUdp:

		// accept tcp connection
		err := t.connection0.(*net.TCPListener).SetDeadline(time.Now().Add(time.Second * 10))
		if err != nil {
			return err
		}
		tcpConn, err := t.connection0.(*net.TCPListener).AcceptTCP()
		if err != nil {
			return err
		}
		loggerTunnel.Debug("Accept tcp connection from ", tcpConn.RemoteAddr().String())

		defer tcpConn.Close()

		t.syncUdp(tcpConn, t.connection1.(*net.UDPConn), false, false)

	case DialQuicDialUdp:

		t.syncUdp(t.connection0, t.connection1.(*net.UDPConn), true, true)

	case DialTcpDialUdp:

		t.syncUdp(t.connection0, t.connection1.(*net.UDPConn), true, true)

	}

	return nil
}

// Ports return port peer: port0, port1.
func (t *Tunnel) Ports() (int, int) {
	return t.configPort0, t.configPort1
}

// Type return TunnelType.
func (t *Tunnel) Type() TunnelType {
	return t.tunnelType
}

// syncUdp sync data between quic connection and udp connection.
// Support quic.Stream and *net.TCPConn.
// quicPing: send ping package to avoid quic stream timeout or not;
// udpConnected: udp is waiting for connection or dial to address
func (t *Tunnel) syncUdp(conn interface{}, udpConn *net.UDPConn, sendQuicPing bool, udpConnected bool) {

	switch conn.(type) {
	case quic.Stream:
	case *net.TCPConn:
	default:
		loggerTunnel.Errorf("Unsupported connection type: %T", conn)
		return
	}

	const maxUdpRemoteNo byte = 0xFF
	type udpRemote struct {
		UdpAddr *net.UDPAddr
		No      byte
	}

	var ch chan int
	var udpAddr *net.UDPAddr
	var pingTime time.Time
	udpRemotes := make(map[string]udpRemote)

	// PING
	if sendQuicPing {

		go func() {
			defer func() {
				ch <- 1
			}()

			for {
				var err error

				switch stream := conn.(type) {
				case quic.Stream:
					_, err = stream.Write(NewDataFrame(PING, nil))
				case *net.TCPConn:
					_, err = stream.Write(NewDataFrame(PING, nil))
				}

				pingTime = time.Now()
				if err != nil {
					loggerTunnel.Error("Send PING package failed")
					break
				}
				// no longer than 5 seconds
				time.Sleep(time.Second * 2)
			}

		}()

	}

	// QUIC -> UDP
	go func() {
		defer func() {
			ch <- 1
		}()

		dataStream := NewDataStream()
		buf := make([]byte, TransBufSize)
		var cnt, wcnt int
		var err error

		for {

			switch stream := conn.(type) {
			case quic.Stream:
				cnt, err = stream.Read(buf)
			case *net.TCPConn:
				cnt, err = stream.Read(buf)
			}

			if err != nil {
				loggerTunnel.WithError(err).Warn("Read data from QUIC/TCP stream error")
				break
			}

			dataStream.Append(buf[:cnt])
			for dataStream.Parse() {
				switch dataStream.Type() {

				case DATA:

					if udpConnected {
						// logger.Info("UDP write")
						wcnt, err = udpConn.Write(dataStream.Data())
						// logger.Info("UDP write finish")
						if err != nil || wcnt != dataStream.Len() {
							loggerTunnel.WithError(err).WithField("count", dataStream.Len()).WithField("sent", wcnt).
								Warn("Send data to connected udp error or send count not match")
						}
					} else if udpAddr != nil {
						wcnt, err = udpConn.WriteToUDP(dataStream.Data(), udpAddr)
						if err != nil || wcnt != dataStream.Len() {
							loggerTunnel.WithError(err).WithField("count", dataStream.Len()).WithField("sent", wcnt).
								Warn("Send data to connected udp error or send count not match")
						}
					}

				case PING:

					if sendQuicPing {
						loggerTunnel.Debugf("Delay %.2f ms", float64(time.Now().Sub(pingTime).Nanoseconds())/1000000)
					} else {
						// not sending so response it
						var err error

						switch stream := conn.(type) {
						case quic.Stream:
							_, err = stream.Write(NewDataFrame(PING, nil))
						case *net.TCPConn:
							_, err = stream.Write(NewDataFrame(PING, nil))
						}

						if err != nil {
							loggerTunnel.Error("Send PING package failed")
							break
						}
					}

				}
			}

		}

		loggerTunnel.Debugf("Average compress rate %.3f", dataStream.CompressRateAva())

	}()

	// UDP -> QUIC
	go func() {
		defer func() {
			ch <- 1
		}()

		buf := make([]byte, TransBufSize)
		var cnt, wcnt int
		var err error

		for {

			var remoteNo byte
			if udpConnected {
				cnt, err = udpConn.Read(buf)
				if err != nil {
					loggerTunnel.WithError(err).Warn("Read data from connected udp error")
					break
				}
			} else {
				cnt, udpAddr, err = udpConn.ReadFromUDP(buf)
				if err != nil {
					loggerTunnel.WithError(err).Warn("Read data from unconnected udp error")
					break
				}

				addrString := udpAddr.IP.String() + ":" + strconv.Itoa(udpAddr.Port)
				if v, ok := udpRemotes[addrString]; ok {
					remoteNo = v.No
				} else {
					ul := len(udpRemotes)
					if ul > int(maxUdpRemoteNo) {
						// drop package
						continue
					}
					remoteNo = byte(ul)
					udpRemotes[addrString] = udpRemote{
						UdpAddr: udpAddr,
						No:      remoteNo,
					}
					loggerTunnel.Debug("New UDP connection from ", udpAddr.IP.String(), " port ", udpAddr.Port)
				}

			}

			var data []byte
			if udpConnected {
				data = NewDataFrame(DATA, buf[:cnt])
			} else {
				// TODO: sign DATA with remoteNo
				data = NewDataFrame(DATA, buf[:cnt])
			}

			switch stream := conn.(type) {
			case quic.Stream:
				wcnt, err = stream.Write(data)
			case *net.TCPConn:
				wcnt, err = stream.Write(data)
			}

			if err != nil || wcnt != len(data) {
				loggerTunnel.WithError(err).WithField("count", len(data)).WithField("sent", wcnt).
					Warn("Send data to QUIC/TCP stream error or send count not match")
				break
			}

		}
	}()

	<-ch

}
