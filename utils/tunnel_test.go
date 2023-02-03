package utils

import (
	"math/rand"
	"net"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
)

func TestQuicTunnel(t *testing.T) {

	logrus.SetLevel(logrus.DebugLevel)

	// tunnel0
	t.Log("Setup quic tunnel 0")
	tunnel0, err := NewTunnel(&TunnelConfig{
		Type:     ListenQuicListenUdp,
		Address0: "0.0.0.0:0",
		Address1: "0.0.0.0:0",
	})
	if err != nil {
		t.Fatal("New quic tunnel 0 error: ", err)
	}
	port00, port01 := tunnel0.Ports()
	defer tunnel0.Close()

	go tunnel0.Serve(nil, nil, nil, nil)

	// udpConn
	t.Log("Setup to quic tunnel udpConn")
	udpAddr, err := net.ResolveUDPAddr("udp", "0.0.0.0:0")
	if err != nil {
		t.Fatal("ResolveUDPAddr error: ", err)
	}
	udpConn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		t.Fatal("ListenUDP error: ", err)
	}
	defer udpConn.Close()
	_, sUdpPort, _ := net.SplitHostPort(udpConn.LocalAddr().String())
	udpPort64, _ := strconv.ParseInt(sUdpPort, 10, 32)

	// tunnel1
	t.Log("Setup quic tunnel 1")
	tunnel1, err := NewTunnel(&TunnelConfig{
		Type:     DialQuicDialUdp,
		Address0: "localhost:" + strconv.Itoa(port00),
		Address1: "localhost:" + strconv.Itoa(int(udpPort64)),
	})
	if err != nil {
		t.Fatal("New quic tunnel 1 error: ", err)
	}
	defer tunnel1.Close()

	go tunnel1.Serve(nil, nil, nil, nil)

	testTunnel(t, udpConn, port01)

}

func TestTcpTunnel(t *testing.T) {

	logrus.SetLevel(logrus.DebugLevel)

	// tunnel0
	t.Log("Setup tcp tunnel 0")
	tunnel0, err := NewTunnel(&TunnelConfig{
		Type:     ListenTcpListenUdp,
		Address0: "0.0.0.0:0",
		Address1: "0.0.0.0:0",
	})
	if err != nil {
		t.Fatal("New tcp tunnel 0 error: ", err)
	}
	port00, port01 := tunnel0.Ports()
	defer tunnel0.Close()

	go tunnel0.Serve(nil, nil, nil, nil)

	// udpConn
	t.Log("Setup to tcp tunnel udpConn")
	udpAddr, err := net.ResolveUDPAddr("udp", "0.0.0.0:0")
	if err != nil {
		t.Fatal("ResolveUDPAddr error: ", err)
	}
	udpConn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		t.Fatal("ListenUDP error: ", err)
	}
	defer udpConn.Close()
	_, sUdpPort, _ := net.SplitHostPort(udpConn.LocalAddr().String())
	udpPort64, _ := strconv.ParseInt(sUdpPort, 10, 32)

	// tunnel1
	t.Log("Setup tcp tunnel 1")
	tunnel1, err := NewTunnel(&TunnelConfig{
		Type:     DialTcpDialUdp,
		Address0: "localhost:" + strconv.Itoa(port00),
		Address1: "localhost:" + strconv.Itoa(int(udpPort64)),
	})
	if err != nil {
		t.Fatal("New tcp tunnel 1 error: ", err)
	}
	defer tunnel1.Close()

	go tunnel1.Serve(nil, nil, nil, nil)

	testTunnel(t, udpConn, port01)

}

// testTunnel goroutine0 <--> udpConn <--> tunnel1 <--> tunnel0 <--> goroutine1
func testTunnel(t *testing.T, udpConn *net.UDPConn, port01 int) {

	// test data
	var wg sync.WaitGroup
	wg.Add(2)

	// compressible random bytes
	data := make([]byte, TransBufSize)
	tmp := make([]byte, rand.Intn(10)+10)
	for i := 0; i < len(tmp); i++ {
		tmp[i] = byte(rand.Int())
	}
	for i := 0; i < TransBufSize; {
		b := rand.Intn(len(tmp))
		for j := 0; j < b && i < TransBufSize; {
			data[i] = tmp[j]
			i++
			j++
		}
	}

	// read data from udpConn
	go func() {
		defer func() {
			wg.Done()
		}()

		buf := make([]byte, TransBufSize)

		for i := 0; i < 5; i++ {
			_ = udpConn.SetReadDeadline(time.Now().Add(time.Millisecond * 500))
			cnt, udpAddr, err := udpConn.ReadFromUDP(buf)
			_ = udpConn.SetReadDeadline(time.Time{})
			if err != nil {
				t.Error("Read from udpConn error", err)
			}
			_ = udpConn.SetWriteDeadline(time.Now().Add(time.Millisecond * 500))
			cnt1, err := udpConn.WriteToUDP(buf[:cnt], udpAddr)
			_ = udpConn.SetWriteDeadline(time.Time{})
			if err != nil {
				t.Error("Write to udpConn error", err)
			}
			if cnt != cnt1 {
				t.Errorf("Write data count not match: %d != %d", cnt, cnt1)
			}
		}

	}()

	// send data to port01
	go func() {
		defer func() {
			wg.Done()
		}()

		time.Sleep(time.Millisecond * 100)
		buf := make([]byte, TransBufSize)

		udpAddr, err := net.ResolveUDPAddr("udp", "localhost:"+strconv.Itoa(port01))
		if err != nil {
			t.Error("Resolve tunnel0 addr error: ", err)
		} else {

			udpConn, err := net.DialUDP("udp", nil, udpAddr)
			if err != nil {
				t.Error("Dial tunnel0 error: ", err)
			} else {
				defer udpConn.Close()

				t.Log("Connect to UDP ", udpConn.RemoteAddr())

				for i := 0; i < 5; i++ {
					_ = udpConn.SetWriteDeadline(time.Now().Add(time.Millisecond * 500))
					cnt, err := udpConn.Write(data[:TransBufSize-1])
					_ = udpConn.SetWriteDeadline(time.Time{})
					if err != nil {
						t.Error("Write to tunnel0 error")
					}
					_ = udpConn.SetReadDeadline(time.Now().Add(time.Millisecond * 500))
					cnt1, err := udpConn.Read(buf)
					_ = udpConn.SetReadDeadline(time.Time{})
					if err != nil {
						t.Error("Read from tunnel0 error")
					}
					if cnt != cnt1 {
						t.Errorf("Write data count not match: %d != %d", cnt, cnt1)
					}

					for i := 0; i < TransBufSize-1; i++ {
						if buf[i] != data[i] {
							t.Error("Transfer data not count")
							break
						}
					}
				}

			}

		}

	}()

	wg.Wait()
}
