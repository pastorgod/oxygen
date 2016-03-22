package xnet

import ()

var (
	BigEndian    = &Endian{true}
	LittleEndian = &Endian{false}
)

type Endian struct {
	bigMode bool
}

func (this *Endian) read(buf []byte, count int) (ret int64) {

	if this.bigMode {
		for i := 0; i < count; i++ {
			tmp := int64(buf[count-i-1])
			ret |= (tmp << uint(i*8))
		}

	} else {
		for i := 0; i < count; i++ {
			tmp := int64(buf[i])
			ret |= (tmp << uint(i*8))
		}
	}
	return ret
}

func (this *Endian) write(buf []byte, count int, val int64) {

	if this.bigMode {
		for i := 0; i < count; i++ {
			buf[i] = byte((val >> uint((count-i-1)*8)) & 0xFF)
		}
	} else {
		for i := 0; i < count; i++ {
			buf[i] = byte((val >> uint(i*8)) & 0xFF)
		}
	}
}

func WriteBytes(dst, src []byte, pos int) int {
	return copy(dst[pos:], src)
}

//////////////////////////////////////////////////////////////////////////
func (this *Endian) ReadInt8(buf []byte) int8 {
	return int8(this.read(buf, 1))
}

func (this *Endian) ReadUInt8(buf []byte) uint8 {
	return uint8(this.read(buf, 1))
}

func (this *Endian) ReadInt16(buf []byte) int16 {
	return int16(this.read(buf, 2))
}

func (this *Endian) ReadUInt16(buf []byte) uint16 {
	return uint16(this.read(buf, 2))
}

func (this *Endian) ReadInt32(buf []byte) int32 {
	return int32(this.read(buf, 4))
}

func (this *Endian) ReadUInt32(buf []byte) uint32 {
	return uint32(this.read(buf, 4))
}

func (this *Endian) ReadInt64(buf []byte) int64 {
	return this.read(buf, 8)
}

func (this *Endian) ReadUInt64(buf []byte) uint64 {
	return uint64(this.read(buf, 8))
}

////////////////////////////////////////////////////////////////////////
func (this *Endian) WriteInt8(buf []byte, val int8) int {
	this.write(buf, 1, int64(val))
	return 1
}

func (this *Endian) WriteUInt8(buf []byte, val uint8) int {
	this.write(buf, 1, int64(val))
	return 1
}

func (this *Endian) WriteInt16(buf []byte, val int16) int {
	this.write(buf, 2, int64(val))
	return 2
}

func (this *Endian) WriteUInt16(buf []byte, val uint16) int {
	this.write(buf, 2, int64(val))
	return 2
}

func (this *Endian) WriteInt32(buf []byte, val int32) int {
	this.write(buf, 4, int64(val))
	return 4
}

func (this *Endian) WriteUInt32(buf []byte, val uint32) int {
	this.write(buf, 4, int64(val))
	return 4
}

func (this *Endian) WriteInt64(buf []byte, val int64) int {
	this.write(buf, 8, int64(val))
	return 8
}

func (this *Endian) WriteUInt64(buf []byte, val uint64) int {
	this.write(buf, 8, int64(val))
	return 8
}
