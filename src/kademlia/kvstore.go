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

type kvTimeToLiveQueue []kvData

type KVStore struct {
    timer           *time.Timer
    timeToLiveQueue kvTimeToLiveQueue
    mapping         map[KademliaID]kvData
    mutex           *sync.Mutex
}

func (s kvTimeToLiveQueue) Len() int {
    return len(s)
}

func (s kvTimeToLiveQueue) Swap(i, j int) {
    s[i], s[j] = s[j], s[i]
}

func (s kvTimeToLiveQueue) Less(i, j int) bool {
    return s[i].timeToLive.Before(s[j].timeToLive)
}

func NewKVStore() *KVStore {
    kvStore := new(KVStore)
    kvStore.mutex = &sync.Mutex{}
    kvStore.mapping = make(map[KademliaID]kvData)
    kvStore.timeToLiveQueue = []kvData{}
    kvStore.timer = time.NewTimer(0)
    <-kvStore.timer.C
    go kvStore.eviction()
    return kvStore
}

func (kvStore *KVStore) eviction() {
    for {
        <-kvStore.timer.C
        kvStore.mutex.Lock()
        fmt.Println("Checking")
        if len(kvStore.timeToLiveQueue) > 0 {
            toEvict := kvStore.timeToLiveQueue[0]
            kvStore.timeToLiveQueue = kvStore.timeToLiveQueue[1:]
            delete(kvStore.mapping, toEvict.id)
            fmt.Printf("Evicted %v\n", toEvict.id.String())
        }
        if len(kvStore.timeToLiveQueue) > 0 {
            nextEvictionDuration := kvStore.timeToLiveQueue[0].timeToLive.Sub(time.Now())
            kvStore.timer.Reset(nextEvictionDuration)
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
        if !pinned {
            kvStore.timeToLiveQueue = append(kvStore.timeToLiveQueue, outData)
            if len(kvStore.timeToLiveQueue) == 1 {
                nextEvictionDuration := kvStore.timeToLiveQueue[0].timeToLive.Sub(time.Now())
                kvStore.timer.Reset(nextEvictionDuration)
            }
        }
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
