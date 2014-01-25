package bitarray

import (
	"io"
	"bytes"
	"reflect"
	"testing"
)

func Setup() []*Buffer {
	ins := [][]byte{
		[]byte{
			0x00, // 00000000
			0xff, // 11111111
		},
		[]byte{
			0xf0, // 11110000
			0x0f, // 00001111
		},
		[]byte{
			0xaa, // 10101010
			0x55, // 01010101
		},
	}

	bufs := make([]*Buffer, len(ins))
	for i, bs := range ins {
		buf := bytes.NewBuffer(bs)
		bufs[i] = NewBuffer(buf)
	}
	return bufs
}

func TestPopUint8(t *testing.T) {
	size := uint8(3)
	ins := Setup()

	// Phase 1) Fetch first 3 bits from elements in `ins`
	uint8Wants := []uint8{
		0x00, // -----|000
		0x07, // -----|111
		0x05, // -----|101
	}
	nWants := []uint8{3, 3, 3}
	extraWants := []uint8{
		0x00, // 00000|---
		0x80, // 10000|---
		0x50, // 01010|---
	}
	uint8Outs := make([]uint8, len(ins))
	for i, c := range ins {
		out, err := c.PopUint8(size)
		if err != nil {
			t.Error(err)
		}
		uint8Outs[i] = out
	}
	if !reflect.DeepEqual(uint8Wants, uint8Outs) {
		t.Errorf("wants: %v, outs=%v", uint8Wants, uint8Outs)
	}
	for i, buf := range ins {
		if buf.n != nWants[i] {
			t.Errorf("%dth element: want: %v, out=%v", i, nWants[i], buf.n)
		}
		if buf.extra != extraWants[i] {
			t.Errorf("%dth element: want: %v, out=%v", i, extraWants[i], buf.extra)
		}
	}

	size2 := uint8(2)
	// Phase 2) Fetch next 2 bits from elements in `ins`
	uint8Wants2 := []uint8{
		0x00, // ------|00
		0x02, // ------|10
		0x01, // ------|01
	}
	nWants2 := []uint8{5, 5, 5}
	extraWants2 := []uint8{
		0x00, // 000|-----
		0x00, // 000|-----
		0x40, // 010|-----
	}
	uint8Outs2 := make([]uint8, len(ins))
	for i, c := range ins {
		out, err := c.PopUint8(size2)
		if err != nil {
			t.Error(err)
		}
		uint8Outs2[i] = out
	}
	if !reflect.DeepEqual(uint8Wants2, uint8Outs2) {
		t.Errorf("wants: %v, outs=%v", uint8Wants2, uint8Outs2)
	}
	for i, buf := range ins {
		if buf.n != nWants2[i] {
			t.Errorf("%dth element: want: %v, out=%v", i, nWants2[i], buf.n)
		}
		if buf.extra != extraWants2[i] {
			t.Errorf("%dth element: want: %v, out=%v", i, extraWants2[i], buf.extra)
		}
	}

	size3 := uint8(5)
	// Phase 3) Fetch next 2 bits from elements in `ins`
	uint8Wants3 := []uint8{
		0x03, // ---|000,11
		0x00, // ---|000,00
		0x09, // ---|010,01
	}
	nWants3 := []uint8{2, 2, 2}
	extraWants3 := []uint8{
		0xfc, // 111111|--
		0x3c, // 001111|--
		0x54, // 010101|--
	}
	uint8Outs3 := make([]uint8, len(ins))
	for i, c := range ins {
		out, err := c.PopUint8(size3)
		if err != nil {
			t.Error(err)
		}
		uint8Outs3[i] = out
	}
	if !reflect.DeepEqual(uint8Wants3, uint8Outs3) {
		t.Errorf("wants: %v, outs=%v", uint8Wants3, uint8Outs3)
	}
	for i, buf := range ins {
		if buf.n != nWants3[i] {
			t.Errorf("%dth element: want: %v, out=%v", i, nWants3[i], buf.n)
		}
		if buf.extra != extraWants3[i] {
			t.Errorf("%dth element: want: %v, out=%v", i, extraWants3[i], buf.extra)
		}
	}

	size4 := uint8(7)
	uint8Wants4 := []uint8{
		0x3f, // --|111111
		0x0f, // --|001111
		0x15, // --|010101
	}
	nWants4 := []uint8{0, 0, 0}
	extraWants4 := []uint8{
		0x00,
		0x00,
		0x00,
	}
	uint8Outs4 := make([]uint8, len(ins))
	for i, c := range ins {
		out, err := c.PopUint8(size4)
		if err != nil && err != io.EOF {
			t.Error(err)
		}
		uint8Outs4[i] = out
	}
	if !reflect.DeepEqual(uint8Wants4, uint8Outs4) {
		t.Errorf("wants: %v, outs=%v", uint8Wants4, uint8Outs4)
	}
	for i, buf := range ins {
		if buf.n != nWants4[i] {
			t.Errorf("%dth element: want: %v, out=%v", i, nWants4[i], buf.n)
		}
		if buf.extra != extraWants3[i] {
			t.Errorf("%dth element: want: %v, out=%v", i, extraWants4[i], buf.extra)
		}
	}
}
