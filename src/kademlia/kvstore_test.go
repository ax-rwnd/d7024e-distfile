package kademlia

import (
    "testing"
    "log"
)

func TestKVSInsertLookup(t *testing.T) {
    kvStore := NewKVStore()
    data := []byte("Test data")
    id := NewKademliaIDFromBytes(data)
    pinned := false
    _, err := kvStore.Insert(*id, pinned, data)
    if err != nil {
        log.Println(err)
        t.Fail()
    }
    storedData, err := kvStore.Lookup(*id)
    if err != nil {
        log.Println(err)
        t.Fail()
    }
    for i := range data {
        if data[i] != storedData[i] {
            t.Fail()
        }
    }
}

func TestNotFoundError(t *testing.T) {
    kvStore := NewKVStore()
    data := []byte("Test data")
    id := NewKademliaIDFromBytes(data)
    // Lookup without inserting first
    _, err := kvStore.Lookup(*id)
    if err != NotFoundError {
        log.Println("Wrong error")
        log.Println(err)
        t.Fail()
    }
}

func TestPinUnpin(t *testing.T) {
    kvStore := NewKVStore()

    data := []byte("Test data")
    id := NewKademliaIDFromBytes(data)
    kvStore.Insert(*id, false, data)

    kvStore.Pin(*id)
    if val, _ := kvStore.mapping[*id]; val.pinned == false {
        log.Println("Failed to pin content.")
        t.Fail()
    }

    kvStore.Unpin(*id)
    if val, _ := kvStore.mapping[*id]; val.pinned == true {
        log.Println("Failed to unpin content.")
        t.Fail()
    }

}
