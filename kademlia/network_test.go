package kademlia

import (
    "fmt"
    "testing"
)

func ping(sender *Network, receiver *Contact, c chan bool) {
    c <- sender.SendPingMessage(receiver)
}

func TestUDP(t *testing.T) {
    var node1 *Network
    var node2 *Network
    var node3 *Network

    node1chan := make(chan int)
    node2chan := make(chan int)
    node3chan := make(chan int)

    go func() {
        fmt.Println("starting node 1")
        node1 = NewNetwork("127.0.0.1", 8000, &node1chan)
    }()
    go func() {
        fmt.Println("starting node 2")
        node2 = NewNetwork("127.0.0.1", 8001, &node2chan)
    }()
    go func() {
        fmt.Println("starting node 3")
        node3 = NewNetwork("127.0.0.1", 8002, &node3chan)
    }()

    nodesListening := 0
    for nodesListening < 3 {
        select {
        case c := <-node1chan:
            if c == NET_STATUS_LISTENING {
                nodesListening++
            }
        case c := <-node2chan:
            if c == NET_STATUS_LISTENING {
                nodesListening++
            }

        case c := <-node3chan:
            if c == NET_STATUS_LISTENING {
                nodesListening++
            }
        }
    }
    // Nodes are now listening to UDP connections
    contact1 := NewContact(NewKademliaIDRandom(), node1.myAddress.String())
    contact2 := NewContact(NewKademliaIDRandom(), node2.myAddress.String())
    contact3 := NewContact(NewKademliaIDRandom(), node3.myAddress.String())

    ping21 := make(chan bool)
    go ping(node2, &contact1, ping21)

    ping23 := make(chan bool)
    go ping(node2, &contact3, ping23)

    ping31 := make(chan bool)
    go ping(node3, &contact1, ping31)

    ping32 := make(chan bool)
    go ping(node3, &contact2, ping32)

    ping13 := make(chan bool)
    go ping(node1, &contact3, ping13)

    ping12 := make(chan bool)
    go ping(node1, &contact2, ping12)

    <-ping21
    <-ping23
    <-ping31
    <-ping32
    <-ping13
    <-ping12

    close(node1chan)
    close(node2chan)
    close(node3chan)
}
