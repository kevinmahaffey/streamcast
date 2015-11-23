package streamcast

import (
	"encoding/binary"
	"net"
	"testing"
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

// Send packets and verify they are received.
func expect(t *testing.T, duplication int, window int, send []uint32, receive []uint32) {
	err, rx := NewBroadcastRx("127.0.0.1", 8888, window)
	if err != nil {
		t.Error(err)
		return
	}
	defer rx.Close()
	rx.SetTimeout(time.Millisecond * 1)

	err, tx := NewBroadcastTx("127.0.0.1", 8888, duplication)
	if err != nil {
		t.Error(err)
		return
	}
	defer tx.Close()

	for _, tid := range send {
		f := makeFrame(tid)
		err = tx.WriteFrame(&f)
		if err != nil {
			t.Error(err)
			return
		}
	}

	for _, expectedRid := range receive {
		_, data, err := rx.Read()
		// If we expected a timeout, be OK with it.
		if expectedRid == 0 {
			if err == nil {
				t.Errorf("Expected timeout, but was none")
				return
			}
			neterr := err.(net.Error)
			if neterr.Timeout() {
				continue
			}
		}

		if err != nil {
			t.Error(err)
			return
		}
		actualRid := extractDataPayload(data)
		if actualRid != expectedRid {
			t.Errorf("Expected payload %d got %d", expectedRid, actualRid)
			return
		}
	}

}

func TestReceiveDuplicates(t *testing.T) {
	expect(t, 1, 3, []uint32{1, 1, 1, 2, 3}, []uint32{1, 2, 3, 0})
}

func TestReceiveLateArrivers(t *testing.T) {
	// Spurious late-arrivers
	expect(t, 1, 3, []uint32{1, 2, 1}, []uint32{1, 2, 0})
}

func TestReceive(t *testing.T) {
	// Golden case
	expect(t, 1, 3, []uint32{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11}, []uint32{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 0})
}

func TestReceiveOOOWithBigWindow(t *testing.T) {
	// Out of order with big window
	expect(t, 1, 10, []uint32{1, 2, 2, 3, 1, 1, 7, 5, 6, 4}, []uint32{1, 2, 3, 4, 5, 6, 7, 0})
}

func TestReceiveOOOWithSmallWindow(t *testing.T) {
	// Out of order with small window. 4 will be skipped because only 5-7 will be in window
	expect(t, 1, 3, []uint32{1, 2, 7, 5, 6, 4}, []uint32{1, 2, 5, 6, 7, 0})
}

func TestReceiveWithHighDuplication(t *testing.T) {
	expect(t, 20, 3, []uint32{1, 2, 3, 5, 6, 4}, []uint32{1, 2, 3, 4, 5, 6, 0})
}
