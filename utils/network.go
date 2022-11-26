package utils

import (
	"bytes"
	"compress/gzip"

	"github.com/sirupsen/logrus"
)

const (
	CmdBufSize   = 64       // command frame size
	TransBufSize = 2048 - 3 // forward frame size
)

var logger = logrus.WithField("utils", "network")

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

	// gzip compress
	buf := bytes.NewBuffer(nil)
	gWriter, err := gzip.NewWriterLevel(buf, gzip.BestSpeed)
	if err != nil {
		logger.WithError(err).Error("New gzip compress session error")
	}
	l, err := gWriter.Write(b)
	if err != nil {
		logger.WithError(err).Error("Data compress error")
	}
	if l != len(b) {
		logger.Errorf("Gzip compress data length not match: expect %d get %d", len(b), l)
	}
	gWriter.Close()
	l = buf.Len()

	return append([]byte{byte(t), byte(l >> 8), byte(l)}, buf.Bytes()...)
}

// DataStream parser to receive and parse data stream
type DataStream struct {
	cache          []byte
	cachedDataLen  int
	cachedDataType int
	RawData        []byte
	Length         int
	Type           DataType

	totalData   float64
	totalDecode float64
}

// DataType type of data frame
type DataType int

const (
	DATA DataType = iota
	PING
	TUNNEL
)

// NewDataStream return a empty data stream parser
func NewDataStream() *DataStream {
	return &DataStream{
		cachedDataType: -1,
		cachedDataLen:  -1,
		RawData:        nil,
	}
}

// Append append new data to data stream
func (c *DataStream) Append(b []byte) {
	if b != nil && len(b) != 0 {
		c.cache = append(c.cache, b...)
	}
}

// Parse when return true, new parsed data frame will sign to RawData, Length and Type
func (c *DataStream) Parse() bool {
	// get protocol header
	if c.cachedDataType < 0 && len(c.cache) >= 3 {
		c.cachedDataType = int(c.cache[0])
		c.cachedDataLen = int(c.cache[1])<<8 + int(c.cache[2])
		c.cache = c.cache[3:]
	}
	// get command body
	if c.cachedDataType >= 0 && len(c.cache) >= c.cachedDataLen {

		if c.cachedDataLen == 0 {
			c.RawData = make([]byte, 0)
			c.Length = 0
		} else {
			// gzip decompress
			result := make([]byte, TransBufSize)
			gReader, err := gzip.NewReader(bytes.NewBuffer(c.cache[:c.cachedDataLen]))
			if err != nil {
				logger.WithError(err).Error("New gzip decompress session error", c.cachedDataLen, c.cachedDataType)
			}
			c.Length, err = gReader.Read(result)
			if err != nil && err.Error() != "EOF" {
				logger.WithError(err).Error("Data decompress error")
			}
			gReader.Close()

			c.RawData = result[:c.Length]
			c.totalData += float64(c.cachedDataLen)
			c.totalDecode += float64(c.Length)
		}

		c.Type = DataType(c.cachedDataType)

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
