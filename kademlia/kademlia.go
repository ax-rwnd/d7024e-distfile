package kademlia

import (
    "fmt"
    "reflect"
)

type Kademlia struct {
    network *Network
}

func NewKademlia(ip string, port int) *Kademlia {
    kademlia := new(Kademlia)
    kademlia.network = NewNetwork(ip, port)
    return kademlia
}

func (kademlia *Kademlia) LookupContact(target *Contact) ([]Contact) {
    me := kademlia.network.routing.me;
    myAddress := kademlia.network.myAddress

    // Find the alpha closest nodes
    closestContacts := kademlia.network.routing.FindClosestContacts(target.ID, ALPHA)
    // How many nodes we have queried so far
    numNodesVisited := 0
    // This holds the nodes we have already queried
    nodesVisited := NewRoutingTable(me)
    for _, contact := range closestContacts {
        nodesVisited.AddContact(contact)
        numNodesVisited++
    }
    // Mutex http://www.golangpatterns.info/concurrency/semaphores
    mut := make(chan struct{}, 1)

    var lookup func(closestContacts []Contact) []Contact
    lookup = func(contactsToQuery []Contact) []Contact {
        // Channels for sending/receiving network messages
        rpcChannels := []chan []Contact{}
        for i := 0; i < len(contactsToQuery); i++ {
            rpcChannels = append(rpcChannels, make(chan []Contact))
        }
        // Send go routine RPCs to the closest contacts, connect to channels
        for i := range contactsToQuery {
            go func(findTarget *Contact, receiver *Contact, channel chan []Contact) int {
                set := []reflect.SelectCase{reflect.SelectCase{
                    Dir:  reflect.SelectSend,
                    Chan: reflect.ValueOf(channel),
                    Send: reflect.ValueOf(kademlia.network.SendFindContactMessage(findTarget, receiver)),
                }}
                to, _, _ := reflect.Select(set)
                return to
            }(target, &contactsToQuery[i], rpcChannels[i])
        }
        // Channels to recursive lookup calls
        lookupChannels := []chan []Contact{}
        // There are as many RPC channels as closest contacts
        for i := 0; i < len(contactsToQuery); i++ {
            // Block until we get one or more responses from RPCs
            set := []reflect.SelectCase{}
            for _, ch := range rpcChannels {
                set = append(set, reflect.SelectCase{
                    Dir:  reflect.SelectRecv,
                    Chan: reflect.ValueOf(ch),
                })
            }
            _, valValue, _ := reflect.Select(set)
            newContacts := valValue.Interface().([]Contact)

            nodesToVisit := []Contact{}
            // Check if we have already visited these contacts. If not, queue them for future visits.
            for i := 0; i < min(ALPHA, len(newContacts)); i++ {
                newContact := newContacts[i]
                // Mutex lock here to synchronize shared variables
                mut <- struct{}{}
                if !nodesVisited.Contains(newContact) && !me.ID.Equals(newContact.ID) && numNodesVisited < bucketSize {
                    nodesToVisit = append(nodesToVisit, newContact)
                    nodesVisited.AddContact(newContact)
                    numNodesVisited++
                    fmt.Printf("%v new contact: %v\n", myAddress, newContact.String())
                }
                <-mut
            }
            // If there were any new nodes, visit them now
            if len(nodesToVisit) > 0 {
                callChannel := make(chan []Contact)
                lookupChannels = append(lookupChannels, callChannel)

                // Make new recursive lookup calls and store the channels
                go func(input []Contact, ch chan []Contact) int {
                    set := []reflect.SelectCase{}
                    set = append(set, reflect.SelectCase{
                        Dir:  reflect.SelectSend,
                        Chan: reflect.ValueOf(ch),
                        Send: reflect.ValueOf(lookup(input)),
                    })
                    to, _, _ := reflect.Select(set)
                    return to
                }(nodesToVisit, callChannel)
            }
        }

        // Gather results from all the recursive calls we made and return them
        allContacts := []Contact{}
        for i := 0; i < len(lookupChannels); i++ {
            set := []reflect.SelectCase{}
            for _, ch := range lookupChannels {
                set = append(set, reflect.SelectCase{
                    Dir:  reflect.SelectRecv,
                    Chan: reflect.ValueOf(ch),
                })
            }
            _, valValue, _ := reflect.Select(set)
            for _, newContact := range valValue.Interface().([]Contact) {
                allContacts = append(allContacts, newContact)
            }
        }
        return allContacts
    }

    // Block on the initial call to the recursive lookup
    lookup(closestContacts)
    // The temporary routing table will contain the closest contacts found during lookup
    return nodesVisited.FindClosestContacts(target.ID, bucketSize)
}

func (kademlia *Kademlia) LookupData(hash string) {
    // TODO
}

func (kademlia *Kademlia) Store(data []byte) {
    // TODO
}
