package bitarray

import (
	"bytes"
	"io"
	"reflect"
	"testing"
)

// Preparing test cases. Using following byte arreay as test cases.
// 1. |00000000|11111111|00000000|11111111|00000000|11111111|00000000|11111111|
// 2. |11110000|00001111|11110000|00001111|11110000|00001111|11110000|00001111|
// 3. |10101010|01010101|10101010|01010101|10101010|01010101|10101010|01010101|
//
// m indicates bytes already read, n does bit position in current byte,
// extras are left extra bits in each cases above, and unread is flag if
// current operation is fetching data from head of byte array.
func Setup(m int, n uint8, extras []uint8, unread bool) []*Buffer {
	if extras == nil {
		extras = []uint8{0x00, 0x00, 0x00}
	}
	ins := [][]byte{
		[]byte{
			0x00, // 00000000
			0xff, // 11111111
			0x00, // 00000000
			0xff, // 11111111
			0x00, // 00000000
			0xff, // 11111111
			0x00, // 00000000
			0xff, // 11111111
		},
		[]byte{
			0xf0, // 11110000
			0x0f, // 00001111
			0xf0, // 11110000
			0x0f, // 00001111
			0xf0, // 11110000
			0x0f, // 00001111
			0xf0, // 11110000
			0x0f, // 00001111
		},
		[]byte{
			0xaa, // 10101010
			0x55, // 01010101
			0xaa, // 10101010
			0x55, // 01010101
			0xaa, // 10101010
			0x55, // 01010101
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

// Legend
//  [...] : bits tring to fetch
//  <...> : extra bits left in previous operation
//  {...} : extra bits to be left for following operations
//    |   : byte border (smallest packet can be fetched)
//  blank : left part is already read, middle part is same as [...], and right part is
//          left for following operations
//    ,   : hex border
//    -   : padding bits (filled in 0)

// PopUint8: Case 1) Fetch first 3 bits from elements in `ins`.
// Use case that poping bits from head of []byte.
// 1. |[000] {00000}|11111111|00000000|11111111|00000000|11111111|00000000|11111111|
// 2. |[111] {10000}|00001111|11110000|00001111|11110000|00001111|11110000|00001111|
// 3. |[101] {01010}|01010101|10101010|01010101|10101010|01010101|10101010|01010101|
func TestPopUint8Case1(t *testing.T) {
	size := uint8(3)
	ins := Setup(0, 0, nil, true)

	uint8Wants := []uint8{
		0x00, // ----,-|000
		0x07, // ----,-|111
		0x05, // ----,-|101
	}
	nWants := []uint8{3, 3, 3}
	extraWants := []uint8{
		0x00, // 0000,0|---
		0x80, // 1000,0|---
		0x50, // 0101,0|---
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
		t.Errorf("wants: %x, outs=%x", uint8Wants, uint8Outs)
	}
	for i, buf := range ins {
		if buf.n != nWants[i] {
			t.Errorf("%dth element: want: %v, out=%v", i, nWants[i], buf.n)
		}
		if buf.extra != extraWants[i] {
			t.Errorf("%dth element: want: %x, out=%x", i, extraWants[i], buf.extra)
		}
	}
}

// PopUint8: Case 2) Fetch next 2 bits from elements in `ins`.
// Use case that there's extra and specified range is behind next byte border.
// 1. |000 [<00] {000>}|11111111|00000000|11111111|00000000|11111111|00000000|11111111|
// 2. |111 [<10] {000>}|00001111|11110000|00001111|11110000|00001111|11110000|00001111|
// 3. |101 [<01] {010>}|01010101|10101010|01010101|10101010|01010101|10101010|01010101|
func TestPopUint8Case2(t *testing.T) {
	extras := []uint8{
		0x00, // 0000,0---
		0x80, // 1000,0---
		0x50, // 0101,0---
	}
	ins := Setup(0, 3, extras, false)
	size := uint8(2)

	uint8Wants := []uint8{
		0x00, // ----,--00
		0x02, // ----,--10
		0x01, // ----,--01
	}
	nWants := []uint8{5, 5, 5}
	extraWants := []uint8{
		0x00, // 000-,----
		0x00, // 000-,----
		0x40, // 010-,----
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
		t.Errorf("wants: %x, outs=%x", uint8Wants, uint8Outs)
	}
	for i, buf := range ins {
		if buf.n != nWants[i] {
			t.Errorf("%dth element: want: %v, out=%v", i, nWants[i], buf.n)
		}
		if buf.extra != extraWants[i] {
			t.Errorf("%dth element: want: %x, out=%x", i, extraWants[i], buf.extra)
		}
	}
}

// PopUint8: Case 3) Fetch next 5 bits from elements in `ins`.
// Use case that there's extra and specified range crosses byte border.
// 1. |00000 [<000>|11] {111111}|00000000|11111111|00000000|11111111|00000000|11111111|
// 2. |11110 [<000>|00] {001111}|11110000|00001111|11110000|00001111|11110000|00001111|
// 3. |10101 [<010>|01] {010101}|10101010|01010101|10101010|01010101|10101010|01010101|
func TestPopUint8Case3(t *testing.T) {
	extras := []uint8{
		0x00, // 000|-,----
		0x00, // 000|-,----
		0x40, // 010|-,----
	}
	ins := Setup(1, 5, extras, false)
	size := uint8(5)
	uint8Wants := []uint8{
		0x03, // ---0,00|11
		0x00, // ---0,00|00
		0x09, // ---0,10|01
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
		t.Errorf("wants: %x, outs=%x", uint8Wants, uint8Outs)
	}
	for i, buf := range ins {
		if buf.n != nWants[i] {
			t.Errorf("%dth element: want: %v, out=%v", i, nWants[i], buf.n)
		}
		if buf.extra != extraWants[i] {
			t.Errorf("%dth element: want: %x, out=%x", i, extraWants[i], buf.extra)
		}
	}
}

// PopUint8: Case 4) Fetch next 7 bits from elements in `ins`
// Use case that all bytes are read, there's extra and specified range is
// beyond tail of []byte. Returning all left bits and io.EOF.
// 1. |00000000|11111111|00000000|11111111|00000000|11111111|00000000|11 [<111111>|-]
// 2. |11110000|00001111|11110000|00001111|11110000|00001111|11110000|00 [<001111>|-]
// 3. |10101010|01010101|10101010|01010101|10101010|01010101|10101010|01 [<010101>|-]
func TestPopUint8Case4(t *testing.T) {
	extras := []uint8{
		0xfc, // 1111,11|--
		0x3c, // 0011,11|--
		0x54, // 0101,01|--
	}
	ins := Setup(8, 2, extras, false)
	size := uint8(7)

	uint8Wants := []uint8{
		0x3f, // --|11,1111
		0x0f, // --|00,1111
		0x15, // --|01,0101
	}
	nWants := []uint8{1, 1, 1}
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
		t.Errorf("wants: %x, outs=%x", uint8Wants, uint8Outs)
	}
	for i, buf := range ins {
		if buf.n != nWants[i] {
			t.Errorf("%dth element: want: %v, out=%v", i, nWants[i], buf.n)
		}
		if buf.extra != extraWants[i] {
			t.Errorf("%dth element: want: %x, out=%x", i, extraWants[i], buf.extra)
		}
	}
}

// PopUint8: Case 5) Fetch 8 bits from elements in `ins`.
// Use case that there's extra and specified range crosses byte border.
// 1. |00000 [<000>|11111] {111}|00000000|11111111|00000000|11111111|00000000|11111111|
// 2. |11110 [<000>|00001] {111}|11110000|00001111|11110000|00001111|11110000|00001111|
// 3. |10101 [<010>|01010] {101}|10101010|01010101|10101010|01010101|10101010|01010101|
func TestPopUint8Case5(t *testing.T) {
	extras := []uint8{
		0x00, // 000|-,----
		0x00, // 000|-,----
		0x40, // 010|-,----
	}
	ins := Setup(1, 5, extras, false)
	size := uint8(8)
	uint8Wants := []uint8{
		0x1f, // 000|1,1111
		0x01, // 000|0,0001
		0x4a, // 010|0,1010
	}
	nWants := []uint8{5, 5, 5}
	extraWants := []uint8{
		0xe0, // 111|-,----
		0xe0, // 111|-,----
		0xa0, // 101|-,----
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
		t.Errorf("wants: %x, outs=%x", uint8Wants, uint8Outs)
	}
	for i, buf := range ins {
		if buf.n != nWants[i] {
			t.Errorf("%dth element: want: %v, out=%v", i, nWants[i], buf.n)
		}
		if buf.extra != extraWants[i] {
			t.Errorf("%dth element: want: %x, out=%x", i, extraWants[i], buf.extra)
		}
	}
}

// PopUint16: Case 1) Fetch next 3 bits from elements in `ins`.
// Use case that range can be handled within Uint8. Reusing same test case
// as TestPopUint8Case3 and only difference is results are expected as []uint16.
func TestPopUint16Case1(t *testing.T) {
	extras := []uint8{
		0x00, // 000|-,----
		0x00, // 000|-,----
		0x40, // 010|-,----
	}
	ins := Setup(1, 5, extras, false)
	size := uint8(5)
	uint16Wants := []uint16{
		0x03, // ----,----,---0,00|11
		0x00, // ----,----,---0,00|00
		0x09, // ----,----,---0,10|01
	}
	nWants := []uint8{2, 2, 2}
	extraWants := []uint8{
		0xfc, // 1111,11|--
		0x3c, // 0011,11|--
		0x54, // 0101,01|--
	}
	uint16Outs := make([]uint16, len(ins))
	for i, c := range ins {
		out, err := c.PopUint16(size)
		if err != nil {
			t.Error(err)
		}
		uint16Outs[i] = out
	}
	if !reflect.DeepEqual(uint16Wants, uint16Outs) {
		t.Errorf("wants: %v, outs=%v", uint16Wants, uint16Outs)
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

// PopUint16: Case 2) Fetch next 11 bits from elements in `ins`.
// Use case that range is across 1 byte border.
// 1. |000 [<00000>|111111] {11}|00000000|11111111|00000000|11111111|00000000|11111111|
// 2. |111 [<10000>|000011] {11}|11110000|00001111|11110000|00001111|11110000|00001111|
// 3. |101 [<01010>|010101] {01}|10101010|01010101|10101010|01010101|10101010|01010101|
func TestPopUint16Case2(t *testing.T) {
	extras := []uint8{
		0x00, // 0000,0---
		0x80, // 1000,0---
		0x50, // 0101,0---
	}
	ins := Setup(1, 3, extras, false)
	size := uint8(11)
	uint16Wants := []uint16{
		0x003f, // ----,-000,0011,1111
		0x0403, // ----,-100,0000,0011
		0x0295, // ----,-010,1001,0101
	}
	nWants := []uint8{6, 6, 6}
	extraWants := []uint8{
		0xc0, // 11--,----
		0xc0, // 11--,----
		0x40, // 01--,----
	}
	uint16Outs := make([]uint16, len(ins))
	for i, c := range ins {
		out, err := c.PopUint16(size)
		if err != nil {
			t.Error(err)
		}
		uint16Outs[i] = out
	}
	if !reflect.DeepEqual(uint16Wants, uint16Outs) {
		t.Errorf("wants: %x, outs=%x", uint16Wants, uint16Outs)
	}
	for i, buf := range ins {
		if buf.n != nWants[i] {
			t.Errorf("%dth element: want: %d, out=%d", i, nWants[i], buf.n)
		}
		if buf.extra != extraWants[i] {
			t.Errorf("%dth element: want: %x, out=%x", i, extraWants[i], buf.extra)
		}
	}
}

// PopUint16: Case 3) Fetch next 15 bits from elements in `ins`.
// Use case that range is across 2 byte border.
// 1. |000 [<00000>|11111111|00] {000000}|11111111|00000000|11111111|00000000|11111111|
// 2. |111 [<10000>|00001111|11] {110000}|00001111|11110000|00001111|11110000|00001111|
// 3. |101 [<01010>|01010101|10] {101010}|01010101|10101010|01010101|10101010|01010101|
func TestPopUint16Case3(t *testing.T) {
	extras := []uint8{
		0x00, // 0000,0|---
		0x80, // 1000,0|---
		0x50, // 0101,0|---
	}
	ins := Setup(1, 3, extras, false)
	size := uint8(15)
	uint16Wants := []uint16{
		0x03fc, // -000,00|11,1111,11|00
		0x403f, // -100,00|00,0011,11|11
		0x2956, // -010,10|01,0101,01|10
	}
	nWants := []uint8{2, 2, 2}
	extraWants := []uint8{
		0x00, // 0000,00--
		0xc0, // 1100,00--
		0xa8, // 1010,10--
	}
	uint16Outs := make([]uint16, len(ins))
	for i, c := range ins {
		out, err := c.PopUint16(size)
		if err != nil {
			t.Error(err)
		}
		uint16Outs[i] = out
	}
	if !reflect.DeepEqual(uint16Wants, uint16Outs) {
		t.Errorf("wants: %x, outs=%x", uint16Wants, uint16Outs)
	}
	for i, buf := range ins {
		if buf.n != nWants[i] {
			t.Errorf("%dth element: want: %d, out=%d", i, nWants[i], buf.n)
		}
		if buf.extra != extraWants[i] {
			t.Errorf("%dth element: want: %x, out=%x", i, extraWants[i], buf.extra)
		}
	}
}

// PopUint16: Case 4) Fetch next 10 bits from elements in `ins`.
// Use case that all bytes are read, there's extra and specified range is
// beyond tail of []byte. Returning all left bits and io.EOF
// 1. |00000000|11111111|00000000|11111111|00000000|11111111|00000000|11 [<111111>|----]
// 2. |11110000|00001111|11110000|00001111|11110000|00001111|11110000|00 [<001111>|----]
// 3. |10101010|01010101|10101010|01010101|10101010|01010101|10101010|01 [<010101>|----]
func TestPopUint16Case4(t *testing.T) {
	extras := []uint8{
		0xfc, // 1111,11|--
		0x3c, // 0011,11|--
		0x54, // 0101,01|--
	}
	ins := Setup(8, 2, extras, false)
	size := uint8(10)
	uint16Wants := []uint16{
		0x003f, // ----,----,--11,1111
		0x000f, // ----,----,--00,1111
		0x0015, // ----,----,--01,0101
	}
	nWants := []uint8{4, 4, 4}
	extraWants := []uint8{
		0x00, // ----,----
		0x00, // ----,----
		0x00, // ----,----
	}
	uint16Outs := make([]uint16, len(ins))
	for i, c := range ins {
		out, err := c.PopUint16(size)
		if err != nil && err != io.EOF {
			t.Error(err)
		}
		uint16Outs[i] = out
	}
	if !reflect.DeepEqual(uint16Wants, uint16Outs) {
		t.Errorf("wants: %x, outs=%x", uint16Wants, uint16Outs)
	}
	for i, buf := range ins {
		if buf.n != nWants[i] {
			t.Errorf("%dth element: want: %v, out=%v", i, nWants[i], buf.n)
		}
		if buf.extra != extraWants[i] {
			t.Errorf("%dth element: want: %x, out=%x", i, extraWants[i], buf.extra)
		}
	}
}

// PopUint16: Case 5) Fetch next 10 bits from elements in `ins`.
// Use case that 1 byte is left, there's extra and specified range is
// beyond tail of []byte. Returning all left bits and io.EOF
// 1. |00000000|11111111|00000000|11111111|00000000|11111111|0000000 [<0>|11111111|-]
// 2. |11110000|00001111|11110000|00001111|11110000|00001111|1111000 [<0>|00001111|-]
// 3. |10101010|01010101|10101010|01010101|10101010|01010101|1010101 [<0>|01010101|-]
func TestPopUint16Case5(t *testing.T) {
	extras := []uint8{
		0x00, // 0|---,----
		0x00, // 0|---,----
		0x00, // 0|---,----
	}
	ins := Setup(7, 7, extras, false)
	size := uint8(10)
	uint16Wants := []uint16{
		0x00ff, // ----,---0,1111,1111
		0x000f, // ----,---0,0000,1111
		0x0055, // ----,---0,0101,0101
	}
	nWants := []uint8{1, 1, 1}
	extraWants := []uint8{
		0x00, // ----,----
		0x00, // ----,----
		0x00, // ----,----
	}
	uint16Outs := make([]uint16, len(ins))
	for i, c := range ins {
		out, err := c.PopUint16(size)
		if err != nil && err != io.EOF {
			t.Error(err)
		}
		uint16Outs[i] = out
	}
	if !reflect.DeepEqual(uint16Wants, uint16Outs) {
		t.Errorf("wants: %x, outs=%x", uint16Wants, uint16Outs)
	}
	for i, buf := range ins {
		if buf.n != nWants[i] {
			t.Errorf("%dth element: want: %v, out=%v", i, nWants[i], buf.n)
		}
		if buf.extra != extraWants[i] {
			t.Errorf("%dth element: want: %x, out=%x", i, extraWants[i], buf.extra)
		}
	}
}
