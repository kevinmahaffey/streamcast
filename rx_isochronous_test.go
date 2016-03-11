package streamcast

import (
	"net"
	"testing"
	"time"
)

// TIMEOUT magic number
const TO uint32 = 0

// Send packets and verify they are received.
func sendIsoc(t *testing.T, duplication int, send []Packet) {
	go func() {
		tx, err := NewUdpTx("127.0.0.1", 8888, duplication)
		if err != nil {
			t.Error(err)
			return
		}
		defer tx.Close()
		for _, packet := range send {
			time.Sleep(packet.delay)
			f := makeFrame(packet.id)
			err = tx.WriteFrame(&f)
			if err != nil {
				t.Error(err)
				return
			}
		}
	}()
	return
}

func expectIsoc(t *testing.T, duplication int, period time.Duration, maxLatency time.Duration, send []Packet, receive []uint32) {
	rx, err := NewRxIsochronous("udp", "127.0.0.1", 8888, period, maxLatency)
	if err != nil {
		t.Error(err)
		return
	}
	defer rx.Close()

	sendIsoc(t, duplication, send)

	for _, expectedRid := range receive {
		frame, err := rx.Read()
		// If we expected a timeout, be OK with it.
		if expectedRid == TO {
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
		actualRid := extractDataPayload(frame.Data)
		if actualRid != expectedRid {
			t.Errorf("Expected payload %d got %d", expectedRid, actualRid)
			return
		}
	}
}

func TestReceiveDuplicates(t *testing.T) {
	expectIsoc(t,
		1,                // duplication
		time.Millisecond, // period
		time.Millisecond, // max latency
		[]Packet{p(1, 0), p(1, 0), p(1, 0), p(2, 0), p(3, 0)},
		[]uint32{1, 2, 3, TO})
}

func TestReceiveLateArrivers(t *testing.T) {
	// Spurious late-arrivers
	expectIsoc(t,
		1,                // duplication
		time.Millisecond, // period
		time.Millisecond, // max latency
		[]Packet{p(1, 0), p(2, 0), p(1, 0)},
		[]uint32{1, 2, TO})
}

func TestReceive(t *testing.T) {
	// Golden case
	expectIsoc(t,
		1,                // duplication
		time.Millisecond, // period
		time.Millisecond, // max latency
		[]Packet{p(1, 0), p(2, 0), p(3, 0), p(4, 0), p(5, 0), p(6, 0), p(7, 0), p(8, 0), p(9, 0), p(10, 0), p(11, 0)},
		[]uint32{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, TO})
}

func TestReceiveOOOWithBigWindow(t *testing.T) {
	// Out of order with big window
	expectIsoc(t,
		1,                  // duplication
		time.Millisecond,   // period
		4*time.Millisecond, // max latency
		[]Packet{p(1, 0), p(2, 0), p(3, 0), p(1, 0), p(1, 0), p(7, 0), p(5, 0), p(6, 0), p(4, 0)},
		[]uint32{1, 2, 3, 4, 5, 6, 7, TO})
}

func TestReceiveOOOWithSmallWindow(t *testing.T) {
	// Out of order with small window. 6 will be dropped because out of window
	expectIsoc(t,
		1,                  // duplication
		time.Millisecond,   // period
		1*time.Millisecond, // max latency
		[]Packet{p(1, 0), p(2, 0), p(6, 0), p(5, 0), p(4, 0), p(3, 0)},
		[]uint32{1, 2, 3, 4, 5, TO})
}

func TestReceiveWithHighDuplication(t *testing.T) {

	expectIsoc(t,
		20,                 // duplication
		time.Millisecond,   // period
		1*time.Millisecond, // max latency
		[]Packet{p(1, 0), p(2, 0), p(3, 0), p(5, 0), p(6, 0), p(4, 0)},
		[]uint32{1, 2, 3, 4, 5, 6, TO})
}

func TestReceiveWithTimeout(t *testing.T) {
	// Out of order with small window. 6 will be dropped because out of window
	expectIsoc(t,
		1,                    // duplication
		time.Microsecond,     // period
		500*time.Microsecond, // max latency
		[]Packet{p(1, 0), p(3, 600), p(4, 0), p(5, 0)},
		[]uint32{1, TO, 3, 4, 5, TO})
}

func TestShouldReturn0DeadlineBeforeRead(t *testing.T) {
	rx, _ := NewRxIsochronous("udp", "127.0.0.1", 8888,
		1*time.Millisecond, // period
		2*time.Millisecond) // max latency

	if !rx.NextDeadlineFromNow().IsZero() {
		t.Errorf("Deadline was not 0")
	}
	rx.Close()
}

func TestShouldSetAppropriateDeadline(t *testing.T) {
	period := 1 * time.Millisecond
	latency := 2 * time.Millisecond
	fudge := 100 * time.Microsecond

	rx, err := NewRxIsochronous("udp", "127.0.0.1", 8888,
		1*time.Millisecond, // period
		2*time.Millisecond) // max latency
	if err != nil {
		t.Error(err)
	}
	sendIsoc(t, 1, []Packet{p(1, 0)})
	rx.Read()
	deadline := rx.NextDeadlineFromNow().Sub(time.Now())

	if deadline > latency+period || deadline < (latency+period)-fudge {
		t.Errorf("deadline was %d", deadline)
	}
}

func TestTCP(t *testing.T) {
	tx, err := NewTx("tcp", "127.0.0.1", 8888)
	if err != nil {
		t.Error(err)
	}
	rx, err := NewRxIsochronous("tcp", "127.0.0.1", 8888, 100*time.Microsecond, 1*time.Millisecond)
	if err != nil {
		t.Error(err)
	}
	f := makeFrame(1)
	tx.Write(f.Metadata, f.Data)
	rxf, err := rx.Read()
	if err != nil {
		t.Error(err)
	}
	if extractDataPayload(rxf.Data) != 1 {
		t.Errorf("Received incorrect frame")
	}
}
