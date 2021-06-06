package cosmap

import "sync"

type ArrayKey uint64
type ArrayVal interface {
	SetArrayKey(ArrayKey)
	GetArrayKey() ArrayKey
}

func NewArray(cap int) *Array {
	arrayMap := &Array{
		seed:   1,
		dirty:  NewArrayIndex(cap),
		values: make([]ArrayVal, cap, cap),
	}
	for i := cap - 1; i >= 0; i-- {
		arrayMap.dirty.Add(i)
	}
	return arrayMap
}

type Array struct {
	seed   uint32 //ID 生成种子
	mutex  sync.Mutex
	dirty  *ArrayIndex
	values []ArrayVal
}

//createSocketId 使用index生成ID
func (s *Array) createId(index int) ArrayKey {
	s.seed++
	return ArrayKey(index)<<32 | ArrayKey(s.seed)
}

//parseSocketId 返回idPack中的index
func (s *Array) parseId(id ArrayKey) int {
	if id == 0 {
		return -1
	}
	return int(id >> 32)
}

//Get 获取
func (s *Array) Get(id ArrayKey) ArrayVal {
	index := s.parseId(id)
	if index < 0 || index >= len(s.values) || s.values[index].GetArrayKey() != id {
		return nil
	}
	return s.values[index]
}

func (s *Array) Set(v ArrayVal) ArrayKey {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	var index = -1
	if index = s.dirty.Get(); index >= 0 {
		s.values[index] = v
	} else {
		index = len(s.values)
		s.values = append(s.values, v)
	}
	id := s.createId(index)
	v.SetArrayKey(id)
	return id
}

//Delete 删除
func (s *Array) Delete(id ArrayKey) bool {
	index := s.parseId(id)
	if index < 0 || index >= len(s.values) || s.values[index].GetArrayKey() != id {
		return false
	}
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.values[index] = nil
	s.dirty.Add(index)
	return true
}

//Size 当前数量
func (s *Array) Size() int {
	return len(s.values) - s.dirty.Size()
}

//遍历
func (s *Array) Range(f func(interface{})) {
	for _, val := range s.values {
		if val != nil {
			f(val)
		}
	}
}