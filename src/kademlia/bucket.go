package kademlia

import (
    "container/list"
)

type bucket struct {
    list *list.List
}

func newBucket() *bucket {
    bucket := &bucket{}
    bucket.list = list.New()
    return bucket
}

func (bucket *bucket) addContact(contact Contact, pingFunc func(*Contact) bool) bool {
    var element *list.Element
    for e := bucket.list.Front(); e != nil; e = e.Next() {
        nodeID := e.Value.(Contact).ID
        if (contact).ID.Equals(nodeID) {
            element = e
        }
    }
    if element == nil {
        if bucket.list.Len() < ReplicationFactor {
            bucket.list.PushFront(contact)
        } else if pingFunc != nil {
            last := bucket.list.Back().Value.(Contact)
            responded := pingFunc(&last)
            if responded {
                bucket.list.MoveToFront(bucket.list.Back())
                // Could not add contact, bucket was full and last responded to ping
                return false
            } else {
                // Remove the last contact since it did not respond to ping
                bucket.list.Remove(bucket.list.Back())
                bucket.list.PushFront(contact)
            }
        }
    } else {
        bucket.list.MoveToFront(element)
    }
    return true
}

func (bucket *bucket) GetContactAndCalcDistance(target *KademliaID) []Contact {
    var contacts []Contact

    for elt := bucket.list.Front(); elt != nil; elt = elt.Next() {
        contact := elt.Value.(Contact)
        contact.CalcDistance(target)
        contacts = append(contacts, contact)
    }

    return contacts
}

func (bucket *bucket) Len() int {
    return bucket.list.Len()
}

func (bucket *bucket) DumpContacts() []Contact {
    blist := make([]Contact, bucket.Len())

    i := 0
    for e := bucket.list.Front(); e != nil; e = e.Next() {
        blist[i] = e.Value.(Contact)
        i++
    }

    return blist
}
