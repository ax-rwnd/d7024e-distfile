package kademlia

import (
    "testing"
    "fmt"
    "io/ioutil"
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
            k[i] = NewKademlia("127.0.0.1", getTestPort(), getTestPort())
            // Connect along x axis
            if x > 0 {
                k[i-1].Net.Routing.AddContact(k[i].Net.Routing.Me, nil)
                k[i].Net.Routing.AddContact(k[i-1].Net.Routing.Me, nil)
            }
        }
        if y > 0 {
            for x := 0; x < width; x++ {
                // Connect between columns, down and diagonally
                me := y*width + x
                down := (y-1)*width + x
                downLeft := down - 1
                downRight := down + 1
                k[me].Net.Routing.AddContact(k[down].Net.Routing.Me, nil)
                k[down].Net.Routing.AddContact(k[me].Net.Routing.Me, nil)
                if x == 0 {
                    k[me].Net.Routing.AddContact(k[downRight].Net.Routing.Me, nil)
                    k[downRight].Net.Routing.AddContact(k[me].Net.Routing.Me, nil)
                } else if x == width-1 {
                    k[me].Net.Routing.AddContact(k[downLeft].Net.Routing.Me, nil)
                    k[downLeft].Net.Routing.AddContact(k[me].Net.Routing.Me, nil)
                } else {
                    k[me].Net.Routing.AddContact(k[downRight].Net.Routing.Me, nil)
                    k[downRight].Net.Routing.AddContact(k[me].Net.Routing.Me, nil)
                    k[me].Net.Routing.AddContact(k[downLeft].Net.Routing.Me, nil)
                    k[downLeft].Net.Routing.AddContact(k[me].Net.Routing.Me, nil)
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
        cc[0] <- kademlias[0].LookupContact(kademlias[numNodes-1].Net.Routing.Me.ID)
    }()
    // Try the reverse concurrently
    go func() {
        cc[1] <- kademlias[numNodes-1].LookupContact(kademlias[0].Net.Routing.Me.ID)
    }()
    contacts1 := <-cc[0]
    contacts2 := <-cc[1]
    fmt.Printf("%v lookup %v found %v\n", kademlias[0].Net.Routing.Me.Address, kademlias[numNodes-1].Net.Routing.Me.ID.String(), contacts1)
    fmt.Printf("%v lookup %v found %v\n", kademlias[numNodes-1].Net.Routing.Me.Address, kademlias[0].Net.Routing.Me.ID.String(), contacts2)
    // Check that the contacts are correct
    if !contacts1[0].ID.Equals(kademlias[numNodes-1].Net.Routing.Me.ID) {
        t.Fail()
    }
    if !contacts2[0].ID.Equals(kademlias[0].Net.Routing.Me.ID) {
        t.Fail()
    }
}

//
// Test storing and finding data
func TestLookupStoreData(t *testing.T) {
    data, _ := ioutil.ReadFile("test.bin")
    // Create some network nodes
    kademlias := createKademliaMesh(5, 10)
    owner := kademlias[0]
    reader := kademlias[len(kademlias)-1]
    // Store some data
    owner.Store(data)
    // Wait for the messages to propagate
    timer := time.NewTimer(time.Second * 2)
    <-timer.C
    // Read data from another node
    hash := NewKademliaIDFromBytes(data)
    candidates := *reader.LookupData(hash)
    fmt.Printf("Found owners %v\n", candidates)
    downloadedData := reader.Download(hash, &candidates[0])

    // Check that we actually got the right contact
    if len(candidates) != 1 || !candidates[0].ID.Equals(owner.Net.Routing.Me.ID) {
        t.Fail()
        log.Printf("Invalid contact list %v\n", candidates)
    }
    // Check if download worked
    if len(downloadedData) != len(data) {
        t.Fail()
    }
    for i := range data {
        if data[i] != downloadedData[i] {
            t.Fail()
        }
    }
}

// Test storing data on multiple nodes and finding it from another
func TestLookupStoreDataMultiple(t *testing.T) {
    // Load a text file to store on network
    data, _ := ioutil.ReadFile("test.txt")
    hash := NewKademliaIDFromBytes(data)
    // Create some network nodes
    kademlias := createKademliaMesh(10, 10)
    owner1 := kademlias[0]
    owner2 := kademlias[1]
    requester := kademlias[len(kademlias)-1]
    // Store the data
    fmt.Printf("Storing on %v and %v: key=%v, value=%v\n", owner1.Net.Routing.Me.String(), owner2.Net.Routing.Me.String(), hash.String(), string(data))
    owner1.Store(data)
    owner2.Store(data)
    // Wait for the messages to propagate
    timer := time.NewTimer(time.Second * 1)
    <-timer.C
    // Read data from another node
    candidates := *requester.LookupData(hash)
    fmt.Printf("Found candidates %v\n", candidates)
    // Download data and check if program state is correct
    downloadedData1 := requester.Download(hash, &candidates[0])
    downloadedData2 := requester.Download(hash, &candidates[1])

    // Check that we actually got the right contact
    if !(candidates[0].ID.Equals(owner1.Net.Routing.Me.ID) && candidates[1].ID.Equals(owner2.Net.Routing.Me.ID) ||
        candidates[1].ID.Equals(owner1.Net.Routing.Me.ID) && candidates[0].ID.Equals(owner2.Net.Routing.Me.ID)) {
        t.Fail()
    }
    // Check that only the correct two owner nodes made it into candidates
    if len(candidates) != 2 {
        t.Fail()
    }
    // Data must have correct length
    if len(downloadedData1) != len(data) || len(downloadedData2) != len(data) {
        t.Fail()
    }
    // Bits must be correct
    for i := range data {
        if data[i] != downloadedData1[i] || data[i] != downloadedData2[i] {
            t.Fail()
        }
    }
    fmt.Printf("Downloaded from %v and %v: %v\n", candidates[0].String(), candidates[1].String(), string(downloadedData1))
}
