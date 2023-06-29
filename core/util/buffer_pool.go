package util

import "bytes"

var bufferPool = newBufferPool()

func newBufferPool() chan *bytes.Buffer {
	bp := make(chan *bytes.Buffer, 100)
	for i := 0; i < 100; i++ {
		bp <- &bytes.Buffer{}
	}
	return bp
}

func acquireBuffer() *bytes.Buffer {
	return <-bufferPool
}

func releaseBuffer(b *bytes.Buffer) {
	b.Reset()
	bufferPool <- b
}
