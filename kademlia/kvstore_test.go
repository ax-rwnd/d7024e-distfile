package kademlia

import (
    "testing"
    "fmt"
)

func TestKVSInsertLookup(t *testing.T) {
    kvStore := NewKVStore()
    data := []byte("Test data")
    id := NewKademliaIDFromBytes(data)
    pinned := false
    _, err := kvStore.Insert(*id, pinned, data)
    if err != nil {
        fmt.Println(err)
        t.Fail()
    }
    storedData, err := kvStore.Lookup(*id)
    if err != nil {
        fmt.Println(err)
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
        fmt.Println("Wrong error")
        fmt.Println(err)
        t.Fail()
    }
}

func TestDuplicateError(t *testing.T) {
    kvStore := NewKVStore()
    data := []byte("Test data")
    id := NewKademliaIDFromBytes(data)
    pinned := false
    _, err := kvStore.Insert(*id, pinned, data)
    if err != nil {
        fmt.Println(err)
        t.Fail()
    }
    _, err = kvStore.Insert(*id, pinned, data)
    if err != DuplicateError {
        fmt.Println("Wrong error")
        fmt.Println(err)
        t.Fail()
    }
}
