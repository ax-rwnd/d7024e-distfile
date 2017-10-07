package kademlia

import (
    "testing"
    "fmt"
    "time"
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

func TestKVSNotFoundError(t *testing.T) {
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

func TestKVSTimeToLiveNoPin(t *testing.T) {
    // Store some non pinned data
    kvStore := NewKVStore()
    data1 := []byte("Test data1")
    data2 := []byte("Test data2")
    id1 := NewKademliaIDFromBytes(data1)
    id2 := NewKademliaIDFromBytes(data2)
    pinned := false

    // Set a short eviction time for testing
    EvictionTime = 3 * time.Second

    // Add some data
    fmt.Printf("Inserted %v\n", id1.String())
    kvStore.Insert(*id1, pinned, data1)
    // Wait two seconds, then check that data is still there
    timer := time.NewTimer(2 * time.Second)
    <-timer.C
    if _, err := kvStore.Lookup(*id1); err != nil {
        t.Fail()
        log.Println("ID1 was removed too early")
    }

    // Insert some more data before previous is evicted
    fmt.Printf("Inserted %v\n", id2.String())
    kvStore.Insert(*id2, pinned, data2)

    // Wait until ID1 should have been evicted
    // Check that ID1 was removed before ID2
    timer = time.NewTimer(2 * time.Second)
    <-timer.C
    if _, err := kvStore.Lookup(*id1); err == nil {
        t.Fail()
        log.Println("ID1 was not removed")
    }
    if _, err := kvStore.Lookup(*id2); err != nil {
        t.Fail()
        log.Println("ID2 was removed too early")
    }

    // Wait until both eviction times have passed, then check data is gone
    timer = time.NewTimer(2 * time.Second)
    <-timer.C
    if _, err := kvStore.Lookup(*id1); err == nil {
        t.Fail()
        log.Println("ID1 was not removed")
    }
    if _, err := kvStore.Lookup(*id2); err == nil {
        t.Fail()
        log.Println("ID2 was not removed")
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
