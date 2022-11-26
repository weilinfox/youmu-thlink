package broker

import (
	"context"
	"crypto/tls"
	"net"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/lucas-clemente/quic-go"
)

const (
	serverHost    = "localhost"
	serverAddress = serverHost + ":4646"
)

func TestRun(t *testing.T) {
	t.Log("Run broker")
	go Main("127.0.0.1:4646")
	//time.Sleep(time.Second)
}

func TestLongData(t *testing.T) {

	buf := make([]byte, CmdBufSize+1)

	// test long data
	for i := 0; i < CmdBufSize+1; i++ {
		buf[i] = byte(i)
	}

	serveTcpAddr, _ := net.ResolveTCPAddr("tcp4", serverAddress)
	for i := 5; i >= 0; i-- {
		conn, err := net.DialTCP("tcp4", nil, serveTcpAddr)
		if err != nil {
			t.Fatal("Fail to connect to server: ", err.Error())
		}

		n, err := conn.Write(buf)
		if err != nil {
			t.Error("Fail to send data: ", err.Error())
		}

		if n != CmdBufSize+1 {
			t.Error("Send data length not matched: ", n)
		}

		conn.Close()
	}
}

func TestPing(t *testing.T) {
	serveTcpAddr, _ := net.ResolveTCPAddr("tcp4", serverAddress)
	conn, err := net.DialTCP("tcp4", nil, serveTcpAddr)
	if err != nil {
		t.Fatal("Fail to connect to server: ", err.Error())
	}
	defer conn.Close()

	buf := make([]byte, CmdBufSize)

	// test ping
	_, err = conn.Write([]byte{0x01})
	if err != nil {
		t.Fatal("Fail to send ping: ", err.Error())
	}

	conn.SetReadDeadline(time.Now().Add(time.Second))
	n, err := conn.Read(buf)
	conn.SetReadDeadline(time.Time{})
	if err != nil {
		t.Fatal("Cannot read from server: ", err.Error())
	}
	if n != 1 {
		t.Fatal("Ping response length not 1")
	}

	if buf[0] != 0x01 || n != 1 {
		t.Error("Not a ping response: ", buf[:n])
	} else {
		t.Log("Ping test passed")
	}
}

const packageCnt = 64

func TestUDP(t *testing.T) {
	brokerTcpAddr, _ := net.ResolveTCPAddr("tcp4", serverAddress)
	conn, err := net.DialTCP("tcp4", nil, brokerTcpAddr)
	if err != nil {
		t.Fatal("Fail to connect to server: ", err.Error())
	}
	defer conn.Close()

	buf := make([]byte, KcpBufSize)

	// test udp
	_, err = conn.Write([]byte{0x02, 'u'})
	if err != nil {
		t.Fatal("Fail to send new udp tunnel command: ", err.Error())
	}

	//conn.SetReadDeadline(time.Now().Add(time.Second))
	n, err := conn.Read(buf)
	if err != nil {
		t.Fatal("Cannot read from server: ", err.Error())
	}

	if buf[0] != 0x02 || n != 5 {
		t.Fatal("Not a new udp tunnel response: ", buf[:n])
	}

	port1 := int(buf[1])<<8 + int(buf[2])
	port2 := int(buf[3])<<8 + int(buf[4])
	if port1 <= 0 || port1 > 65535 || port2 <= 0 || port2 > 65535 {
		t.Fatal("Invalid port peer", port1, port2)
	}

	// QUIC
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{"myonTHlink"},
	}
	if err != nil {
		t.Fatal("Generate TLS Config error ", err)
	}
	qConn, err := quic.DialAddr(serverHost+":"+strconv.Itoa(port1), tlsConfig, nil)
	if err != nil {
		t.Fatal("QUIC connection failed ", err)
	}
	qStream, err := qConn.OpenStreamSync(context.Background())
	if err != nil {
		t.Fatal("QUIC stream open error", err)
	}
	defer qStream.Close()

	// UDP
	serveUdpAddr, _ := net.ResolveUDPAddr("udp", serverHost+":"+strconv.Itoa(port2))
	uConn, err := net.DialUDP("udp", nil, serveUdpAddr)
	if err != nil {
		t.Fatal("UDP connection failed")
	}
	defer uConn.Close()

	var writeQuicCnt, readQuicCnt, writeUdpCnt, readUdpCnt int
	var wg sync.WaitGroup

	wg.Add(4)

	writeQuicCnt = KcpBufSize / 2 * packageCnt
	writeUdpCnt = KcpBufSize / 2 * packageCnt

	// write udp
	go func() {

		defer wg.Done()

		buf := make([]byte, KcpBufSize/2)

		for i := 0; i < packageCnt; i++ {
			n, err := uConn.Write(buf)
			if n != KcpBufSize/2 || err != nil {
				t.Fatal("Error write to udp: ", err, " count ", strconv.Itoa(n))
			}
			time.Sleep(time.Millisecond)
		}

		t.Log("UDP send finish")

	}()

	// write quic
	go func() {

		defer wg.Done()

		buf := make([]byte, KcpBufSize/2)

		for i := 0; i < packageCnt; i++ {
			n, err := qStream.Write(buf)
			if n != KcpBufSize/2 || err != nil {
				t.Fatal("Error write to quic: ", err, " count ", strconv.Itoa(n))
			}
			time.Sleep(time.Millisecond)
		}

		t.Log("QUIC send finish")

	}()

	// read udp
	go func() {

		defer wg.Done()

		buf := make([]byte, KcpBufSize)

		for i := 0; i < packageCnt; i++ {
			if readUdpCnt == writeUdpCnt {
				break
			}
			uConn.SetReadDeadline(time.Now().Add(time.Second))
			n, err := uConn.Read(buf)
			if err != nil {
				t.Log("Error read from udp: ", err, " count ", strconv.Itoa(n))
				if n == 0 {
					break
				}
			}

			readUdpCnt += n
		}

		t.Log("UDP resv finish")

	}()

	// read quic
	go func() {

		defer wg.Done()

		buf := make([]byte, KcpBufSize)

		for i := 0; i < packageCnt; i++ {
			if readQuicCnt == writeQuicCnt {
				break
			}
			n, err := qStream.Read(buf)
			if err != nil {
				t.Log("Error read from quic: ", err, " count ", strconv.Itoa(n))
				if n == 0 {
					break
				}
			}

			readQuicCnt += n
		}

		t.Log("QUIC resv finish")

	}()

	wg.Wait()

	if writeQuicCnt != readQuicCnt {
		t.Errorf("QUIC write read bytes not match write %d read %d", writeQuicCnt, readQuicCnt)
	} else {
		t.Log("QUIC write read bytes matched", writeQuicCnt)
	}
	if writeUdpCnt != readUdpCnt {
		t.Errorf("UDP write read bytes not match write %d read %d", writeUdpCnt, readUdpCnt)
	} else {
		t.Log("UDP write read bytes matched ", writeUdpCnt)
	}

}
