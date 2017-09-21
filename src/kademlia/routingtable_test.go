package kademlia

import (
    "fmt"
    "testing"
)

func TestRoutingTableFindClosestContacts(t *testing.T) {
    const numContacts = 0x42
    const maxResults = 20
    const toFind = 0x232323

    idToFind := NewKademliaID(fmt.Sprintf("%040x", toFind))
    var allcontacts [numContacts]Contact

    // Create a routing table with some contacts, also store all contacts for comparison
    me := NewContact(NewKademliaID(fmt.Sprintf("%040x", 0)), "127.0.0.1", 0, 0)
    allcontacts[0] = me
    rt := NewRoutingTable(me)
    for i := 1; i < numContacts; i++ {
        contact := NewContact(NewKademliaID(fmt.Sprintf("%040x", i)), "127.0.0.1", 0, 0)
        allcontacts[i] = contact
        rt.AddContact(contact, nil)
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
