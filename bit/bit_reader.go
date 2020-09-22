package bit

import (
	"fmt"
	"math"
)

// Reader ...
type Reader struct {
	pos    int
	offs   uint32 // 0-7
	buffer []byte
	EOF    bool
}

// NewReader ...
func NewReader(in []byte) *Reader {
	r := &Reader{
		pos:  0,
		offs: 0,
		EOF:  false,
	}
	r.buffer = append(r.buffer, in...)

	return r
}

// SetPos ...
func (r *Reader) SetPos(n int) {
	if n < len(r.buffer)-1 {
		r.pos = n
		r.offs = 0
	}
}

// GetSlice ...
func (r *Reader) GetSlice(n int) []byte {
	if r.pos+int(n) > len(r.buffer) {
		return nil
	}

	startIndex := r.pos
	endIndex := r.pos + n
	r.pos += n
	r.offs = 0
	return r.buffer[startIndex:endIndex]
}

// GetLeftSliceCopy ...
func (r *Reader) GetLeftSliceCopy() (slice []byte) {
	slice = append(slice, r.buffer[0:r.pos]...)
	return
}

// GetSliceCopy ...
func (r *Reader) GetSliceCopy(n uint) (slice []byte, err error) {
	if r.pos+int(n) > len(r.buffer) {
		return slice, fmt.Errorf("bit.Reader: not enough minerals")
	}

	slice = append(slice, r.buffer[r.pos:r.pos+int(n)]...)
	r.pos += int(n)
	r.offs = 0
	return
}

// ReadInt ...
func (r *Reader) ReadInt(n uint32) int {
	return int(r.Read(n))
}

// ReadUInt ...
func (r *Reader) ReadUInt(n uint32) uint {
	return uint(r.Read(n))
}

// Read ...
func (r *Reader) Read(n uint32) uint64 {
	var d uint32
	var v uint64
	for n > 0 {
		if r.pos >= len(r.buffer) {
			r.EOF = true
			return 0
		}
		if r.offs+n > 8 {
			d = 8 - r.offs
		} else {
			d = n
		}
		v = v << d
		v += (uint64(r.buffer[r.pos] >> (8 - r.offs - d))) & (0xff >> (8 - d))
		r.offs += d
		n -= d

		if r.offs == 8 {
			r.pos++
			r.offs = 0
		}
	}

	return v
}

// Read8 ...
func (r *Reader) Read8() uint {
	return uint(r.Read(8))
}

// ReadGolomb ...
func (r *Reader) ReadGolomb() uint64 {
	var n uint32
	for n = 0; r.Read(1) == 0 && !r.EOF; n++ {
	}
	return uint64(1<<n) + r.Read(n) - 1
}

// ReadSeGolomb ...
func (r *Reader) ReadSeGolomb() int {
	ueVal := int(r.ReadGolomb())
	k := float64(ueVal)

	seVal := int(math.Ceil(k / 2))
	if ueVal%2 == 0 {
		seVal = -seVal
	}

	return seVal
}

// Left ...
func (r *Reader) Left() int {
	return len(r.buffer) - r.pos
}

// Skip ...
func (r *Reader) Skip(n uint32) {
	var d uint32
	for n > 0 {
		if r.pos >= len(r.buffer) {
			r.EOF = true
			return
		}
		if r.offs+n > 8 {
			d = 8 - r.offs
		} else {
			d = n
		}
		r.offs += d
		n -= d

		if r.offs == 8 {
			r.pos++
			r.offs = 0
		}
	}
}
