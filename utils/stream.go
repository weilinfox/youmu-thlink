package utils

import (
	"bytes"
	"compress/lzw"

	"github.com/sirupsen/logrus"
)

const (
	CmdBufSize    = 64       // command frame size
	TransBufSize  = 2048 - 3 // forward frame size
	BrokersCntMax = 40       // max broker count
)

var loggerStream = logrus.WithField("utils", "stream")

// NewDataFrame build data frame, b can be nil
//
//	+------+--------+--------------+
//	| type | length |   raw data   |
//	| 0  7 | 8   23 | 24    < 2047 |
//	+------+--------+--------------+
//
// type is defined in DataType
func NewDataFrame(t DataType, b []byte) []byte {
	if b == nil || len(b) == 0 {
		return []byte{byte(t), 0x00, 0x00}
	}

	// emmm just remove this feature due to performance concern
	/*if t == DATA {
		useLZW := true

		// lzw compression
		result := bytes.NewBuffer(nil)
		lw := lzw.NewWriter(result, lzw.LSB, 8)
		n, err := lw.Write(b)
		lw.Close()
		if n != len(b) || err != nil {
			loggerStream.WithError(err).Error("LZW compression error")
			useLZW = false
		} else if result.Len() >= len(b) {
			useLZW = false
		}

		if useLZW {
			return append([]byte{byte(LZW_DATA), byte(result.Len() >> 8), byte(result.Len())}, result.Bytes()...)
		} else {
			return append([]byte{byte(DATA), byte(len(b) >> 8), byte(len(b))}, b...)
		}
	}*/

	return append([]byte{byte(t), byte(len(b) >> 8), byte(len(b))}, b...)
}

// DataStream parser to receive and parse data stream
type DataStream struct {
	cache          []byte
	cachedDataLen  int
	cachedDataType int
	rawData        []byte
	dataLength     int
	dataType       DataType

	totalData   float64
	totalDecode float64
}

// DataType 4bit type of data frame
type DataType int

const (
	DATA            DataType = iota // DATA pure data
	PING                            // PING ping
	TUNNEL                          // TUNNEL ask for new tunnel
	LZW_DATA                        // LZW_DATA lzw compressed data
	NET_INFO                        // NET_INFO ask for all broker address in this net
	NET_INFO_UPDATE                 // NET_INFO_UPDATE add or delete broker address in net
	BROKER_INFO                     // BROKER_INFO info of this broker
	VERSION                         // VERSION of tunnel
)

// NewDataStream return a empty data stream parser
func NewDataStream() *DataStream {
	return &DataStream{
		cachedDataType: -1,
		cachedDataLen:  -1,
		rawData:        nil,
	}
}

// Append append new data to data stream
func (c *DataStream) Append(b []byte) {
	if b != nil && len(b) != 0 {
		c.cache = append(c.cache, b...)
	}
}

// Parse when return true, new parsed data frame will sign to rawData, dataLength and dataType
func (c *DataStream) Parse() bool {
	// get protocol header
	if c.cachedDataType < 0 && len(c.cache) >= 3 {

		c.cachedDataType = int(c.cache[0])
		c.cachedDataLen = int(c.cache[1])<<8 + int(c.cache[2])
		c.cache = c.cache[3:]

	}

	// get command body
	if c.cachedDataType >= 0 && len(c.cache) >= c.cachedDataLen {

		c.rawData = c.cache[:c.cachedDataLen]
		c.dataLength, c.dataType = c.cachedDataLen, DataType(c.cachedDataType)

		c.totalData += float64(c.cachedDataLen)

		if c.dataType == LZW_DATA {

			// lzw decompress
			result := make([]byte, TransBufSize)
			lr := lzw.NewReader(bytes.NewReader(c.rawData), lzw.LSB, 8)
			n, err := lr.Read(result)
			lr.Close()
			if err != nil {
				loggerStream.WithError(err).Error("LZW decompression error")
			}

			c.rawData = result[:n]
			c.dataLength = n
			c.dataType = DATA

			c.totalDecode += float64(n)

		} else {

			c.totalDecode += float64(c.cachedDataLen)

		}

		c.cache = c.cache[c.cachedDataLen:]
		c.cachedDataLen = -1
		c.cachedDataType = -1

		return true
	}

	return false
}

// CompressRateAva average compress rate (calculated from decompressed data)
func (c *DataStream) CompressRateAva() float64 {

	if c.totalData == 0 {
		return 0
	}

	return c.totalData / c.totalDecode
}

func (c *DataStream) Type() DataType {
	return c.dataType
}

func (c *DataStream) Len() int {
	return c.dataLength
}

func (c *DataStream) Data() []byte {
	return c.rawData
}
