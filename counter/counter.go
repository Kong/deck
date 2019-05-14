package counter

import "sync/atomic"

// Counter can be incremeneted and read.
// It is safe to use concurrently.
type Counter uint64

// Inc can be used to increment the value.
func (c *Counter) Inc() uint64 {
	return atomic.AddUint64((*uint64)(c), 1)
}

// Value returns the value of the counter.
func (c *Counter) Value() uint64 {
	return atomic.LoadUint64((*uint64)(c))
}

// Reset resets the counter to 0.
func (c *Counter) Reset() {
	atomic.StoreUint64((*uint64)(c), 0)
}
