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
var RepublishTime = 24 * time.Hour

type kvData struct {
    id            KademliaID
    data          []byte
    evictionTime  time.Time
    pinned        bool
    republishTime time.Time
    republishFunc func(id *KademliaID)
}

// For REST/debug only
type KVPair struct {
    Hash KademliaID
    Data kvData
}

type KVStore struct {
    evictionTimer  *time.Timer
    evictionQueue  []*kvData
    republishTimer *time.Timer
    republishQueue []*kvData
    mapping        map[KademliaID]*kvData
    mutex          *sync.Mutex
}

func NewKVStore() *KVStore {
    kvStore := new(KVStore)
    kvStore.mutex = &sync.Mutex{}
    kvStore.mapping = make(map[KademliaID]*kvData)

    kvStore.republishQueue = []*kvData{}
    kvStore.republishTimer = time.NewTimer(0)
    <-kvStore.republishTimer.C
    go kvStore.republishThread()

    kvStore.evictionQueue = []*kvData{}
    kvStore.evictionTimer = time.NewTimer(0)
    <-kvStore.evictionTimer.C
    go kvStore.evictionThread()

    return kvStore
}

func (kvStore *KVStore) scheduleRepublish(data *kvData) {
    kvStore.republishQueue = append(kvStore.republishQueue, data)
    newDuration := data.republishTime.Sub(time.Now())
    // If the republish thread was idle before, restart it
    if len(kvStore.republishQueue) == 1 {
        kvStore.republishTimer.Reset(newDuration)
    }
    if data.republishFunc != nil {
        fmt.Println("Republishing", data.id.String(), "in", newDuration.String(), ", queue size", len(kvStore.evictionQueue))
    }
}

func (kvStore *KVStore) republishThread() {
    for {
        <-kvStore.republishTimer.C
        if len(kvStore.republishQueue) > 0 {
            toRepublish := kvStore.republishQueue[0]
            kvStore.republishQueue = kvStore.republishQueue[1:]
            if _, err := kvStore.Lookup(toRepublish.id); err == nil && toRepublish.republishFunc != nil {
                // Store contains this entry, republish it and add last in queue
                fmt.Println("Republishing", toRepublish.id.String())
                toRepublish.republishFunc(&toRepublish.id)
                toRepublish.republishTime = time.Now().Add(RepublishTime)
                kvStore.republishQueue = append(kvStore.republishQueue, toRepublish)
            } else {
                // This entry was removed from store before republish timeout, or has no republish function
            }
        }
        if len(kvStore.republishQueue) > 0 {
            newDuration := kvStore.republishQueue[0].republishTime.Sub(time.Now())
            if newDuration < 0 {
                newDuration = 0
            }
            fmt.Println("Next republish scheduled in", newDuration)
            kvStore.republishTimer.Reset(newDuration)
        }
    }

}

func (kvStore *KVStore) scheduleEviction(data *kvData) {
    kvStore.evictionQueue = append(kvStore.evictionQueue, data)
    newDuration := data.evictionTime.Sub(time.Now())
    // If the eviction thread was idle before, restart it
    if len(kvStore.evictionQueue) == 1 {
        kvStore.evictionTimer.Reset(newDuration)
    }
    fmt.Println("Evicting", data.id.String(), "in", newDuration.String(), ", queue size", len(kvStore.evictionQueue))
}

func (kvStore *KVStore) evictionThread() {
    for {
        <-kvStore.evictionTimer.C
        fmt.Println("Eviction timeout...")
        kvStore.mutex.Lock()
        if len(kvStore.evictionQueue) > 0 {
            toEvict := kvStore.evictionQueue[0]
            kvStore.evictionQueue = kvStore.evictionQueue[1:]
            if !toEvict.pinned {
                // Remove from store
                delete(kvStore.mapping, toEvict.id)
                fmt.Println("Evicted unpinned", toEvict.id.String())
            } else {
                fmt.Println("Ignoring pinned", toEvict.id.String())
            }
        } else {
            fmt.Println("Eviction queue empty...")
        }
        if len(kvStore.evictionQueue) > 0 {
            newDuration := kvStore.evictionQueue[0].evictionTime.Sub(time.Now())
            if newDuration < 0 {
                newDuration = 0
            }
            fmt.Println("Next eviction scheduled in", newDuration)
            kvStore.evictionTimer.Reset(newDuration)
        }
        kvStore.mutex.Unlock()
    }
}

func (kvStore *KVStore) Insert(hash KademliaID, pinned bool, data []byte,
    republishFunc func(*KademliaID)) (outData kvData, err error) {
    kvStore.mutex.Lock()
    if kvStore.mapping == nil {
        err = NotInitializedError
    } else {
        outData = kvData{id: hash, data: data, pinned: pinned, evictionTime: time.Now().Add(EvictionTime),
            republishTime: time.Now().Add(RepublishTime), republishFunc: republishFunc}
        kvStore.mapping[hash] = &outData
        kvStore.scheduleEviction(&outData)
        kvStore.scheduleRepublish(&outData)
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
        val.evictionTime = time.Now().Add(EvictionTime)
        kvStore.scheduleEviction(val)
        kvStore.mapping[hash] = val
    } else {
        err = NotFoundError
    }
    kvStore.mutex.Unlock()
    return
}

// Grab all data in the table and dump it
func (kvStore *KVStore) DumpStore() []KVPair {
    els := make([]KVPair, len(kvStore.mapping))

    i := 0
    for k, v := range kvStore.mapping {
        els[i] = KVPair{k, *v}
        i++
    }

    return els
}
