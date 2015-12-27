package buffer_pool

import (
	"testing"
)

import (
	"math/rand"
	"runtime"
	"strings"
	"sync"
	"time"
)

func TestBufferPool(t *testing.T) {
	runtime.GOMAXPROCS(runtime.NumCPU())
	var wg sync.WaitGroup
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			for j := 0; j < 100; j++ {
				pool := Get()
				buf := pool.GetByteSlice(10)
				for k := 0; k < 10; k++ {
					buf[k] = byte(rand.Int() % 128)
				}
				arr := string(buf)
				time.Sleep(time.Microsecond * time.Duration((rand.Int() % 100)))
				str := string(buf)
				if !strings.EqualFold(arr, str) {
					t.Errorf("buffer_pool test: pool data changed!")
				}
				Put(pool)
			}
			wg.Done()
		}()
	}

	wg.Wait()
	for i := 0; i < len(pools.ids); i++ {
		if i != pools.pools[pools.ids[i]].id {
			t.Errorf("buffer_pool test not pass!")
			return
		}
	}
}
