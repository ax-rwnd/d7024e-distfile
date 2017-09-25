package kademlia

import (
    "fmt"
    "reflect"
    "sync"
)

const ALPHA = 3
const K = bucketSize

type Kademlia struct {
    Net *Network
}

func NewKademlia(ip string, tcpPort int, udpPort int) *Kademlia {
    kademlia := new(Kademlia)
    kademlia.Net = NewNetwork(ip, tcpPort, udpPort)
    return kademlia
}

// Lookup the k participants which have a kademlia ID closest to another ID
func (kademlia *Kademlia) LookupContact(target *KademliaID) ([]Contact) {
    me := kademlia.Net.Routing.Me
    // The lookup intiator starts by picking \alpha nodes from its closest non-empty k-bucket...
    closestContacts := kademlia.Net.Routing.FindClosestContacts(target, ALPHA)
    // This holds the nodes we have already queried
    contactsVisited := make(map[KademliaID]Contact)
    contactsVisited[*me.ID] = me
    for _, contact := range closestContacts {
        contactsVisited[*contact.ID] = contact
    }
    var mutex = &sync.Mutex{}
    var lookup func(closestContacts []Contact) []Contact
    lookup = func(initialContacts []Contact) []Contact {
        // Channels for sending/receiving network messages
        rpcChannels := []chan []Contact{}
        for i := 0; i < len(initialContacts); i++ {
            rpcChannels = append(rpcChannels, make(chan []Contact))
        }
        // The initiator then sends parallel, asynchronous FIND_NODE RPCs to the \alpha nodes it has chosen...
        // ... In the recursive step, the initiator resends the FIND_NODE to nodes that it has learned about
        // from its previous RPCs
        go func(findTarget KademliaID, receivers []Contact, channels []chan []Contact) {
            for i := range initialContacts {
                reflect.Select(
                    []reflect.SelectCase{{
                        Dir:  reflect.SelectSend,
                        Chan: reflect.ValueOf(channels[i]),
                        Send: reflect.ValueOf(kademlia.Net.SendFindContactMessage(&findTarget, &receivers[i])),
                    }})
            }
        }(*target, initialContacts, rpcChannels)

        // This will gather all the contacts from recursive calls
        newContacts := []Contact{}
        // Channels to recursive lookup calls
        lookupChannels := []chan []Contact{}
        // There are as many RPC channels as query contacts
        for i := 0; i < len(initialContacts); i++ {
            // Block until we get one or more responses from FIND_NODE RPCs
            set := []reflect.SelectCase{}
            for _, ch := range rpcChannels {
                set = append(set, reflect.SelectCase{
                    Dir:  reflect.SelectRecv,
                    Chan: reflect.ValueOf(ch),
                })
            }
            _, valValue, _ := reflect.Select(set)
            receivedContacts := valValue.Interface().([]Contact)

            contactsToVisit := []Contact{}
            // Check if we have already visited these contacts. If not, queue them for future visits.
            // Simpler version than the rules in the paper
            for i := 0; i < min(K, len(receivedContacts)); i++ {
                newContact := receivedContacts[i]
                mutex.Lock()
                if _, ok := contactsVisited[*newContact.ID]; !ok {
                    newContacts = append(newContacts, newContact)
                    contactsToVisit = append(contactsToVisit, newContact)
                    contactsVisited[*newContact.ID] = newContact
                    fmt.Printf("%v new contact: %v\n", me.Address, newContact.String())
                }
                mutex.Unlock()
            }
            // If there were any new nodes, visit them now
            if len(contactsToVisit) > 0 {
                callChannel := make(chan []Contact)
                lookupChannels = append(lookupChannels, callChannel)
                // Make new recursive lookup calls and Store the channels
                go func(input []Contact, callChannel chan []Contact) int {
                    set := []reflect.SelectCase{}
                    set = append(set, reflect.SelectCase{
                        Dir:  reflect.SelectSend,
                        Chan: reflect.ValueOf(callChannel),
                        Send: reflect.ValueOf(lookup(input)),
                    })
                    to, _, _ := reflect.Select(set)
                    return to
                }(contactsToVisit, callChannel)
            }
        }
        // Gather results from all the recursive calls we made and return them
        for i := 0; i < len(lookupChannels); i++ {
            set := []reflect.SelectCase{}
            for _, ch := range lookupChannels {
                set = append(set, reflect.SelectCase{
                    Dir:  reflect.SelectRecv,
                    Chan: reflect.ValueOf(ch),
                })
            }
            _, valValue, _ := reflect.Select(set)
            newContacts = append(newContacts, valValue.Interface().([]Contact)...)
        }
        return newContacts
    }
    // Block on the initial call to the recursive lookup
    candidates := lookup(closestContacts)
    candidates = append(candidates, closestContacts...)
    // Array will contain the closest contacts found during lookup
    fmt.Printf("%v search for %v found %v candidates\n", me.Address, target.String(), len(candidates))
    var contactCandidates ContactCandidates
    for _, candidate := range candidates {
        candidate.CalcDistance(target)
        contactCandidates.contacts = append(contactCandidates.contacts, candidate)
    }
    contactCandidates.Sort()
    return contactCandidates.contacts[0:min(K, len(candidates))]
}

// Find the owner of a file with specific hash.
func (kademlia *Kademlia) LookupData(hash *KademliaID) *[]Contact {
    // First find the contacts of the nodes with closest ID to hash
    closestContacts := kademlia.LookupContact(hash)
    rpcChannels := []chan []Contact{}
    for i := 0; i < len(closestContacts); i++ {
        rpcChannels = append(rpcChannels, make(chan []Contact))
    }
    for i := range closestContacts {
        // Send concurrent find data requests RPCs
        go func(findTarget *KademliaID, receiver *Contact, channel chan []Contact) int {
            set := []reflect.SelectCase{{
                Dir:  reflect.SelectSend,
                Chan: reflect.ValueOf(channel),
                Send: reflect.ValueOf(kademlia.Net.SendFindDataMessage(hash, receiver)),
            }}
            to, _, _ := reflect.Select(set)
            return to
        }(hash, &closestContacts[i], rpcChannels[i])
    }
    ownerMap := make(map[KademliaID]Contact)

    for range closestContacts {
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
        if newContacts != nil && len(newContacts) > 0 {
            // Return all of the owners if there are many?
            for _, newContact := range newContacts {
                if _, ok := ownerMap[*newContact.ID]; !ok {
                    ownerMap[*newContact.ID] = newContact
                }
            }
        }
    }
    owners := []Contact{}
    for _, value := range ownerMap {
        owners = append(owners, value)
    }
    return &owners
}

// Store the data locally, then have other nodes Store the contact of ones holding the data
func (kademlia *Kademlia) Store(data []byte) KademliaID {
    hash := NewKademliaIDFromBytes(data)
    kademlia.Net.Store.Insert(*hash, false, data)
    contacts := kademlia.LookupContact(hash)
    for _, contact := range contacts {
        go kademlia.Net.SendStoreMessage(hash, &contact)
    }
    return *hash
}

// Download data from another kademlia participant
func (kademlia *Kademlia) Download(hash *KademliaID, from *Contact) []byte {
    return kademlia.Net.SendDownloadMessage(hash, from)
}
