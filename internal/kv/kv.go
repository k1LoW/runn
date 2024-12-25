package kv

import "sync"

type KV struct {
	mu sync.RWMutex
	m  map[string]any
}

func New() *KV {
	return &KV{m: map[string]any{}}
}

func (kv *KV) Set(k string, v any) {
	kv.mu.Lock()
	defer kv.mu.Unlock()
	kv.m[k] = v
}

func (kv *KV) Get(k string) any { //nostyle:getters
	kv.mu.RLock()
	defer kv.mu.RUnlock()
	v, ok := kv.m[k]
	if !ok {
		return nil
	}
	return v
}

func (kv *KV) Keys() []string {
	kv.mu.RLock()
	defer kv.mu.RUnlock()
	keys := make([]string, 0, len(kv.m))
	for k := range kv.m {
		keys = append(keys, k)
	}
	return keys
}

func (kv *KV) Del(k string) {
	kv.mu.RLock()
	defer kv.mu.RUnlock()
	delete(kv.m, k)
}

func (kv *KV) Clear() {
	kv.mu.Lock()
	defer kv.mu.Unlock()
	kv.m = map[string]any{}
}
