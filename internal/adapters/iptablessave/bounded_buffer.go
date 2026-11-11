package iptablessave

import "bytes"

// boundedBuffer keeps at most limit bytes and silently discards the rest, so a
// misbehaving process cannot exhaust memory.
type boundedBuffer struct {
	buffer bytes.Buffer
	limit  int
	// truncated becomes true as soon as a byte had to be dropped.
	truncated bool
}

// Write implements io.Writer; excess bytes are dropped but reported as written.
func (bounded *boundedBuffer) Write(chunk []byte) (int, error) {
	remaining := bounded.limit - bounded.buffer.Len()
	if remaining > 0 {
		if len(chunk) < remaining {
			remaining = len(chunk)
		}
		bounded.buffer.Write(chunk[:remaining])
	}
	if remaining < len(chunk) {
		bounded.truncated = true
	}
	return len(chunk), nil
}

// Bytes returns the retained bytes.
func (bounded *boundedBuffer) Bytes() []byte {
	return bounded.buffer.Bytes()
}
