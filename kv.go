package runn

import "sync"

type kv struct {
	mu sync.RWMutex
	m  map[string]any
}

func newKV() *kv {
	return &kv{m: map[string]any{}}
}

func (kv *kv) set(k string, v any) {
	kv.mu.Lock()
	defer kv.mu.Unlock()
	kv.m[k] = v
}

func (kv *kv) get(k string) any { //nostyle:getters
	kv.mu.RLock()
	defer kv.mu.RUnlock()
	v, ok := kv.m[k]
	if !ok {
		return nil
	}
	return v
}
