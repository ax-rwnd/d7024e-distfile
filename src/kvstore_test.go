package d7024e

import (
    "testing"
    "fmt"
)

func TestInit(t *testing.T) {
    if err := KVSInit(); err != nil {
        fmt.Println("Initialization of KVStore failed.")
        t.Fail()
    }

    if err := KVSInit(); err == nil {
        fmt.Println("KVStore was silently reinitialized!")
        t.Fail()
    }
}

func TestKVSInsertLookup(t *testing.T) {
    KVSClear()
    data := []byte("Test data")
    id := NewKademliaIDFromBytes(data)
    pinned := false
    _, err := KVSInsert(*id, pinned, data)
    if err != nil {
        fmt.Println(err)
        t.Fail()
    }
    storedData, err := KVSLookup(*id)
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
    KVSClear()
    data := []byte("Test data")
    id := NewKademliaIDFromBytes(data)
    // Lookup without inserting first
    _, err := KVSLookup(*id)
    if err != NotFoundError {
        fmt.Println("Wrong error")
        fmt.Println(err)
        t.Fail()
    }
}

func TestDuplicateError(t *testing.T) {
    KVSClear()
    data := []byte("Test data")
    id := NewKademliaIDFromBytes(data)
    pinned := false
    _, err := KVSInsert(*id, pinned, data)
    if err != nil {
        fmt.Println(err)
        t.Fail()
    }
    _, err = KVSInsert(*id, pinned, data)
    if err != DuplicateError {
        fmt.Println("Wrong error")
        fmt.Println(err)
        t.Fail()
    }
    delete(KVStore, *id)
}
