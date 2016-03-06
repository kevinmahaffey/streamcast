package streamcast

import (
	"encoding/binary"
	"time"
)

func makeFrame(i uint32) (f Frame) {
	var b [4]byte
	binary.BigEndian.PutUint32(b[:], i)
	f.Data = b[:]
	f.FrameId = i
	return f
}

func extractDataPayload(data []byte) uint32 {
	return binary.BigEndian.Uint32(data)
}

type Packet struct {
	id    uint32
	delay time.Duration
}

// ID, delay since last
func p(id uint32, delayMicroseconds uint32) (p Packet) {
	p = Packet{id: id, delay: time.Duration(delayMicroseconds) * time.Microsecond}
	return
}
