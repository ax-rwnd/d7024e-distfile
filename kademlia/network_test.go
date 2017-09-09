package kademlia

import (
    "fmt"
    "testing"
)

func TestUDP(t *testing.T) {
    var node1 *Network
    var node2 *Network

    node1chan := make(chan int)
    node2chan := make(chan int)

    go func() {
        fmt.Println("starting node 1")
        node1 = NewNetwork("127.0.0.1", 8000, &node1chan)

    }()
    go func() {
        fmt.Println("starting node 2")
        node2 = NewNetwork("127.0.0.1", 8001, &node2chan)
    }()
    <-node1chan
    <-node2chan
    contact := NewContact(NewKademliaIDRandom(), node1.myAddress.String())
    node2.SendPingMessage(&contact);
    close(node1chan)
    close(node2chan)
}
