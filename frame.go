package streamcast

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type Frame struct {
	// TODO: Implement rollover or migrate frameid to uint64
	FrameId  uint32
	Metadata []byte
	Data     []byte
}

const MAX_FRAME_LENGTH = 1400

func (f *Frame) Read(b []byte) (err error) {
	var length uint16
	if len(b) > MAX_FRAME_LENGTH {
		return fmt.Errorf("Frame larger than max frame length")
	}
	buf := bytes.NewBuffer(b)

	if err = binary.Read(buf, binary.BigEndian, &f.FrameId); err != nil {
		return err
	}
	if err = binary.Read(buf, binary.BigEndian, &length); err != nil {
		return err
	}
	f.Metadata = buf.Next(int(length))
	if err = binary.Read(buf, binary.BigEndian, &length); err != nil {
		return err
	}
	f.Data = buf.Next(int(length))
	return
}

func (f *Frame) Write(b []byte) (n int, err error) {
	var length uint16
	var buf bytes.Buffer
	if len(b) < MAX_FRAME_LENGTH {
		return 0, fmt.Errorf("Input buffer too small")
	}
	if len(f.Metadata)+len(f.Data)+8 > MAX_FRAME_LENGTH { // 2xuint16 + 1xuint32
		return 0, fmt.Errorf("Frame larger than max frame length")
	}
	if err = binary.Write(&buf, binary.BigEndian, &f.FrameId); err != nil {
		return 0, err
	}
	length = uint16(len(f.Metadata))
	if err = binary.Write(&buf, binary.BigEndian, &length); err != nil {
		return 0, err
	}
	if err = binary.Write(&buf, binary.BigEndian, f.Metadata); err != nil {
		return 0, err
	}
	length = uint16(len(f.Data))
	if err = binary.Write(&buf, binary.BigEndian, &length); err != nil {
		return 0, err
	}
	if err = binary.Write(&buf, binary.BigEndian, f.Data); err != nil {
		return 0, err
	}
	bufbytes := buf.Bytes()
	copy(b, bufbytes)
	return len(bufbytes), nil
}
