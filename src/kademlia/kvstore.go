package kademlia

import (
    "time"
    "errors"
    "sync"
    "fmt"
)

// Error states
var NotInitializedError = errors.New("KVS not initialized")
var DuplicateError = errors.New("value is already in map")
var NotFoundError = errors.New("value was not found in map")

// Globals
var EvictionTime = 24 * time.Hour

type kvData struct {
    id         KademliaID
    data       []byte
    timeToLive time.Time
    pinned     bool
}

type KVStore struct {
    timer         *time.Timer
    evictionQueue []*kvData
    mapping       map[KademliaID]kvData
    mutex         *sync.Mutex
}

func NewKVStore() *KVStore {
    kvStore := new(KVStore)
    kvStore.mutex = &sync.Mutex{}
    kvStore.mapping = make(map[KademliaID]kvData)
    kvStore.evictionQueue = []*kvData{}
    kvStore.timer = time.NewTimer(0)
    <-kvStore.timer.C
    go kvStore.evictionThread()
    return kvStore
}

func (kvStore *KVStore) scheduleEviction(data *kvData) {
    kvStore.evictionQueue = append(kvStore.evictionQueue, data)
    // If the eviction thread was idle before, restart it
    if len(kvStore.evictionQueue) == 1 {
        newDuration := kvStore.evictionQueue[0].timeToLive.Sub(time.Now())
        kvStore.timer.Reset(newDuration)
        fmt.Println("Evicting", data.id.String(), "in", newDuration.String(),"eviction queue size", len(kvStore.evictionQueue))
    }
}

func (kvStore *KVStore) evictionThread() {
    for {
        <-kvStore.timer.C
        fmt.Println("Eviction timeout...")
        kvStore.mutex.Lock()
        if len(kvStore.evictionQueue) > 0 {
            toEvict := kvStore.evictionQueue[0]
            kvStore.evictionQueue = kvStore.evictionQueue[1:]
            if !toEvict.pinned {
                delete(kvStore.mapping, toEvict.id)
                fmt.Println("Evicted unpinned", toEvict.id.String())
            } else {
                fmt.Println("Ignoring pinned", toEvict.id.String())
            }
        } else {
            kvStore.timer.Stop()
            fmt.Println("Eviction queue empty...")
        }
        if len(kvStore.evictionQueue) > 0 {
            newDuration := kvStore.evictionQueue[0].timeToLive.Sub(time.Now())
            if newDuration < 0 {
                newDuration = 0
            }
            fmt.Println("Next eviction scheduled in",newDuration)
            kvStore.timer.Reset(newDuration)
        }
        kvStore.mutex.Unlock()
    }
}

// Don't silently update duplicate data (in case of collision)
func (kvStore *KVStore) Insert(hash KademliaID, pinned bool, data []byte) (outData kvData, err error) {
    kvStore.mutex.Lock()
    if kvStore.mapping == nil {
        err = NotInitializedError
    } else {
        outData = kvData{id: hash, data: data, timeToLive: time.Now().Add(EvictionTime), pinned: pinned}
        kvStore.mapping[hash] = outData
        kvStore.scheduleEviction(&outData)
    }
    kvStore.mutex.Unlock()
    return
}

// Lookup data from table
func (kvStore *KVStore) Lookup(hash KademliaID) (output []byte, err error) {
    kvStore.mutex.Lock()
    if val, ok := kvStore.mapping[hash]; ok {
        output = val.data
    } else {
        err = NotFoundError
    }
    kvStore.mutex.Unlock()
    return
}

func (kvStore *KVStore) Pin(hash KademliaID) (err error) {
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

func (kvStore *KVStore) Unpin(hash KademliaID) (err error) {
    kvStore.mutex.Lock()
    if val, ok := kvStore.mapping[hash]; ok {
        val.pinned = false
        val.timeToLive = time.Now().Add(EvictionTime)
        kvStore.scheduleEviction(&val)
        kvStore.mapping[hash] = val
    } else {
        err = NotFoundError
    }
    kvStore.mutex.Unlock()
    return
}
