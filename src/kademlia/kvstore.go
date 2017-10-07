package kademlia

import (
    "time"
    "errors"
    "sync"
)

// Error states
var NotInitializedError = errors.New("KVS not initialized")
var DuplicateError = errors.New("value is already in map")
var NotFoundError = errors.New("value was not found in map")

// Globals
var EvictionTime, _ = time.ParseDuration("1h")

type KVStore struct {
    mapping map[KademliaID]kvData
    mutex   *sync.Mutex
}

type kvData struct {
    data      []byte
    timestamp time.Time
    pinned    bool
}

func NewKVStore() *KVStore {
    kvStore := new(KVStore)
    kvStore.mutex = &sync.Mutex{}
    kvStore.mapping = make(map[KademliaID]kvData)
    return kvStore
}

// Evict data after some time
func (kvStore KVStore) Evict(now time.Time) (err error) {
    kvStore.mutex.Lock()
    for key, value := range kvStore.mapping {
        if now.Sub(value.timestamp) > EvictionTime && !value.pinned {
            delete(kvStore.mapping, key) //TODO: check for runtimes
        }
    }
    kvStore.mutex.Unlock()
    return
}

// Don't silently update duplicate data (in case of collision)
func (kvStore KVStore) Insert(hash KademliaID, pinned bool, data []byte) (outData kvData, err error) {
    kvStore.mutex.Lock()
    if kvStore.mapping == nil {
        err = NotInitializedError
    } else {
        outData = kvData{data: data, timestamp: time.Now(), pinned: pinned}
        kvStore.mapping[hash] = outData
    }
    kvStore.mutex.Unlock()
    return
}

// Lookup data from table
func (kvStore KVStore) Lookup(hash KademliaID) (output []byte, err error) {
    kvStore.mutex.Lock()
    if val, ok := kvStore.mapping[hash]; ok {
        output = val.data
    } else {
        err = NotFoundError
    }
    kvStore.mutex.Unlock()
    return
}

func (kvStore KVStore) Pin(hash KademliaID) (err error) {
    kvStore.mutex.Lock()
    if val, ok := kvStore.mapping[hash]; ok {
        val.pinned = true
        kvStore.mapping[hash] = val
    } else {
        err = NotFoundError
    }
    kvStore.mutex.Unlock()
    return
}

func (kvStore KVStore) Unpin(hash KademliaID) (err error) {
    kvStore.mutex.Lock()
    if val, ok := kvStore.mapping[hash]; ok {
        val.pinned = false
        kvStore.mapping[hash] = val
    } else {
        err = NotFoundError
    }
    kvStore.mutex.Unlock()
    return
}
