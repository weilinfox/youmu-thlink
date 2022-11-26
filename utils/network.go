package utils

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

	l := len(b)
	return append([]byte{byte(t), byte(l >> 8), byte(l)}, b...)
}

// DataStream parser to receive and parse data stream
type DataStream struct {
	cache          []byte
	cachedDataLen  int
	cachedDataType int
	RawData        []byte
	Length         int
	Type           DataType
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
		c.RawData = c.cache[:c.cachedDataLen]
		c.Length, c.Type = c.cachedDataLen, DataType(c.cachedDataType)

		c.cache = c.cache[c.cachedDataLen:]
		c.cachedDataLen = -1
		c.cachedDataType = -1

		return true
	}

	return false
}
