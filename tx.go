package streamcast

import (
	"time"
)

type Transmitter interface {
  Write(metadata []byte, data []byte) (err error)
  SetTimeout(t time.Duration)
  Close()
}
