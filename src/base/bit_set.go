package base

import ()

const ELEMENT_SIZE = 64
const ELEMENT_ALL uint64 = 0xFFFFFFFFFFFFFFFF

type BitSet struct {
	bits []uint64
	max  int
}

func NewBitSet(max int) *BitSet {
	Assert(max < 1024, "too large")

	length := int(max / ELEMENT_SIZE)

	if 0 != (max % ELEMENT_SIZE) {
		length += 1
	}

	return &BitSet{
		bits: make([]uint64, length),
		max:  max,
	}
}

// Access bit
func (this *BitSet) At(pos int) bool {
	return this.Test(pos)
}

// Count bits set
func (this *BitSet) Count() int {
	count := 0

	for i := 0; i < this.max; i++ {
		if this.At(i) {
			count++
		}
	}
	return count
}

// Return size
func (this *BitSet) Size() int {
	return this.max
}

// Return bit value
func (this *BitSet) Test(pos int) bool {
	Assert(pos < this.max, "overflow.")
	offset := int(pos / ELEMENT_SIZE)
	move := uint64(pos % ELEMENT_SIZE)
	return 0 != (this.bits[offset] & (1 << move))
}

// Test if any bit is set
func (this *BitSet) Any() bool {
	for _, mask := range this.bits {
		if 0 != mask {
			return true
		}
	}
	return false
}

// Test if no bit is set
func (this *BitSet) None() bool {
	return !this.Any()
}

// Test if all bits are set
func (this *BitSet) All() bool {

	length := this.max / ELEMENT_SIZE

	for i := 0; i < length; i++ {
		if ELEMENT_ALL != this.bits[i] {
			return false
		}
	}

	for i := 0; i < this.max%ELEMENT_SIZE; i++ {
		if 0 == (this.bits[length+1] & (1 << uint64(i))) {
			return false
		}
	}

	return true
}

// Set bits
func (this *BitSet) Set(pos int, val bool) {
	Assert(pos < this.max, "overflow.")
	length := pos / ELEMENT_SIZE
	move := uint64(pos % ELEMENT_SIZE)

	if val {
		this.bits[length] |= 1 << move
	} else {
		this.bits[length] ^= 1 << move
	}
}

// to string
func (this *BitSet) String() string {

	str := make([]rune, this.max)

	for i := 0; i < this.max; i++ {
		if this.At(i) {
			str[i] = '1'
		} else {
			str[i] = '0'
		}
	}

	return string(str)
}
