package pool

import (
	"github.com/panjf2000/ants/v2"
	"sync"
)

type RunPool struct {
	sync.WaitGroup
	p *ants.Pool
}

func NewPool(max int) *RunPool {
	p, _ := ants.NewPool(max)
	return &RunPool{
		p: p,
	}
}

func (e *RunPool) Submit(f func()) {
	e.Add(1)
	e.p.Submit(func() {
		f()
		e.Done()
	})
}

func (e *RunPool) Release() {
	e.p.Release()
}
