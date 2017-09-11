package kademlia

import (
    "testing"
    "fmt"
)

func ping(sender *Network, receiver *Contact, c chan bool) {
    c <- sender.SendPingMessage(receiver)
}

func TestUDPing(t *testing.T) {
    var node1 *Network
    var node2 *Network
    var node3 *Network

    node1chan := make(chan int)
    node2chan := make(chan int)
    node3chan := make(chan int)

    go func() {
        node1 = NewNetwork("127.0.0.1", 8000, &node1chan)
    }()
    go func() {
        node2 = NewNetwork("127.0.0.1", 8001, &node2chan)
    }()
    go func() {
        node3 = NewNetwork("127.0.0.1", 8002, &node3chan)
    }()

    nodesListening := 0
    for nodesListening < 3 {
        var c int
        select {
        case c = <-node1chan:
        case c = <-node2chan:
        case c = <-node3chan:
        }
        if c == NET_STATUS_LISTENING {
            nodesListening++
        }
    }
    // Nodes are now listening to UDP connections
    ping21 := make(chan bool)
    go ping(node2, &node1.routing.me, ping21)

    ping23 := make(chan bool)
    go ping(node2, &node3.routing.me, ping23)

    ping31 := make(chan bool)
    go ping(node3, &node1.routing.me, ping31)

    ping32 := make(chan bool)
    go ping(node3, &node2.routing.me, ping32)

    ping13 := make(chan bool)
    go ping(node1, &node3.routing.me, ping13)

    ping12 := make(chan bool)
    go ping(node1, &node2.routing.me, ping12)

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

func TestUDPFindContact(t *testing.T) {
    var node1 *Network
    var node2 *Network
    var node3 *Network
    var node4 *Network
    var node5 *Network

    node1chan := make(chan int)
    node2chan := make(chan int)
    node3chan := make(chan int)
    node4chan := make(chan int)
    node5chan := make(chan int)

    go func() {
        node1 = NewNetwork("127.0.0.1", 9000, &node1chan)
    }()
    go func() {
        node2 = NewNetwork("127.0.0.1", 9001, &node2chan)
    }()
    go func() {
        node3 = NewNetwork("127.0.0.1", 9002, &node3chan)
    }()
    go func() {
        node4 = NewNetwork("127.0.0.1", 9003, &node4chan)
    }()
    go func() {
        node5 = NewNetwork("127.0.0.1", 9004, &node5chan)
    }()
    nodesListening := 0
    for nodesListening < 5 {
        var c int
        select {
        case c = <-node1chan:
        case c = <-node2chan:
        case c = <-node3chan:
        case c = <-node4chan:
        case c = <-node5chan:
        }
        if c == NET_STATUS_LISTENING {
            nodesListening++
        }
    }
    // Nodes are now listening to UDP connections
    fmt.Printf("Network %v <-> %v <-> %v <-> %v <-> %v\n",
        node1.routing.me.String(),
        node2.routing.me.String(),
        node3.routing.me.String(),
        node4.routing.me.String(),
        node5.routing.me.String())

    node1.routing.AddContact(node2.routing.me)
    node1.routing.AddContact(node3.routing.me)
    node2.routing.AddContact(node1.routing.me)
    node2.routing.AddContact(node3.routing.me)
    node3.routing.AddContact(node2.routing.me)
    node3.routing.AddContact(node4.routing.me)
    node4.routing.AddContact(node3.routing.me)
    node4.routing.AddContact(node5.routing.me)
    node5.routing.AddContact(node4.routing.me)
    fmt.Printf("looking up %v\n", node5.routing.me.String())

    closestContacts := make(chan []Contact)
    go func() {
        closestContacts <- node1.SendFindContactMessage(&node5.routing.me)
    }()
    contacts := <-closestContacts
    fmt.Printf("%v lookup %v found %v\n", node1.routing.me.Address, node5.routing.me.ID.String() , contacts)
    if !contacts[0].ID.Equals(node5.routing.me.ID) {
        t.Fail()
    }
    close(closestContacts)

    close(node1chan)
    close(node2chan)
    close(node3chan)
    close(node4chan)
}
