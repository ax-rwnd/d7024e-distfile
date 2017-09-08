package d7024e

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
var KVStore map[KademliaID]kvData
//var LastCheck time.Time

type kvData struct {
    data []byte
    timestamp time.Time
    pinned bool
}

// Manage the initialization of the storage
func KVSInit() (err error) {
    if KVStore != nil {
        return errors.New("KVS already initialized!")
    }
    KVStore = make(map[KademliaID]kvData)
    return
}

// Evict data after some time
func Evict (now time.Time) (err error) {
    for key, value := range KVStore {
        if now.Sub(value.timestamp) > EvictionTime && !value.pinned {
            delete(KVStore, key) //TODO: check for runtimes
        }
    }
    return
}

// Don't silently update duplicate data (in case of collision)
func KVSInsert(hash KademliaID, pinned bool, data []byte) (err error) {
    if KVStore == nil {
        err = NotInitializedError
    } else if _, ok := KVStore[hash]; ok {
        err = DuplicateError
    } else {
        KVStore[hash] = kvData{data: data, timestamp: time.Now(), pinned: pinned}
    }
    return
}

// Lookup data from table
func KVSLookup (hash KademliaID) (output kvData, err error) {
    if val, ok := KVStore[hash]; ok {
        output = val
    } else {
        err = NotFoundError
    }
    return
}
