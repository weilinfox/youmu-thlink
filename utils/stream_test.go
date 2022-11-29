package utils

import (
	"math/rand"
	"testing"
)

func TestStream(t *testing.T) {
	data := make([]byte, TransBufSize)

	for i := 0; i < 10; i++ {
		// compressible random bytes
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

		t.Log("Test compressible random bytes round ", i)
		checkData(data, t)
	}

	for i := 0; i < 10; i++ {
		// incompressible random bytes
		for i := 0; i < TransBufSize; i++ {
			data[i] = byte(rand.Int())
		}

		t.Log("Test incompressible random bytes round ", i)
		checkData(data, t)
	}
}

func checkData(data []byte, t *testing.T) {
	dataStream := NewDataStream()
	dataStream.Append(NewDataFrame(DATA, data))
	if dataStream.Parse() {
		if dataStream.Len() == TransBufSize {
			for i := 0; i < TransBufSize; i++ {
				if dataStream.Data()[i] != data[i] {
					t.Error("Compressible DataStream parse result not match original data")
					break
				}
			}
		} else {
			t.Error("Compressible DataStream parse length not match")
		}
	} else {
		t.Error("Compressible DataStream parse failed")
	}

	t.Log("DataStream compression rate ", dataStream.CompressRateAva())
}
