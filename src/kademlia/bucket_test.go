package kademlia

import (
    "fmt"
    "testing"
)

func TestNewBucket(t *testing.T) {
    element := newBucket()
    if element.list == nil {
        t.Fail()
    }
}

func TestAddContact(t *testing.T) {
    storage := newBucket()
    a := NewContact(NewKademliaID("0000000000000000000000000000000000000000"), "localhost", 0, 0)
    b := NewContact(NewKademliaID("0000000000000000000000000000000000000001"), "localhost", 0, 0)
    c := NewContact(NewKademliaID("0000000000000000000000000000000000000011"), "localhost", 0, 0)
    d := NewContact(NewKademliaID("0000000000000000000000000000000000000111"), "localhost", 0, 0)

    storage.addContact(a, nil)
    storage.addContact(b, nil)
    storage.addContact(c, nil)
    storage.addContact(d, nil)

    if storage.list.Remove(storage.list.Back()) != a ||
        storage.list.Remove(storage.list.Back()) != b ||
        storage.list.Remove(storage.list.Back()) != c ||
        storage.list.Remove(storage.list.Back()) != d {
        t.Fail()
    }
}

func TestGetContactAndCalcDistance(t *testing.T) {
    storage := newBucket()
    a := NewContact(NewKademliaID("1010101010101010101010101010101010101010"), "localhost", 0, 0)
    b := NewContact(NewKademliaID("0101010101010101010101010101010101010101"), "localhost", 0, 0)
    e := NewContact(NewKademliaID("0000000000000000000000000000000000000000"), "localhost", 0, 0)

    storage.addContact(a, nil)
    storage.addContact(b, nil)

    testVals := storage.GetContactAndCalcDistance(e.ID)
    if testVals[1].ID != a.ID {
        fmt.Println(testVals[1].ID)
        fmt.Println(a.ID)
        t.Fail()
    }

    if testVals[0].ID != b.ID {
        fmt.Println(testVals[0].ID)
        fmt.Println(b.ID)
        t.Fail()
    }
}

func TestLen(t *testing.T) {
    storage := newBucket()

    for i := 0; i < 20; i++ {
        storage.addContact(NewContact(NewKademliaID("0000000000000000000000000000000000000000"), "localhost", 0, 0), nil)
    }

    if storage.Len() != 1 {
        fmt.Println("Wrong length", storage.Len(), " returned, expected 1")
        t.Fail()
    }

    storage = newBucket()
    storage.addContact(NewContact(NewKademliaID("0000000000000000000000000000000000000000"), "localhost", 0, 0), nil)
    storage.addContact(NewContact(NewKademliaID("1000000000000000000000000000000000000000"), "localhost", 0, 0), nil)
    storage.addContact(NewContact(NewKademliaID("2000000000000000000000000000000000000000"), "localhost", 0, 0), nil)
    storage.addContact(NewContact(NewKademliaID("3000000000000000000000000000000000000000"), "localhost", 0, 0), nil)
    storage.addContact(NewContact(NewKademliaID("4000000000000000000000000000000000000000"), "localhost", 0, 0), nil)

    if storage.Len() != 5 {
        fmt.Println("Wrong length", storage.Len(), " returned, expected 5")
        t.Fail()
    }

}

func TestDumpContacts(t *testing.T) {
    b := newBucket()
    con1 := NewContact(NewKademliaID("0000000000000000000000000000000000000000"), "123.123.123.123", 1234, 4321)
    con2 := NewContact(NewKademliaID("1010101010101010101010101010101010101010"), "456.456.456.456", 5678, 8765)

    b.addContact(con1, nil)
    b.addContact(con2, nil)

    c_out := b.DumpContacts()
    if c_out[0] != con1 && c_out[1] != con1  {
        t.Fail()
    }
    if c_out[0] != con2 && c_out[1] != con2  {
        t.Fail()
    }
}
