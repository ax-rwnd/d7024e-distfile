package kademlia

import (
    "fmt"
    "testing"
    "encoding/hex"
)

func TestRoutingTable(t *testing.T) {
    rt := NewRoutingTable(NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000000"), "localhost:8000"))

    rt.AddContact(NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000000"), "localhost:8001"))
    rt.AddContact(NewContact(NewKademliaID("1111111100000000000000000000000000000000"), "localhost:8002"))
    rt.AddContact(NewContact(NewKademliaID("1111111200000000000000000000000000000000"), "localhost:8003"))
    rt.AddContact(NewContact(NewKademliaID("1111111300000000000000000000000000000000"), "localhost:8004"))
    rt.AddContact(NewContact(NewKademliaID("1111111400000000000000000000000000000000"), "localhost:8005"))
    rt.AddContact(NewContact(NewKademliaID("2111111400000000000000000000000000000000"), "localhost:8006"))

    contacts := rt.FindClosestContacts(NewKademliaID("2111111400000000000000000000000000000000"), 20)
    for i := range contacts {
        fmt.Println(contacts[i].String())
    }
}

func TestRoutingTableFindClosestContacts(t *testing.T) {
    const numContacts = 0x42
    const portStart = 8000
    const maxResults = 20
    const toFind = 0x232323

    idToFind := NewKademliaID(fmt.Sprintf("%040x", toFind));
    var allcontacts [numContacts]Contact

    // Create a routing table with some contacts, also store all contacts for comparison
    me := NewContact(NewKademliaID(fmt.Sprintf("%040x", 0)), fmt.Sprintf("localhost:%04d", portStart))
    allcontacts[0] = me
    rt := NewRoutingTable(me)
    for i := 1; i < numContacts; i++ {
        contact := NewContact(NewKademliaID(fmt.Sprintf("%040x", i)), fmt.Sprintf("localhost:%04d", portStart+i))
        allcontacts[i] = contact
        rt.AddContact(contact)
    }
    // Find the closest contacts to 'toFind'. Contacts are sorted (test?) according to min distance
    contacts := rt.FindClosestContacts(idToFind, maxResults)
    if len(contacts) == 0 || len(contacts) > maxResults {
        t.Fail()
    }
    minDistance := contacts[0].distance
    // Check all contacts if there are any closer
    for i := range allcontacts {
        contact := allcontacts[i]
        // Use the algorithm from the paper
        distance := KademliaID{}
        for i := 0; i < IDLength; i++ {
            distance[i] = idToFind[i] ^ contact.ID[i]
        }
        for i := 0; i < IDLength; i++ {
            if idToFind[i] < minDistance[i] {
                t.Fail()
            }
        }
        // Try the API functions for it
        if idToFind.CalcDistance(contact.ID).Less(minDistance) {
            t.Fail()
        }
    }
}

func TestFileHash(t *testing.T) {
    decoded := "9ae65414b4803a999452f0c320eb41bec1e14bc1"
    fileHash, err := NewKademliaIDFromFile("test.bin")
    if err != nil {
        t.Fail()
    } else if hex.EncodeToString(fileHash[0:IDLength]) != decoded {
        t.Fail()
    }
}
