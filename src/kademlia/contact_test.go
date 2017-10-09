package kademlia

import (
    "log"
    "testing"
)

func TestNewContactAndEqual(t *testing.T) {
    contact := NewContact(NewKademliaID("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF"), "test", 0, 0)
    if !contact.ID.Equals(NewKademliaID("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF")) {
        log.Println("Wrong ID!")
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

func TestContactLess(t *testing.T) {
    defer func() {
        if r := recover(); r != nil {
            log.Println("Recovering from panic:", r)
            t.Fail()
        }
    }()

    origin := NewContact(NewKademliaID("0000000000000000000000000000000000000000"), "test", 0, 0)
    a := NewContact(NewKademliaID("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF"), "test", 0, 0)
    b := NewContact(NewKademliaID("FFFFFFFFFFFFFFFFFF0000000000000000000000"), "test", 0, 0)
    a.CalcDistance(origin.ID)
    b.CalcDistance(origin.ID)

    if a.Less(&b) {
        log.Println("a should not be less than b")
        t.Fail()
    }

    origin.CalcDistance(origin.ID)
    if origin.Less(&origin) {
        log.Println("less returns less on equal")
        t.Fail()
    }
}

func TestContactCalcDistance(t *testing.T) {
    origin := NewContact(NewKademliaID("0000000000000000000000000000000000000000"), "test", 0, 0)
    c := NewContact(NewKademliaID("1010101010101010101010101010101010101010"), "test", 0, 0)

    c.CalcDistance(origin.ID)
    if !c.distance.Equals(c.ID) {
        log.Println("Distance does not represent distance from origo.")
        t.Fail()
    }

    origin.CalcDistance(origin.ID)
    if !origin.distance.Equals(origin.ID) {
        log.Println("Origin is not the zero-vector", origin.distance, origin.ID)
        t.Fail()
    }
}

func TestContactCandiatesAppend(t *testing.T) {
    a := NewContact(NewKademliaID("0000000000000000000000000000000000000000"), "test", 0, 0)
    b := NewContact(NewKademliaID("1111111111111111111111111111111111111111"), "test", 0, 0)

    lsta := ContactCandidates{[]Contact{a, a, a, a}}
    lstb := ContactCandidates{[]Contact{b, b, b, b}}
    lsta.Append(lstb.contacts)

    lstr := ContactCandidates{[]Contact{a, a, a, a, b, b, b, b}}

    if lsta.Len() != lstr.Len() {
        log.Println("Wrong length of list")
        t.Fail()
    }
}

func TestSort(t *testing.T) {

    id1 := NewKademliaID("FFFFFFFF00000000000000000000000000000000")
    id2 := NewKademliaID("FFFFFFFF00000000000000000000000000000001")
    id3 := NewKademliaID("0000000000000000000000000000000000000002")
    id4 := NewKademliaID("FFFF000000000000000000000000000000000003")
    id5 := NewKademliaID("FFFFFFFF00000000000000000000000000000004")
    contact1 := NewContact(id1, "localhost", 0, 0)
    contact2 := NewContact(id2, "localhost", 0, 0)
    contact3 := NewContact(id3, "localhost", 0, 0)
    contact4 := NewContact(id4, "localhost", 0, 0)
    contact5 := NewContact(id5, "localhost", 0, 0)

    cons := []Contact{contact1, contact2, contact3, contact4, contact5}
    cand := ContactCandidates{contacts: cons}

    contact := NewContact(NewKademliaID("0000000000000000000000000000000000000000"), "localhost", 0, 0)
    for i, cont := range cand.contacts {
        cont.CalcDistance(contact.ID)
        cand.contacts[i] = cont
    }
    cand.Sort()
    for i := 0; i < (len(cand.contacts) - 1); i++ {
        con := cand.contacts[i]
        b := con.Less(&cand.contacts[i+1])
        if !b {
            t.Fail()
        }
    }
}
