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

const (
	// TunnelVersion tunnel compatible version
	TunnelVersion byte = 2
)

// Tunnel just like a bidirectional pipe
type Tunnel struct {
	tunnelType TunnelType

	pingDelay time.Duration

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

		err = tcpConn.SetNoDelay(true)
		if err != nil {
			tcpConn.Close()
			return nil, err
		}

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

// PluginCallback read/write DATA, in udp tunnel, first bit is udp multiplex id
type PluginCallback func([]byte) (bool, []byte)

// PluginGoroutine goroutine for plugin
type PluginGoroutine func(interface{}, *net.UDPConn)

// Serve wait for connection and sync data
// readFunc, writeFunc: see syncUdp
func (t *Tunnel) Serve(readFunc, writeFunc PluginCallback, plRoutine PluginGoroutine) error {

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

		t.syncUdp(quicStream, t.connection1.(*net.UDPConn), readFunc, writeFunc, plRoutine, false, false)

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

		t.syncUdp(tcpConn, t.connection1.(*net.UDPConn), readFunc, writeFunc, plRoutine, false, false)

	case DialQuicDialUdp:

		t.syncUdp(t.connection0, t.connection1.(*net.UDPConn), readFunc, writeFunc, plRoutine, true, true)

	case DialTcpDialUdp:

		t.syncUdp(t.connection0, t.connection1.(*net.UDPConn), readFunc, writeFunc, plRoutine, true, true)

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

// PingDelay delay between two tunnel
func (t *Tunnel) PingDelay() time.Duration {
	return t.pingDelay
}

// syncUdp sync data between quic connection and udp connection.
// Support quic.Stream and *net.TCPConn.
// readFunc, writeFunc: PluginCallback of when read and write data into tunnel
// quicPing: send ping package to avoid quic stream timeout or not;
// udpConnected: udp is waiting for connection or dial to address
func (t *Tunnel) syncUdp(conn interface{}, udpConn *net.UDPConn, readFunc, writeFunc PluginCallback, plRoutine PluginGoroutine, sendQuicPing, udpConnected bool) {

	switch conn.(type) {
	case quic.Stream:
	case *net.TCPConn:
	default:
		loggerTunnel.Errorf("Unsupported connection type: %T", conn)
		return
	}

	const maxUdpRemoteNo byte = 0xFF

	udpRemoteID := make(map[string]byte) // remote ip record
	var udpRemotes []*net.UDPAddr
	udpVClients := make(map[byte]chan []byte) // local virtual client

	var pingTime time.Time
	ch := make(chan int, int(maxUdpRemoteNo)*2+2)

	if readFunc == nil {
		readFunc = func(data []byte) (bool, []byte) {
			return false, data
		}
	}
	if writeFunc == nil {
		writeFunc = func(data []byte) (bool, []byte) {
			return false, data
		}
	}

	if plRoutine != nil {
		go plRoutine(conn, udpConn)
	}

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
				time.Sleep(time.Second)
			}

		}()

	}

	// UDP -> QUIC
	udpVirtualClient := func(id byte, msg chan []byte) {

		if udpConnected {

			udpAddr, _ := net.ResolveUDPAddr("udp", udpConn.RemoteAddr().String())
			myUdpConn, err := net.DialUDP("udp", nil, udpAddr)
			if err != nil {
				loggerTunnel.WithError(err).Error("New udp virtual client failed with dial udp address error")
				return
			}
			loggerTunnel.WithField("ID", id).Debug("New udp virtual client")

			go func() {
				defer func() {
					ch <- 1
				}()

				for {
					_, _ = myUdpConn.Write(<-msg)

					/*if err != nil {
						loggerTunnel.WithError(err).Warn("Write data to connected udp error")
					}*/
				}

			}()

			go func() {
				defer func() {
					ch <- 1
				}()
				defer myUdpConn.Close()

				buf := make([]byte, TransBufSize)

				for {
					cnt, err := myUdpConn.Read(buf)
					if err != nil {
						// loggerTunnel.WithError(err).Warn("Read data from connected udp error")
						time.Sleep(time.Millisecond * 100)
						continue
					}

					if cnt != 0 {
						reply, data := writeFunc(append([]byte{id}, buf[:cnt]...))
						if data != nil && len(data) > 0 {
							if reply {
								_, _ = myUdpConn.Write(data[1:])
							} else {
								switch stream := conn.(type) {
								case quic.Stream:
									_, err = stream.Write(NewDataFrame(DATA, data))
								case *net.TCPConn:
									_, err = stream.Write(NewDataFrame(DATA, data))
								}

								if err != nil {
									loggerTunnel.WithError(err).Warn("Write data to tunnel error")
									break
								}
							}
						}
					}
				}

			}()

		}

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

					// first byte of data is 8bit guest id
					if udpConnected {

						reply, data := readFunc(dataStream.Data())
						if data != nil && len(data) > 0 {
							if reply {
								switch stream := conn.(type) {
								case quic.Stream:
									_, err = stream.Write(NewDataFrame(DATA, data))
								case *net.TCPConn:
									_, err = stream.Write(NewDataFrame(DATA, data))
								}
								if err != nil {
									loggerTunnel.Error("Send reply package failed")
									break
								}
							} else {
								if ch, ok := udpVClients[data[0]]; ok {
									ch <- data[1:]
								} else {
									ch = make(chan []byte, 32)
									udpVClients[data[0]] = ch
									udpVirtualClient(data[0], ch)
									ch <- data[1:]
								}
							}
						}

					} else if len(udpRemotes) > int(dataStream.Data()[0]) {

						reply, data := readFunc(dataStream.Data())
						if data != nil && len(data) > 0 {
							if reply {
								switch stream := conn.(type) {
								case quic.Stream:
									_, err = stream.Write(NewDataFrame(DATA, data))
								case *net.TCPConn:
									_, err = stream.Write(NewDataFrame(DATA, data))
								}
								if err != nil {
									loggerTunnel.Error("Send reply package failed")
									break
								}
							} else {
								wcnt, err = udpConn.WriteToUDP(data[1:], udpRemotes[data[0]])
								if err != nil || wcnt != len(data)-1 {
									loggerTunnel.WithError(err).WithField("count", len(data)-1).WithField("sent", wcnt).
										Warn("Send data to connected udp error or send count not match")

									// reconnect
									// localAddr := udpConn.LocalAddr()
									// udpLocalAddr, _ := net.ResolveUDPAddr("udp", localAddr.String())
									// _ = udpConn.Close()
									// udpConn, _ = net.DialUDP("udp", nil, udpLocalAddr)
								}
							}
						}

					}

				case PING:

					if sendQuicPing {
						t.pingDelay = time.Now().Sub(pingTime)
						loggerTunnel.Debugf("Delay %.2f ms", float64(t.pingDelay.Nanoseconds())/1000000)
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
	if !udpConnected {

		go func() {
			defer func() {
				ch <- 1
			}()

			buf := make([]byte, TransBufSize)
			var cnt, wcnt int
			var udpAddr *net.UDPAddr
			var remoteNo byte = 0
			var err error

			for {

				cnt, udpAddr, err = udpConn.ReadFromUDP(buf)
				if err != nil {
					loggerTunnel.WithError(err).Warn("Read data from unconnected udp error")
					break
				}

				addrString := udpAddr.IP.String() + ":" + strconv.Itoa(udpAddr.Port)
				if v, ok := udpRemoteID[addrString]; ok {
					remoteNo = v
				} else {
					ul := len(udpRemotes)
					if ul > int(maxUdpRemoteNo) {
						// drop package
						continue
					}
					remoteNo = byte(ul)

					udpRemoteID[addrString] = remoteNo
					udpRemotes = append(udpRemotes, udpAddr)

					loggerTunnel.Debug("New UDP connection from ", udpAddr.IP.String(), " port ", udpAddr.Port)
				}

				var data []byte
				var reply bool

				// first byte of data is 8bit guest id
				reply, data = writeFunc(append([]byte{remoteNo}, buf[:cnt]...))
				if data != nil && len(data) > 0 {
					if reply {
						_, _ = udpConn.WriteToUDP(data[1:], udpAddr)
					} else {
						data = NewDataFrame(DATA, data)

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
				}

			}

		}()

	}

	<-ch

}
