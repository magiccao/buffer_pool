package buffer_pool

import (
	"sync"
)

const (
	KB = 1 << 10
	MB = KB << 10
)

const (
	M = 200 * KB // capacity of byte buffer
	N = 3000     // number of byte buffers
)

type Pool struct {
	data  [M]byte    // pool byte slice
	id    int        // pool id in ids
	mu    sync.Mutex // guard following data
	unidx int        // unused index
}

type Pools struct {
	mtx   sync.Mutex
	pools [N]Pool
	ids   [N]int
	inuse int // the first inused object index
	unuse int // the first unused object index
	left  int // the left capacity for pool
}

var pools Pools

func init() {
	for i := 0; i < N; i++ {
		pools.ids[i] = i
		pools.pools[i].id = i
	}
	pools.inuse = -1
	pools.unuse = -1
	pools.left = N
}

func NewPool() *Pool {
	return &Pool{id: -1}
}

func Get() *Pool {
	return get()
}

func Put(pool *Pool) {
	if pool != nil {
		// if the pool is not come from pools, but NewPool,
		// return
		if pool.id < 0 {
			return
		}
		put(pool)
	}
}

func get() *Pool {
	pools.mtx.Lock()
	defer pools.mtx.Unlock()

	if pools.left > 0 {
		pools.left -= 1
		pools.unuse += 1
		if pools.unuse == N {
			pools.unuse = 0
		}

		id := pools.ids[pools.unuse]
		return &pools.pools[id]
	}

	return NewPool()
}

func put(pool *Pool) {
	pools.mtx.Lock()
	defer pools.mtx.Unlock()

	pools.left += 1
	pools.inuse += 1
	if pools.inuse == N {
		pools.inuse = 0
	}

	// if the free pool is the inuse point,
	// do not need do swap
	if pool.id == pools.inuse {
		return
	}

	// swap pools index
	idx := pools.ids[pools.inuse]
	pools.ids[pools.inuse] = pools.ids[pool.id]
	pools.ids[pool.id] = idx

	// swap ids index
	pools.pools[idx].id = pool.id
	pool.id = pools.inuse

	// clear the pool
	// do not need pool.mu.Lock guard
	pool.unidx = 0
}

func (pool *Pool) GetByteSlice(size int) []byte {
	if pool == nil {
		return make([]byte, size)
	}

	pool.mu.Lock()
	defer pool.mu.Unlock()
	if len(pool.data)-pool.unidx < size {
		return make([]byte, size)
	}

	b := pool.data[pool.unidx : pool.unidx+size]
	pool.unidx += size
	return b
}
