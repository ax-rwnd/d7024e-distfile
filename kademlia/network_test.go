package kademlia

import (
    "testing"
    "fmt"
)

func ping(sender *Network, receiver *Contact, c chan bool) {
    c <- sender.SendPingMessage(receiver)
}

func TestUDPing(t *testing.T) {
    node1c := make(chan *Network)
    node2c := make(chan *Network)
    node3c := make(chan *Network)
    NewNetwork(&node1c, "127.0.0.1", 8000)
    NewNetwork(&node2c, "127.0.0.1", 8001)
    NewNetwork(&node3c, "127.0.0.1", 8002)
    node1 := <-node1c
    node2 := <-node2c
    node3 := <-node3c
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

    // If any node did not respond to ping, fail the test
    for i := 0; i < 6; i++ {
        var c bool
        select {
        case c = <-ping21:
        case c = <-ping23:
        case c = <-ping31:
        case c = <-ping32:
        case c = <-ping13:
        case c = <-ping12:
        }
        if c == false {
            t.Fail()
        }
    }
}

func TestSendReceiveMessage(t *testing.T) {
    node1c := make(chan *Network)
    node2c := make(chan *Network)
    NewNetwork(&node1c, "127.0.0.1", 8100)
    NewNetwork(&node2c, "127.0.0.1", 8101)
    node1 := <-node1c
    node2 := <-node2c
    msg := &NetworkMessage{MsgType: PING, Origin: node1.routing.me, RpcID: *NewKademliaIDRandom()}
    response := node1.SendReceiveMessage(msg, &node2.routing.me)
    if response.MsgType != PONG || !response.RpcID.Equals(&msg.RpcID) || !response.Origin.ID.Equals(node2.routing.me.ID) {
        t.Fail()
    }
}

func TestSendReceiveMessageTimeout(t *testing.T) {
    node1c := make(chan *Network)
    node2c := make(chan *Network)
    NewNetwork(&node1c, "127.0.0.1", 8200)
    NewNetwork(&node2c, "127.0.0.1", 8201)
    node1 := <-node1c
    node2 := <-node2c
    msg := &NetworkMessage{MsgType: PONG, Origin: node1.routing.me, RpcID: *NewKademliaIDRandom()}
    response := node1.SendReceiveMessage(msg, &node2.routing.me)
    if response != nil {
        fmt.Printf("%v\n", response.String())
        t.Fail()
    }
}

func TestUDPFindContact(t *testing.T) {
    // Create nodes array and double link contact information between them.
    // This means for more nodes than bucketSize+1 this test will fail
    const numNodes = bucketSize + 1
    var nodeChans [numNodes]chan *Network
    var nodes [numNodes] *Network
    for i := range nodeChans {
        nodeChans[i] = make(chan *Network)
        NewNetwork(&nodeChans[i], "127.0.0.1", 8300+i)
        nodes[i] = <-nodeChans[i] // Node is now listening to UDP connections
    }
    fmt.Printf("looking up %v\n", nodes[numNodes-1].routing.me.String())
    // Add some contacts between them
    for i := range nodeChans {
        if i == 0 {
            nodes[0].routing.AddContact(nodes[1].routing.me)
        } else if i == numNodes-1 {
            nodes[numNodes-1].routing.AddContact(nodes[numNodes-2].routing.me)
        } else {
            nodes[i].routing.AddContact(nodes[i-1].routing.me)
            nodes[i].routing.AddContact(nodes[i+1].routing.me)
        }
    }
    var cc = []chan []Contact{make(chan []Contact), make(chan []Contact),}
    // First node does not yet have last node as a contact. Find it.
    go func() {
        cc[0] <- nodes[0].SendFindContactMessage(&nodes[numNodes-1].routing.me)
    }()
    // Try the reverse concurrently
    go func() {
        cc[1] <- nodes[numNodes-1].SendFindContactMessage(&nodes[0].routing.me)
    }()
    contacts1 := <-cc[0]
    contacts2 := <-cc[1]
    fmt.Printf("%v lookup %v found %v\n", nodes[0].routing.me.Address, nodes[numNodes-1].routing.me.ID.String(), contacts1)
    fmt.Printf("%v lookup %v found %v\n", nodes[numNodes-1].routing.me.Address, nodes[0].routing.me.ID.String(), contacts2)
    if !contacts1[0].ID.Equals(nodes[numNodes-1].routing.me.ID) {
        t.Fail()
    }
    if !contacts2[0].ID.Equals(nodes[0].routing.me.ID) {
        t.Fail()
    }
}
