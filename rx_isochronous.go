package streamcast

import (
	"fmt"
	"log"
	"net"
	"time"
)

type rxTimeout struct{}

func (e *rxTimeout) Error() string   { return "rx timeout" }
func (e *rxTimeout) Timeout() bool   { return true }
func (e *rxTimeout) Temporary() bool { return true }

// Desired behavior:
// 1. On first received frame, wait for max_latency to establish read buffer. Return frame.
// 2. On subsequent received frames, return immediately.
// 3. If "next" frame not received by deadline. Error
// 4.
// - Receive a certain number of frames per second, with a max latency
// - If frames do not arrive by deadline drop and move to next
type RxIsochronous struct {
	conn        *net.UDPConn
	addr        *net.UDPAddr
	cache       *FrameCache
	framePeriod time.Duration
	buffer      time.Duration
	baseTime    time.Time
	baseFrameId uint32
	nextFrameId uint32
}

func NewBroadcastRxIsochronous(network string, port int, framePeriod time.Duration, buffer time.Duration) (r *RxIsochronous, err error) {
	addr, err := net.ResolveUDPAddr("udp4", fmt.Sprintf("%s:%d", network, port))
	if err != nil {
		return nil, err
	}
	return NewRxIsochronous(addr, framePeriod, buffer)
}

func NewRxIsochronous(addr *net.UDPAddr, framePeriod time.Duration, buffer time.Duration) (r *RxIsochronous, err error) {
	r = new(RxIsochronous)
	r.addr = addr
	err = r.Reset()
	if err != nil {
		return nil, err
	}

	r.framePeriod = framePeriod
	r.buffer = buffer

	// Give ourselves little extra buffer, because we'll reconcile exact max latency below.
	windowSize := uint32(buffer/framePeriod) + 2
	r.cache = NewFrameCache(windowSize)
	return
}

func (r *RxIsochronous) Reset() (err error) {
	if r.conn != nil {
		r.conn.Close()
	}
	r.conn, err = net.ListenUDP("udp4", r.addr)
	if err != nil {
		return err
	}
	return
}

func (r *RxIsochronous) NextDeadlineFromNow() time.Time {
	if r.baseTime.IsZero() {
		return time.Time{}
	}

	//Next frame expected time
	nextTime := r.baseTime.Add(time.Duration(r.nextFrameId-r.baseFrameId) * r.framePeriod).Add(r.buffer)
	//How much time till then
	return nextTime
}

func (r *RxIsochronous) underrun() (err error) {
	r.baseFrameId = 0
	r.nextFrameId = 0
	r.baseTime = time.Time{}
	if debug {
		log.Printf("Rx Underrun")
	}
	return new(rxTimeout)
}

func (r *RxIsochronous) Read() (f *Frame, err error) {
	for {
		if debug {
			log.Printf("Trying frame %d\n", r.nextFrameId)
		}

		// Check cache
		f := r.cache.Get(r.nextFrameId)
		if f != nil {
			if debug {
				log.Printf("Found %d in cache", r.nextFrameId)
			}
			r.nextFrameId++
			return f, nil
		}

		// Set deadline
		nextDeadline := r.NextDeadlineFromNow()
		if debug {
			if !nextDeadline.IsZero() {
				log.Printf("Next deadline %d us from now", nextDeadline.Sub(time.Now())/time.Microsecond)
			} else {
				log.Printf("First read")
			}
		}

		if !nextDeadline.IsZero() && nextDeadline.Before(time.Now()) {
			if debug {
				log.Printf("Tried to read after deadline!")
			}
			return nil, r.underrun()
		}

		r.conn.SetDeadline(nextDeadline)

		f = new(Frame)
		var b [MAX_FRAME_LENGTH]byte
		n, _, err := r.conn.ReadFromUDP(b[:])

		// Timeout, report buffer underrun
		neterr, ok := err.(net.Error)
		if ok && neterr.Timeout() {
			return nil, r.underrun()
		}

		// Misc error handling
		if err != nil {
			return nil, err
		}
		if n >= len(b) {
			return nil, fmt.Errorf("Read overflow")
		}

		// Parse frame from received data
		if err = f.Read(b[:]); err != nil {
			return nil, err
		}
		if debug {
			log.Printf("Rx Frame %d\n", f.FrameId)
		}

		// Handle first frame: setup cache and timing
		if r.baseFrameId == 0 {
			r.nextFrameId = f.FrameId
			r.baseFrameId = f.FrameId
			r.cache.FastForwardTo(f.FrameId)
			r.baseTime = time.Now()
		}

		// If we've already seen this frame, discard it.
		if f.FrameId < r.nextFrameId {
			continue
		}

		// If we receive the current frame, return it.
		if f.FrameId == r.nextFrameId {
			r.nextFrameId++
			if debug {
				log.Printf("Returning frame %d", f.FrameId)
			}
			return f, nil
		}

		// If we receive a future frame, cache it
		if f.FrameId > r.nextFrameId {
			r.cache.Put(f)
		}
	}
}

func (r *RxIsochronous) Close() {
	if r.conn != nil {
		r.conn.Close()
	}
}
