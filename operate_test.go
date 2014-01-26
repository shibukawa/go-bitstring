package bitarray

import (
	"bytes"
	"io"
	"reflect"
	"testing"
)

func Setup(m int, n uint8, extras []uint8, unread bool) []*Buffer {
	if extras == nil {
		extras = []uint8{0x00, 0x00, 0x00}
	}
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
		_ = buf.Next(m)
		bufs[i] = NewBuffer(buf)
		bufs[i].n = n
		bufs[i].extra = extras[i]
		bufs[i].unread = unread
	}
	return bufs
}

// PopUint8: Phase 1) Fetch first 3 bits from elements in `ins`
func TestPopUint8Phase1(t *testing.T) {
	size := uint8(3)
	ins := Setup(0, 0, nil, true)

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
}

// PopUint8: Phase 2) Fetch next 2 bits from elements in `ins`
func TestPopUint8Phase2(t *testing.T) {
	extras := []uint8{
		0x00, // 00000|---
		0x80, // 10000|---
		0x50, // 01010|---
	}
	ins := Setup(0, 3, extras, false)
	size := uint8(2)

	uint8Wants := []uint8{
		0x00, // ------|00
		0x02, // ------|10
		0x01, // ------|01
	}
	nWants := []uint8{5, 5, 5}
	extraWants := []uint8{
		0x00, // 000|-----
		0x00, // 000|-----
		0x40, // 010|-----
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
}

// PopUint8: Phase 3) Fetch next 2 bits from elements in `ins`
func TestPopUint8Phase3(t *testing.T) {
	extras := []uint8{
		0x00, // 000|-----
		0x00, // 000|-----
		0x40, // 010|-----
	}
	ins := Setup(1, 5, extras, false)
	size := uint8(5)
	uint8Wants := []uint8{
		0x03, // ---|000,11
		0x00, // ---|000,00
		0x09, // ---|010,01
	}
	nWants := []uint8{2, 2, 2}
	extraWants := []uint8{
		0xfc, // 1111,11|--
		0x3c, // 0011,11|--
		0x54, // 0101,01|--
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
}

func TestPopUint8Phase4(t *testing.T) {
	extras := []uint8{
		0xfc, // 1111,11|--
		0x3c, // 0011,11|--
		0x54, // 0101,01|--
	}
	ins := Setup(2, 2, extras, false)
	size := uint8(7)

	uint8Wants := []uint8{
		0x3f, // --|11,1111
		0x0f, // --|00,1111
		0x15, // --|01,0101
	}
	nWants := []uint8{0, 0, 0}
	extraWants := []uint8{
		0x00,
		0x00,
		0x00,
	}
	uint8Outs := make([]uint8, len(ins))
	for i, c := range ins {
		out, err := c.PopUint8(size)
		if err != nil && err != io.EOF {
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
}
