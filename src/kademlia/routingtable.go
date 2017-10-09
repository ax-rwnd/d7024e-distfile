package kademlia

import (
    "sync"
)

type RoutingTable struct {
    Me      Contact
    buckets [IDLength * 8]*bucket
    mutex   *sync.Mutex
}

func (routingTable *RoutingTable) GetBucket(index int) bucket {
	b := *routingTable.buckets[index]
	return b
}

func NewRoutingTable(me Contact) *RoutingTable {
    routingTable := &RoutingTable{}
    for i := 0; i < IDLength*8; i++ {
        routingTable.buckets[i] = newBucket() // 160 new buckets
    }
    routingTable.Me = me
    routingTable.mutex = &sync.Mutex{}
    return routingTable
}

func (routingTable *RoutingTable) AddContact(contact Contact, pingFunc func(*Contact) bool) (bool, *Contact) {
    routingTable.mutex.Lock()
    bucketIndex := routingTable.getBucketIndex(contact.ID) // contact we want to add, get what bucket we should insert the contact in
    bucket := routingTable.buckets[bucketIndex] // get the bucket from list of buckets in routingtable
    wasAdded := bucket.addContact(contact, pingFunc)
    routingTable.mutex.Unlock()
    return wasAdded, &contact
}

func (routingTable *RoutingTable) FindClosestContacts(target *KademliaID, count int) []Contact {
    routingTable.mutex.Lock()
    var candidates ContactCandidates
    bucketIndex := routingTable.getBucketIndex(target)
    bucket := routingTable.buckets[bucketIndex]

    candidates.Append(bucket.GetContactAndCalcDistance(target))

    for i := 1; (bucketIndex-i >= 0 || bucketIndex+i < IDLength*8) && candidates.Len() < count; i++ {
        if bucketIndex-i >= 0 {
            bucket = routingTable.buckets[bucketIndex-i]
            candidates.Append(bucket.GetContactAndCalcDistance(target))
        }
        if bucketIndex+i < IDLength*8 {
            bucket = routingTable.buckets[bucketIndex+i]
            candidates.Append(bucket.GetContactAndCalcDistance(target))
        }
    }
    routingTable.mutex.Unlock()

    candidates.Sort()
    if count > candidates.Len() {
        count = candidates.Len()
    }
    return candidates.GetContacts(count)
}

func (routingTable *RoutingTable) getBucketIndex(id *KademliaID) int {
    distance := id.CalcDistance(routingTable.Me.ID) 
    for i := 0; i < IDLength; i++ { // IDLength is 20
        for j := 0; j < 8; j++ { // Each ID index is one byte
            if (distance[i]>>uint8(7-j))&0x1 != 0 {
                return i*8 + j
            }
        }
    }

    return IDLength*8 - 1
}
