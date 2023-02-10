package client

import (
	"testing"
)

func TestZlib(t *testing.T) {
	l, z := zlibDataEncodeConf()
	if l == 0 || z == nil {
		t.Fatal("Zlib compress data error")
	}

	if z[0] != 0x78 || z[1] != 0x9c {
		t.Fatalf("Zlib returned wrong data %d %d %d", l, z[0], z[1])
	}
	/*
		fmt.Println(l)
		for i := 0; i < int(l); i++ {
			fmt.Printf("%x, ", z[i])
		}
	*/
}
