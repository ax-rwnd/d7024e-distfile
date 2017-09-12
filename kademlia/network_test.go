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
    // This message must get the correct response
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
    // This message should not get a response, so node1 should timeout when listening
    msg := &NetworkMessage{MsgType: PONG, Origin: node1.routing.me, RpcID: *NewKademliaIDRandom()}
    response := node1.SendReceiveMessage(msg, &node2.routing.me)
    if response != nil {
        fmt.Printf("%v\n", response.String())
        t.Fail()
    }
}

func TestSendFindContactMessage(t *testing.T) {
    node1c := make(chan *Network)
    node2c := make(chan *Network)
    NewNetwork(&node1c, "127.0.0.1", 8400)
    NewNetwork(&node2c, "127.0.0.1", 8401)
    node1 := <-node1c
    node2 := <-node2c
    // Do not sort by ID when inputting contacts
    contact1 := node2.routing.AddContact(NewContact(NewKademliaID("FFFFFFFF00000000000000000000000001000000"), "127.0.0.1:8402"))
    contact2 := node2.routing.AddContact(NewContact(NewKademliaID("FFFFFFFF00000000000000000001000000000000"), "127.0.0.1:8402"))
    contact0 := node2.routing.AddContact(NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000000"), "127.0.0.1:8402"))
    // Send find message
    contacts := node1.SendFindContactMessage(contact0, &node2.routing.me)
    // Contacts should be sorted in the response
    if contacts == nil || !contacts[0].Equals(contact0) || !contacts[1].Equals(contact1) || !contacts[2].Equals(contact2) {
        t.Fail()
    }
}
