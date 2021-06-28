package utils

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAtomicInt32Counter(t *testing.T) {
	var a AtomicInt32Counter
	var wg sync.WaitGroup

	wg.Add(10)
	for i := 0; i < 10; i++ {
		go func() {
			defer wg.Done()
			a.Increment(int32(1))
		}()
	}
	wg.Wait()
	assert.Equal(t, int32(10), a.Count())
}
