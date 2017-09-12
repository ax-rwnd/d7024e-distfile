package kademlia

import (
    "testing"
    "fmt"
)

func TestLookupContact(t *testing.T) {
    // Create nodes array and double link contact information between them.
    // This means for more nodes than bucketSize+1 this test will fail
    const numNodes = bucketSize + 1
    kademlias := []*Kademlia{}
    for i := 0; i < numNodes; i++ {
        kademlias = append(kademlias, NewKademlia("127.0.0.1", getNetworkTestPort()))
    }
    fmt.Printf("looking up %v\n", kademlias[numNodes-1].network.routing.me.String())
    // Add some contacts between them
    for i := range kademlias {
        if i == 0 {
            kademlias[0].network.routing.AddContact(kademlias[1].network.routing.me)
        } else if i == numNodes-1 {
            kademlias[numNodes-1].network.routing.AddContact(kademlias[numNodes-2].network.routing.me)
        } else {
            kademlias[i].network.routing.AddContact(kademlias[i-1].network.routing.me)
            kademlias[i].network.routing.AddContact(kademlias[i+1].network.routing.me)
        }
    }
    var cc = []chan []Contact{make(chan []Contact), make(chan []Contact),}
    // First node does not yet have last node as a contact. Find it.
    go func() {
        cc[0] <- kademlias[0].LookupContact(&kademlias[numNodes-1].network.routing.me)
    }()
    // Try the reverse concurrently
    go func() {
        cc[1] <- kademlias[numNodes-1].LookupContact(&kademlias[0].network.routing.me)
    }()
    contacts1 := <-cc[0]
    contacts2 := <-cc[1]
    fmt.Printf("%v lookup %v found %v\n", kademlias[0].network.routing.me.Address, kademlias[numNodes-1].network.routing.me.ID.String(), contacts1)
    fmt.Printf("%v lookup %v found %v\n", kademlias[numNodes-1].network.routing.me.Address, kademlias[0].network.routing.me.ID.String(), contacts2)
    if !contacts1[0].ID.Equals(kademlias[numNodes-1].network.routing.me.ID) {
        t.Fail()
    }
    if !contacts2[0].ID.Equals(kademlias[0].network.routing.me.ID) {
        t.Fail()
    }
}
