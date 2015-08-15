package codereg

import (
	"crypto/rand"
	"fmt"
	"sync"
	"time"
)

type Reg struct {
	ItemTTL time.Duration

	keyLen         int
	vacuumInterval time.Duration
	items          map[string]record
	mutex          sync.RWMutex
}

type record struct {
	Created time.Time
	Value   interface{}
}

func New(keyLen int, itemTTL, vacuumInterval time.Duration) *Reg {
	r := &Reg{
		keyLen:         keyLen,
		vacuumInterval: vacuumInterval,
		ItemTTL:        itemTTL,
		items:          make(map[string]record),
	}
	r.run()
	return r
}

func (r *Reg) Add(val interface{}) (key string) {
	b := make([]byte, r.keyLen)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	key = fmt.Sprintf("%x", b)
	r.mutex.Lock()
	r.items[key] = record{Created: time.Now(), Value: val}
	r.mutex.Unlock()
	return
}

func (r *Reg) Get(key string) (val interface{}, found bool) {
	r.mutex.RLock()
	rec, found := r.items[key]
	r.mutex.RUnlock()
	if found {
		if rec.Created.Before(time.Now().Add(-r.ItemTTL)) {
			found = false
		} else {
			val = rec.Value
		}
	}
	return
}

func (r *Reg) Del(key string) {
	r.mutex.Lock()
	delete(r.items, key)
	r.mutex.Unlock()
}

func (r *Reg) run() {
	go func() {
		// Vacuum
		for {
			time.Sleep(r.vacuumInterval)
			cutTime := time.Now().Add(-r.ItemTTL)
			r.mutex.Lock()
			for t, it := range r.items {
				if it.Created.Before(cutTime) {
					delete(r.items, t)
				}
			}
			r.mutex.Unlock()
		}
	}()

}
