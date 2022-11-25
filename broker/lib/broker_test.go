package broker

import (
	"github.com/xtaci/kcp-go/v5"
	"net"
	"strconv"
	"sync"
	"testing"
	"time"
)

func TestRun(t *testing.T) {
	t.Log("Run broker")
	go Main("127.0.0.1:4646")
	time.Sleep(time.Second)
}

func TestLongData(t *testing.T) {

	buf := make([]byte, CmdBufSize+1)

	// test long data
	for i := 0; i < CmdBufSize+1; i++ {
		buf[i] = byte(i)
	}

	serveUdpAddr, _ := net.ResolveUDPAddr("udp4", "127.0.0.1:4646")
	for i := 5; i >= 0; i-- {
		conn, err := net.DialUDP("udp4", nil, serveUdpAddr)
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
	serveUdpAddr, _ := net.ResolveUDPAddr("udp4", "127.0.0.1:4646")
	conn, err := net.DialUDP("udp4", nil, serveUdpAddr)
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

func TestUDP(t *testing.T) {
	brokerUdpAddr, _ := net.ResolveUDPAddr("udp4", "127.0.0.1:4646")
	conn, err := net.DialUDP("udp4", nil, brokerUdpAddr)
	if err != nil {
		t.Fatal("Fail to connect to server: ", err.Error())
	}
	defer conn.Close()

	buf := make([]byte, KcpBufSize)

	// test ping
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

	kConn, err := kcp.Dial("localhost:" + strconv.Itoa(port1))
	if err != nil {
		t.Fatal("KCP connection failed")
	}
	defer kConn.Close()
	serveUdpAddr, _ := net.ResolveUDPAddr("udp", "localhost:"+strconv.Itoa(port2))
	conn, err = net.DialUDP("udp", nil, serveUdpAddr)
	if err != nil {
		t.Fatal("UDP connection failed")
	}
	defer conn.Close()

	// connect kcp
	_, err = kConn.Write([]byte{0x01})

	// connect udp
	_, err = conn.Write([]byte{0x01})
	if err != nil {
		t.Fatal("UDP write failed: ", err)
	}
	n, err = kConn.Read(buf)
	if err != nil {
		t.Fatal("KCP read failed: ", err)
	}
	if n != 1 || buf[0] != 0x01 {
		t.Fatal("KCP read wrong UDP ping data")
	}

	var writeCnt, readCnt int
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {

		defer wg.Done()

		for i := 0; i < 5; i++ {
			n, err := kConn.Write([]byte{byte(i)})
			if err != nil {
				t.Error("KCP write failed: ", err)
				continue
			}
			writeCnt += n
		}

		t.Log("KCP send pass")

	}()

	go func() {

		defer wg.Done()

		time.Sleep(time.Millisecond * 10)
		for i := 0; i < 5; i++ {
			conn.SetReadDeadline(time.Now().Add(time.Second))
			n, err := conn.Read(buf)
			if n != 1 || err != nil {
				t.Error("Error read from udp: "+err.Error()+" count ", strconv.Itoa(n))
				continue
			}

			readCnt += n
		}

		t.Log("UDP resv pass")

	}()

	wg.Wait()

	if writeCnt != readCnt {
		t.Error("Write read bytes not match")
	} else {
		t.Log("Write read bytes matched")
	}
}

func TestWait(t *testing.T) {
	time.Sleep(time.Second)
}
