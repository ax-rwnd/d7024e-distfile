package kademlia

import (
    "testing"
    "fmt"
    "time"
    "log"
)

// Makes a grid/mesh of nodes and adds contacts for each node to 8 of its neighbours (fewer at borders).
func createKademliaMesh(width int, height int) []*Kademlia {
    k := make([]*Kademlia, width*height)
    // Loop over columns
    for y := 0; y < height; y++ {
        // Fill the row
        for x := 0; x < width; x++ {
            i := y*width + x
            k[i] = NewKademlia("127.0.0.1", getNetworkTestPort())
            // Connect along x axis
            if x > 0 {
                k[i-1].network.routing.AddContact(k[i].network.routing.me)
                k[i].network.routing.AddContact(k[i-1].network.routing.me)
            }
        }
        if y > 0 {
            for x := 0; x < width; x++ {
                // Connect between columns, down and diagonally
                me := y*width + x
                down := (y-1)*width + x
                downLeft := down - 1
                downRight := down + 1
                k[me].network.routing.AddContact(k[down].network.routing.me)
                k[down].network.routing.AddContact(k[me].network.routing.me)
                if x == 0 {
                    k[me].network.routing.AddContact(k[downRight].network.routing.me)
                    k[downRight].network.routing.AddContact(k[me].network.routing.me)
                } else if x == width-1 {
                    k[me].network.routing.AddContact(k[downLeft].network.routing.me)
                    k[downLeft].network.routing.AddContact(k[me].network.routing.me)
                } else {
                    k[me].network.routing.AddContact(k[downRight].network.routing.me)
                    k[downRight].network.routing.AddContact(k[me].network.routing.me)
                    k[me].network.routing.AddContact(k[downLeft].network.routing.me)
                    k[downLeft].network.routing.AddContact(k[me].network.routing.me)
                }
            }
        }
    }
    return k
}

// Test looking up a contact with specific kademlia ID
func TestLookupContact(t *testing.T) {
    kademlias := createKademliaMesh(10, 10)
    numNodes := len(kademlias)
    var cc = []chan []Contact{make(chan []Contact), make(chan []Contact),}
    // First node does not yet have last node as a contact. Find it.
    go func() {
        cc[0] <- kademlias[0].LookupContact(kademlias[numNodes-1].network.routing.me.ID)
    }()
    // Try the reverse concurrently
    go func() {
        cc[1] <- kademlias[numNodes-1].LookupContact(kademlias[0].network.routing.me.ID)
    }()
    contacts1 := <-cc[0]
    contacts2 := <-cc[1]
    fmt.Printf("%v lookup %v found %v\n", kademlias[0].network.routing.me.Address, kademlias[numNodes-1].network.routing.me.ID.String(), contacts1)
    fmt.Printf("%v lookup %v found %v\n", kademlias[numNodes-1].network.routing.me.Address, kademlias[0].network.routing.me.ID.String(), contacts2)
    // Check that the contacts are correct
    if !contacts1[0].ID.Equals(kademlias[numNodes-1].network.routing.me.ID) {
        t.Fail()
    }
    if !contacts2[0].ID.Equals(kademlias[0].network.routing.me.ID) {
        t.Fail()
    }
}

// Test storing and finding data
func TestLookupStoreData(t *testing.T) {
    kademlias := createKademliaMesh(5, 10)
    numK := len(kademlias)
    // Store some data
    owner1 := kademlias[0]
    data := []byte("message")
    owner1.Store(data)
    // Wait for the messages to propagate
    timer := time.NewTimer(time.Second * 2)
    <-timer.C
    // Read data from another node
    hash := NewKademliaIDFromBytes(data)
    reader := kademlias[numK-1]
    candidates := *reader.LookupData(hash)
    // Check that we actually got the right contact
    fmt.Printf("Found owners %v\n", candidates)
    if !candidates[0].ID.Equals(owner1.network.routing.me.ID) {
        t.Fail()
        log.Printf("Invalid contact list %v\n", candidates)
    }
    // TODO: Actually transfer the data, not just owner contact
}

// Test storing and finding data
func TestLookupStoreDataMultiple(t *testing.T) {
    kademlias := createKademliaMesh(3, 3)
    numK := len(kademlias)
    // Store some data
    owner1 := kademlias[0]
    owner2 := kademlias[1]
    data := []byte("message")
    owner1.Store(data)
    owner2.Store(data)
    // Wait for the messages to propagate
    timer := time.NewTimer(time.Second)
    <-timer.C
    // Read data from another node
    hash := NewKademliaIDFromBytes(data)
    reader := kademlias[numK-1]
    candidates := *reader.LookupData(hash)
    // Check that we actually got the right contact
    fmt.Printf("Found owners %v\n", candidates)
    if !(candidates[0].ID.Equals(owner1.network.routing.me.ID) && candidates[1].ID.Equals(owner2.network.routing.me.ID) ||
        candidates[1].ID.Equals(owner1.network.routing.me.ID) && candidates[0].ID.Equals(owner2.network.routing.me.ID)) {
        t.Fail()
        log.Printf("Invalid contact list %v\n", candidates)
    }
    // TODO: Actually transfer the data, not just owner contact
}
