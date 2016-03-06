package streamcast

import (
	"log"
)

type FrameCache struct {
	currentFrame uint32
	cache        []*Frame
	cacheSize    uint32
}

func NewFrameCache(cacheSize uint32) (fc *FrameCache) {
	fc = new(FrameCache)
	fc.cacheSize = cacheSize
	fc.cache = make([]*Frame, cacheSize)
	return fc
}

func (fc *FrameCache) FastForwardTo(frameId uint32) {
	// Clear cached frames that we advanced over.
	maxToClear := frameId - fc.currentFrame
	if debug {
		log.Printf("FFWD from %d to %d", fc.currentFrame, frameId)
	}
	// Don't clear more than 1 window.
	if maxToClear > fc.cacheSize {
		maxToClear = fc.cacheSize
	}

	for i := fc.currentFrame; i < fc.currentFrame+maxToClear; i++ {
		if debug {
			log.Printf("Clearing %d idx %d \n", i, i%fc.cacheSize)
		}
		fc.cache[i%fc.cacheSize] = nil
	}
	fc.currentFrame = frameId
}

func (fc *FrameCache) Get(frameId uint32) (f *Frame) {
	if fc.currentFrame != frameId {
		fc.FastForwardTo(frameId)
	}

	cacheIndex := fc.currentFrame % fc.cacheSize
	if fc.cache[cacheIndex] != nil {
		f := fc.cache[cacheIndex]
		fc.cache[cacheIndex] = nil
		if debug {
			log.Printf("Returning from cache %d idx %d \n", fc.currentFrame, cacheIndex)
		}
		fc.currentFrame++
		return f
	}
	return nil
}

func (fc *FrameCache) Put(f *Frame) {
	// Drop frames so far in the future, they're outside our window
	distanceFromNow := f.FrameId - fc.currentFrame
	if debug {
		log.Printf("cache dist from now == %d; cache size == %d", distanceFromNow, fc.cacheSize)
	}
	if distanceFromNow >= fc.cacheSize {
		if debug {
			log.Printf("Dropped received frame (data coming faster than expected)")
		}
		return
	}

	if debug {
		log.Printf("Setting %d idx %d\n", f.FrameId, f.FrameId%fc.cacheSize)
	}
	cacheidx := f.FrameId % fc.cacheSize
	if fc.cache[cacheidx] == nil {
		fc.cache[cacheidx] = f
	}
}
