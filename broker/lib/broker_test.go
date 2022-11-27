package broker

import (
	"context"
	"crypto/tls"
	"math/rand"
	"net"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/weilinfox/youmu-thlink/utils"

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

	buf := make([]byte, utils.CmdBufSize+1)

	// test long data
	for i := 0; i < utils.CmdBufSize+1; i++ {
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

		if n != utils.CmdBufSize+1 {
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

	buf := make([]byte, utils.CmdBufSize)

	// test ping
	_, err = conn.Write(utils.NewDataFrame(utils.PING, nil))
	if err != nil {
		t.Fatal("Fail to send ping: ", err.Error())
	}

	conn.SetReadDeadline(time.Now().Add(time.Second))
	n, err := conn.Read(buf)
	conn.SetReadDeadline(time.Time{})
	if err != nil {
		t.Fatal("Cannot read from server: ", err.Error())
	}

	dataStream := utils.NewDataStream()
	dataStream.Append(buf[:n])
	if !dataStream.Parse() || dataStream.Type != utils.PING {
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

	buf := make([]byte, utils.TransBufSize)

	// test udp
	_, err = conn.Write(utils.NewDataFrame(utils.TUNNEL, []byte{'u'}))
	if err != nil {
		t.Fatal("Fail to send new udp tunnel command: ", err.Error())
	}

	//conn.SetReadDeadline(time.Now().Add(time.Second))
	n, err := conn.Read(buf)
	if err != nil {
		t.Fatal("Cannot read from server: ", err.Error())
	}

	dataStream := utils.NewDataStream()
	dataStream.Append(buf[:n])
	if !dataStream.Parse() || dataStream.Type != utils.TUNNEL {
		t.Fatal("Not a new udp tunnel response: ", buf[:n])
	}

	port1 := int(dataStream.RawData[0])<<8 + int(dataStream.RawData[1])
	port2 := int(dataStream.RawData[2])<<8 + int(dataStream.RawData[3])
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

	writeQuicCnt = utils.TransBufSize / 2 * packageCnt
	writeUdpCnt = utils.TransBufSize / 2 * packageCnt

	// write udp
	go func() {

		defer wg.Done()

		buf := make([]byte, utils.TransBufSize/2)
		// compressible random bytes
		tmp := make([]byte, rand.Intn(10)+10)
		for i := 0; i < len(tmp); i++ {
			tmp[i] = byte(rand.Int())
		}
		for i := 0; i < utils.TransBufSize/2; {
			b := rand.Intn(len(tmp))
			for j := 0; j < b && i < utils.TransBufSize/2; {
				buf[i] = tmp[j]
				i++
				j++
			}
		}

		for i := 0; i < packageCnt; i++ {
			n, err := uConn.Write(buf)
			if n != utils.TransBufSize/2 || err != nil {
				t.Fatal("Error write to udp: ", err, " count ", strconv.Itoa(n))
			}
			time.Sleep(time.Millisecond)
		}

		t.Log("UDP send finish")

	}()

	// write quic
	go func() {

		defer wg.Done()

		buf := make([]byte, utils.TransBufSize/2)
		// incompressible random bytes
		for i := 0; i < utils.TransBufSize/2; i++ {
			buf[i] = byte(rand.Int())
		}

		for i := 0; i < packageCnt; i++ {
			n, err := qStream.Write(utils.NewDataFrame(utils.DATA, buf))
			if n-3 != utils.TransBufSize/2 || err != nil {
				//t.Fatal("Error write to quic: ", err, " count ", strconv.Itoa(n))
			}
			time.Sleep(time.Millisecond)
		}

		t.Log("QUIC send finish")

	}()

	// read udp
	go func() {

		defer wg.Done()

		buf := make([]byte, utils.TransBufSize)

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

		dataStream := utils.NewDataStream()
		buf := make([]byte, utils.TransBufSize)

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

			dataStream.Append(buf[:n])
			for dataStream.Parse() {
				if dataStream.Type != utils.DATA {
					t.Error("Not a DATA frame")
				}
				readQuicCnt += dataStream.Length
			}

		}

		t.Log("QUIC resv finish")
		t.Logf("Average compress rate %.3f", dataStream.CompressRateAva())

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
