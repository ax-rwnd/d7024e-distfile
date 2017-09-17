package kademlia

import (
    "fmt"
    "testing"
)

func TestNewContactAndEqual(t *testing.T) {
    contact := NewContact(NewKademliaID("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF"), "test", 0, 0)
    if !contact.ID.Equals(NewKademliaID("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF")) {
        fmt.Println("Wrong ID!")
        t.Fail()
    }
}

func TestAppendAndGetCandidates(t *testing.T) {
    a := NewContact(NewKademliaID("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF"), "test", 0, 0)
    b := NewContact(NewKademliaID("FFFFFFFFFFFFFFFFFF0000000000000000000000"), "test", 0, 0)
    cands := ContactCandidates{[]Contact{a, b}}

    c := NewContact(NewKademliaID("00000000000000000000FFFFFFFFFFFFFFFFFFFF"), "test", 0, 0)
    d := NewContact(NewKademliaID("0DEADBEEFFFFFFFFFF0000000000000000000000"), "test", 0, 0)
    candsb := ContactCandidates{[]Contact{c, d}}

    cands.Append(candsb.GetContacts(2))
    clist := cands.GetContacts(4)
    cref := []Contact{a, b, c, d}
    var found bool

    for i := 0; i < 4; i++ {
        found = false

        for j := 0; j < 4; j++ {
            if cref[j] == clist[j] {
                found = true
            }
        }
        if found == false {
            t.Fail()
        }
    }
}
