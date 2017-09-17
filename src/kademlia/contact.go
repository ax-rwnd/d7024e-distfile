package kademlia

import (
    "fmt"
    "sort"
)

type Address struct {
    IP      string
    TcpPort int
    UdpPort int
}

type Contact struct {
    ID       *KademliaID
    Address  Address
    distance *KademliaID
}

func NewContact(id *KademliaID, ip string, tcpPort int, udpPort int) Contact {
    return Contact{id, Address{IP: ip, TcpPort: tcpPort, UdpPort: udpPort}, nil}
}

func (contact *Contact) CalcDistance(target *KademliaID) {
    contact.distance = contact.ID.CalcDistance(target)
}

func (contact *Contact) Less(otherContact *Contact) bool {
    return contact.distance.Less(otherContact.distance)
}

func (contact *Contact) String() string {
    return fmt.Sprintf(`contact(ID=%v, IP=%v, tcpPort=%v, udpPort=%v)`, contact.ID, contact.Address.IP, contact.Address.TcpPort, contact.Address.UdpPort)
}

type ContactCandidates struct {
    contacts []Contact
}

func (candidates *ContactCandidates) Append(contacts []Contact) {
    candidates.contacts = append(candidates.contacts, contacts...)
}

func (candidates *ContactCandidates) GetContacts(count int) []Contact {
    return candidates.contacts[:count]
}

func (candidates *ContactCandidates) Sort() {
    sort.Sort(candidates)
}

func (candidates *ContactCandidates) Len() int {
    return len(candidates.contacts)
}

func (candidates *ContactCandidates) Swap(i, j int) {
    candidates.contacts[i], candidates.contacts[j] = candidates.contacts[j], candidates.contacts[i]
}

func (candidates *ContactCandidates) Less(i, j int) bool {
    return candidates.contacts[i].Less(&candidates.contacts[j])
}

func (contact *Contact) Equals(target *Contact) bool {
    return contact.ID.Equals(target.ID) && contact.Address == target.Address
}
