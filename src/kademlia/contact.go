package kademlia

import (
    "fmt"
    "sort"
    "strings"
    "strconv"
    "errors"
)

type Contact struct {
    ID       *KademliaID
    Address  string
    distance *KademliaID
}

func NewContact(id *KademliaID, address string) Contact {
    return Contact{id, address, nil}
}

func (contact *Contact) CalcDistance(target *KademliaID) {
    contact.distance = contact.ID.CalcDistance(target)
}

func (contact *Contact) Less(otherContact *Contact) bool {
    return contact.distance.Less(otherContact.distance)
}

func (contact *Contact) String() string {
    return fmt.Sprintf(`contact("%s", "%s")`, contact.ID, contact.Address)
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

func (contact *Contact) ParseAddress() (ip string, port int, err error) {
    err = errors.New(fmt.Sprintf("%v Error parsing address\n", contact.Address))
    split := strings.Split(contact.Address, ":")
    if len(split) == 2 {
        var p int64
        p, err = strconv.ParseInt(split[1], 10, 32)
        if err == nil {
            ip = split[0]
            port = int(p)
        }
    }
    return ip, port, err
}
