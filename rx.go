package streamcast

import (
	"fmt"
	"log"
	"net"
	"time"
)

// Desired behavior:
// - De-dupliate received frames
// - If we have received frame n + x before receiving frame n, assume frame n is lost.
type Rx struct {
	conn         *net.UDPConn
	addr         net.Addr
	CurrentFrame uint32
	receiveCache []*Frame
	windowSize   uint32
	timeout      time.Duration
}

func NewBroadcastRx(network string, port int, windowSize int) (err error, r *Rx) {
	addr, err := net.ResolveUDPAddr("udp4", fmt.Sprintf("%s:%d", network, port))
	if err != nil {
		return err, nil
	}
	return NewRx(addr, windowSize)
}

func NewRx(addr *net.UDPAddr, windowSize int) (err error, r *Rx) {
	r = new(Rx)
	r.addr = addr
	r.conn, err = net.ListenUDP("udp4", addr)
	if err != nil {
		return err, nil
	}
	r.windowSize = uint32(windowSize)
	r.receiveCache = make([]*Frame, windowSize)
	r.CurrentFrame = 0
	r.timeout = 500 * time.Millisecond
	return
}

func (r *Rx) Reset() {
	for i := range r.receiveCache {
		r.receiveCache[i] = nil
		r.CurrentFrame = 0
	}
}

func (r *Rx) Read() (metadata []byte, data []byte, err error) {
	for {
		if debug {
			fmt.Printf("Trying frame %d\n", r.CurrentFrame)
		}
		// If we already have this frame, return it
		cacheIndex := r.CurrentFrame % r.windowSize
		if r.receiveCache[cacheIndex] != nil {
			f := r.receiveCache[cacheIndex]
			r.receiveCache[cacheIndex] = nil
			if debug {
				fmt.Printf("Returning from cache %d idx %d \n", r.CurrentFrame, cacheIndex)
			}
			r.CurrentFrame += 1
			return f.Metadata, f.Data, nil
		}

		var b [MAX_FRAME_LENGTH]byte
		var f Frame
		r.conn.SetDeadline(time.Now().Add(r.timeout))
		n, _, err := r.conn.ReadFromUDP(b[:])
		if err != nil {
			return nil, nil, err
		}
		if n >= len(b) {
			return nil, nil, fmt.Errorf("Read overflow")
		}
		if err = f.Read(b[:]); err != nil {
			return nil, nil, err
		}
		if debug {
			fmt.Printf("Rx %d\n", f.FrameId)
		}

		if r.CurrentFrame == 0 {
			r.CurrentFrame = f.FrameId
		}
		// If we've already seen this frame, discard it.
		if f.FrameId < r.CurrentFrame {
			continue
		}

		// If we receive the current frame, return it.
		if f.FrameId == r.CurrentFrame {
			r.CurrentFrame = f.FrameId + 1
			return f.Metadata, f.Data, nil
		}

		// If we receive a future frame, cache it
		if f.FrameId > r.CurrentFrame {
			// If we receive a frame so far in the future, it's outside our window
			// save it, advance our current frame counter to the minimum
			// that would put this new frame in the window
			// and flag that we have lost a frame
			distanceFromNow := f.FrameId - r.CurrentFrame
			if distanceFromNow > r.windowSize {
				log.Printf("Dropped frame")
				// Clear cached frames that we advanced over.
				// NOTE: in this implementation, we choose to discard cached
				// frames that didn't make it in the new window
				newCurrentFrame := f.FrameId - r.windowSize + 1

				// Clear cached frames that we're deleting.
				maxToClear := newCurrentFrame - r.CurrentFrame
				// Don't clear more than 1 window.
				if maxToClear > r.windowSize {
					maxToClear = r.windowSize
				}
				if debug {
					fmt.Printf("New current frame %d from rx %d; distance %d\n", newCurrentFrame, f.FrameId, distanceFromNow)
				}
				for i := r.CurrentFrame; i <= r.CurrentFrame+maxToClear; i++ {
					if debug {
						fmt.Printf("Clearing %d idx %d \n", i, i%r.windowSize)
					}
					r.receiveCache[i%r.windowSize] = nil
				}

				r.CurrentFrame = newCurrentFrame
			}
			if debug {
				fmt.Printf("Setting %d idx %d\n", f.FrameId, f.FrameId%r.windowSize)
			}
			r.receiveCache[f.FrameId%r.windowSize] = &f
		}
	}
}

func (r *Rx) SetTimeout(t time.Duration) {
	r.timeout = t
}

func (r *Rx) Close() {
	r.conn.Close()
}
