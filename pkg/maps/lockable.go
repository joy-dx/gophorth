package maps

import (
	"encoding/json"
	"fmt"
	"sync"
)

type Lockable[T ~string, T2 any] struct {
	Map map[T]T2 `json:"map" yaml:"map"`
	mu  *sync.RWMutex
}

func NewLockableMap[T ~string, T2 any]() Lockable[T, T2] {
	return Lockable[T, T2]{
		Map: make(map[T]T2),
		mu:  &sync.RWMutex{},
	}
}

func (v Lockable[T, T2]) MarshalJSON() ([]byte, error) {
	jsonData, err := json.Marshal(v.Map)
	if err != nil {
		return nil, fmt.Errorf("error serializing lockable map to JSON: %w", err)
	}
	return jsonData, nil
}

func (v *Lockable[T, T2]) Lock() {
	v.mu.Lock()
}

func (v *Lockable[T, T2]) Unlock() {
	v.mu.Unlock()
}

func (v *Lockable[T, T2]) Set(key T, entry T2) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.Map[key] = entry
}

func (v *Lockable[T, T2]) Remove(key T) {
	v.mu.Lock()
	defer v.mu.Unlock()
	delete(v.Map, key)
}

func (v *Lockable[T, T2]) Get(key T) (T2, error) {
	v.mu.RLock()
	defer v.mu.RUnlock()
	if val, found := v.Map[key]; found {
		return val, nil
	}
	var t T2
	return t, &KeyNotFoundError{Key: string(key)}
}

func (v *Lockable[T, T2]) GetAll() map[T]T2 {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.Map
}
func (v *Lockable[T, T2]) GetAllSlice() []T2 {
	v.mu.RLock()
	defer v.mu.RUnlock()
	entries := make([]T2, len(v.Map))
	idx := 0
	for _, mapValue := range v.Map {
		entries[idx] = mapValue
		idx++
	}
	return entries
}
