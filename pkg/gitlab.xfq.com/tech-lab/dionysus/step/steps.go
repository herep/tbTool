package step

import (
	"fmt"
	"log"
	"reflect"
	"sync"

	"gitlab.xfq.com/tech-lab/dionysus/pkg/algs"
)

type Steps struct {
	q *algs.PriorityQueue // 全局启动依赖项
	m sync.Map
}

func New() *Steps {
	return &Steps{
		q: algs.GetPQ(),
		m: sync.Map{},
	}
}

func (s *Steps) RegActionSteps(value string, priority int, fn func() error) {
	item := algs.GetItem(value, priority)
	s.m.Store(item, fn)
	s.q.Push(item)
}

func (s *Steps) RegActionStepsE(value string, priority int, fn func() error) error {
	if priority < 0 {
		return fmt.Errorf(" Priority can not be negtive: %d ", priority)
	}

	if fn == nil {
		return fmt.Errorf(" Function can not be nil: %T ", fn)
	}

	s.RegActionSteps(value, priority, fn)
	return nil
}

// 初始化加载router middle afterstart等等
func (s *Steps) Run() error {
	// Take the items out; they arrive in decreasing priority order.
	i := 1
	pqLen := s.q.Len()
	for s.q.Len() > 0 {
		item, _ := s.q.Pop()
		if fn, ok := s.m.Load(item); ok && !reflect.ValueOf(fn).IsNil() {
			if f, ok := fn.(func() error); ok {
				if err := f(); err != nil {
					ef := fmt.Errorf("[step %d/%d] %s err: %v", i, pqLen, item.Value(), err)
					log.Print(ef)
					return ef
				} else {
					log.Printf("[step %d/%d] %s success", i, pqLen, item.Value())
				}
			}
		} else {
			log.Printf("[warn] load step false %v \n", item)
		}
		i++
	}
	return nil
}
