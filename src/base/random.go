package base

import ()

const DEFAULT_MT_SIZE = 624

type MersenneTwister struct {
	mt     []uint
	idx    uint
	size   uint
	isInit bool
}

func NewCustomMersenneTwister(seed int, size uint) *MersenneTwister {

	rand := &MersenneTwister{mt: make([]uint, size), size: size}

	rand.init(seed)
	return rand
}

func NewMersenneTwister(seed int) *MersenneTwister {

	rand := &MersenneTwister{mt: make([]uint, DEFAULT_MT_SIZE), size: DEFAULT_MT_SIZE}

	rand.init(seed)
	return rand
}

func (this *MersenneTwister) init(seed int) {

	var i, p uint

	this.idx = 0

	this.mt[0] = uint(seed)
	for i = 1; i < this.size; i++ {
		p = 1812433253*(this.mt[i-1]^(this.mt[i-1]>>30)) + i
		this.mt[i] = p & 0xffffffff
	}

	this.isInit = true
}

func (this *MersenneTwister) msRand() uint {

	if !this.isInit {
		return 0
	}

	if 0 == this.idx {
		this.msRenerate()
	}

	y := this.mt[this.idx]

	y = y ^ (y >> 11)
	y = y ^ ((y << 7) & 2636928640)
	y = y ^ ((y << 15) & 4022730752)
	y = y ^ (y >> 18)

	this.idx = (this.idx + 1) % this.size
	return y
}

func (this *MersenneTwister) msRenerate() {

	var i, y uint

	half := this.size / 2

	for i = 0; i < this.size; i++ {
		y = (this.mt[i] & 0x80000000) + (this.mt[(i+1)%this.size] & 0x7fffffff)
		this.mt[i] = this.mt[(i+half)%this.size] ^ (y >> 1)

		if y%2 != 0 {
			this.mt[i] = this.mt[i] ^ 2567483615
		}
	}
}

func (this *MersenneTwister) Rseed(seed int) {
	this.init(seed)
}

func (this *MersenneTwister) Rand() uint {
	return this.msRand()
}

func (this *MersenneTwister) Next(min, max int) int {

	if max <= 0 {
		max = min
	}

	val := this.Rand()
	return int(val%uint(max-min+1)) + min
}
