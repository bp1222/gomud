package game

import "sync"

type (
	SafeSlice interface {
		Add(interface{})
		Remove(interface{})
		Foreach(func(interface{}))
		Wait()
	}

	safeSliceImpl struct {
		list []interface{}
		mtx sync.Mutex
		grp sync.WaitGroup
	}
)

func NewSafeSlice() SafeSlice {
	return &safeSliceImpl{}
}

func (s *safeSliceImpl) Foreach(f func(interface{})) {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	for _, item := range s.list {
		f(item)
	}
}

func (s *safeSliceImpl) Add(i interface{}) {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	s.list = append(s.list, i)
	s.grp.Add(1)
}

func (s *safeSliceImpl) Remove(i interface{}) {
	s.mtx.Lock()
	defer s.grp.Done()
	defer s.mtx.Unlock()
	for idx, item := range s.list {
		if item == i {
			s.list = append(s.list[0:idx], s.list[idx+1:]...)
			break
		}
	}
}

func (s *safeSliceImpl) Wait() {
	s.grp.Wait()
}
