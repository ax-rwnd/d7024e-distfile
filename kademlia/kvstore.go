package kademlia

import (
    "time"
    "errors"
)

// Error states
var NotInitializedError = errors.New("KVS not initialized.")
var DuplicateError = errors.New("Value is already in map.")
var NotFoundError = errors.New("Value was not found in map.")

// Globals
var EvictionTime, _ = time.ParseDuration("1h")

type KVStore map[KademliaID]kvData
type kvData struct {
    data      []byte
    timestamp time.Time
    pinned    bool
}

func NewKVStore() KVStore {
    return make(map[KademliaID]kvData)
}

// Evict data after some time
func (kvStore KVStore) Evict(now time.Time) (err error) {
    for key, value := range kvStore {
        if now.Sub(value.timestamp) > EvictionTime && !value.pinned {
            delete(kvStore, key) //TODO: check for runtimes
        }
    }
    return
}

// Don't silently update duplicate data (in case of collision)
func (kvStore KVStore) Insert(hash KademliaID, pinned bool, data []byte) (outData kvData, err error) {
    if kvStore == nil {
        err = NotInitializedError
    } else if _, ok := kvStore[hash]; ok {
        err = DuplicateError
    } else {
        outData = kvData{data: data, timestamp: time.Now(), pinned: pinned}
        kvStore[hash] = outData
    }
    return
}

// Lookup data from table
func (kvStore KVStore) Lookup(hash KademliaID) (output []byte, err error) {
    if val, ok := kvStore[hash]; ok {
        output = val.data
    } else {
        err = NotFoundError
    }
    return
}

func (kvStore KVStore) Clear() {
    for k := range kvStore {
        delete(kvStore, k)
    }
}
