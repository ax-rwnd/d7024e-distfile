package kademlia

import (
    "testing"
    "fmt"
    "rpc"
    "github.com/vmihailenco/msgpack"
    "io/ioutil"
    "encoding/hex"
)

var testPort int = 7000

func getTestPort() int {
    testPort++
    return testPort
}

func ping(sender *Network, receiver *Contact, c chan bool) {
    c <- sender.SendPingMessage(receiver)
}

// Test UDP packet pinging between nodes
func TestUDPing(t *testing.T) {
    node1 := NewNetwork("127.0.0.1", getTestPort(), getTestPort())
    node2 := NewNetwork("127.0.0.1", getTestPort(), getTestPort())
    node3 := NewNetwork("127.0.0.1", getTestPort(), getTestPort())
    // Nodes are now listening to UDP connections
    ping21 := make(chan bool)
    go ping(node2, &node1.Routing.Me, ping21)

    ping23 := make(chan bool)
    go ping(node2, &node3.Routing.Me, ping23)

    ping31 := make(chan bool)
    go ping(node3, &node1.Routing.Me, ping31)

    ping32 := make(chan bool)
    go ping(node3, &node2.Routing.Me, ping32)

    ping13 := make(chan bool)
    go ping(node1, &node3.Routing.Me, ping13)

    ping12 := make(chan bool)
    go ping(node1, &node2.Routing.Me, ping12)

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
    node1.Close()
    node2.Close()
    node3.Close()
}

// Test sending a ping message between two nodes generates the correct response
func TestSendReceiveMessage(t *testing.T) {
    node1 := NewNetwork("127.0.0.1", getTestPort(), getTestPort())
    node2 := NewNetwork("127.0.0.1", getTestPort(), getTestPort())
    // This message must get the correct response
    msg := &NetworkMessage{MsgType: rpc.PING_MSG, Origin: node1.Routing.Me, RpcID: *NewKademliaIDRandom()}
    response := node1.SendReceiveMessage(UDP, msg, &node2.Routing.Me)
    if response.MsgType != rpc.PONG_MSG || !response.RpcID.Equals(&msg.RpcID) || !response.Origin.ID.Equals(node2.Routing.Me.ID) {
        t.Fail()
    }
    node1.Close()
    node2.Close()
}

// This UDP message should not generate a response from the other node, it should time out waiting for it.
func TestSendReceiveMessageTimeoutUDP(t *testing.T) {
    node1 := NewNetwork("127.0.0.1", getTestPort(), getTestPort())
    node2 := NewNetwork("127.0.0.1", getTestPort(), getTestPort())
    // This message should not get a response, so node1 should timeout when listening
    msg := &NetworkMessage{MsgType: rpc.PONG_MSG, Origin: node1.Routing.Me, RpcID: *NewKademliaIDRandom()}
    response := node1.SendReceiveMessage(UDP, msg, &node2.Routing.Me)
    if response != nil {
        fmt.Printf("%v\n", response.String())
        t.Fail()
    }
    node1.Close()
    node2.Close()
}

// This TCP message should not generate a response from the other node, it should time out waiting for it.
func TestSendReceiveMessageTimeoutTCP(t *testing.T) {
    node1 := NewNetwork("127.0.0.1", getTestPort(), getTestPort())
    node2 := NewNetwork("127.0.0.1", getTestPort(), getTestPort())
    // This message should not get a response, so node1 should timeout when listening
    msg := &NetworkMessage{MsgType: rpc.PONG_MSG, Origin: node1.Routing.Me, RpcID: *NewKademliaIDRandom()}
    response := node1.SendReceiveMessage(TCP, msg, &node2.Routing.Me)
    if response != nil {
        fmt.Printf("%v\n", response.String())
        t.Fail()
    }
    node1.Close()
    node2.Close()
}

// Test that the correct response is given when finding contacts on other nodes
func TestSendFindContactMessage(t *testing.T) {
    node1 := NewNetwork("127.0.0.1", getTestPort(), getTestPort())
    node2 := NewNetwork("127.0.0.1", getTestPort(), getTestPort())
    // Do not sort by ID when inputting contacts
    _, contact1 := node2.Routing.AddContact(NewContact(NewKademliaID("FFFFFFFF00000000000000000000000001000000"), "127.0.0.1", getTestPort(), getTestPort()), nil)
    _, contact2 := node2.Routing.AddContact(NewContact(NewKademliaID("FFFFFFFF00000000000000000001000000000000"), "127.0.0.1", getTestPort(), getTestPort()), nil)
    _, contact0 := node2.Routing.AddContact(NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000000"), "127.0.0.1", getTestPort(), getTestPort()), nil)
    // Send find message
    contacts := node1.SendFindContactMessage(contact0.ID, &node2.Routing.Me)
    // Contacts should be sorted in the response
    if contacts == nil || !contacts[0].Equals(contact0) || !contacts[1].Equals(contact1) || !contacts[2].Equals(contact2) {
        t.Fail()
    }
    node1.Close()
    node2.Close()
}

// Test that UDP based SendReceiveMessage fails correctly on connection failure
func TestUDPConnectionFail(t *testing.T) {
    node1 := NewNetwork("127.0.0.1", getTestPort(), getTestPort())
    _, contact := node1.Routing.AddContact(NewContact(NewKademliaIDRandom(), "127.0.0.1", 999998, 999999), nil)
    // Connection will fail since port is invalid. - response should be nil
    msg := &NetworkMessage{MsgType: 0, Origin: node1.Routing.Me, RpcID: *NewKademliaIDRandom()}
    response := node1.SendReceiveMessage(UDP, msg, contact)
    if response != nil {
        t.Fail()
    }
    node1.Close()
}

// Send Store message from one node to another, check if it was received and stored
func TestSendStoreMessage(t *testing.T) {
    node1 := NewNetwork("127.0.0.1", getTestPort(), getTestPort())
    node2 := NewNetwork("127.0.0.1", getTestPort(), getTestPort())
    node2.listenChannel = make(chan NetworkMessage)
    hash := NewRandomKademliaID()
    // Send Store message
    node1.SendStoreMessage(hash, &node2.Routing.Me)
    var err error
    var data []byte
    // Wait until node2 has stored the hash
    <-node2.listenChannel
    node2.listenChannel = nil
    data, err = node2.Store.Lookup(*hash)
    // Unmarshal and check if value is ok (file owner contact)
    var value []Contact
    err = msgpack.Unmarshal(data, &value)
    if err != nil || !value[0].Equals(&node1.Routing.Me) {
        t.Fail()
    }
    node1.Close()
    node2.Close()
}

// Put a file hash and file owner into kvStore of node2. See if node1 finds it.
func TestSendFindDataMessage(t *testing.T) {
    node1 := NewNetwork("127.0.0.1", getTestPort(), getTestPort())
    node2 := NewNetwork("127.0.0.1", getTestPort(), getTestPort())
    hash := NewRandomKademliaID()
    marshaledContact, err := msgpack.Marshal([]Contact{node1.Routing.Me})
    if err != nil {
        t.Fail()
    }
    node2.Store.Insert(*hash, false, marshaledContact, nil)
    contacts := node1.SendFindDataMessage(hash, &node2.Routing.Me)
    if contacts == nil || len(contacts) == 0 || !contacts[0].Equals(&node1.Routing.Me) {
        t.Fail()
    }
    node1.Close()
    node2.Close()
}

// Send Store message from one node to another, find if it was received and stored
func TestSendStoreFindMessages(t *testing.T) {
    node1 := NewNetwork("127.0.0.1", getTestPort(), getTestPort())
    node2 := NewNetwork("127.0.0.1", getTestPort(), getTestPort())
    node2.listenChannel = make(chan NetworkMessage)
    hash := NewRandomKademliaID()
    // Send Store message
    node1.SendStoreMessage(hash, &node2.Routing.Me)
    // Wait until node2 has stored the hash
    <-node2.listenChannel
    node2.listenChannel = nil
    // Find the data (owner of file hash)
    contacts := node1.SendFindDataMessage(hash, &node2.Routing.Me)
    if contacts == nil || len(contacts) == 0 || !contacts[0].Equals(&node1.Routing.Me) {
        t.Fail()
    }
    delete(node2.Store.mapping, *hash)
    // Check that node2 no longer finds the data
    contacts = node1.SendFindDataMessage(hash, &node2.Routing.Me)
    if len(contacts) != 0 {
        t.Fail()
    }
    node1.Close()
    node2.Close()
}

// Download data by TCP from one node to another
func TestTcpTransfer(t *testing.T) {
    node1 := NewNetwork("127.0.0.1", getTestPort(), getTestPort())
    node2 := NewNetwork("127.0.0.1", getTestPort(), getTestPort())
    data, _ := ioutil.ReadFile("test.bin")
    hash := NewKademliaIDFromBytes(data)
    // Store data in node 2, then transfer it to node 1
    node2.Store.Insert(*hash, false, data, nil)
    // Send TCP download request
    downloadedData := node1.SendDownloadMessage(hash, &node2.Routing.Me)
    // Check if download worked
    if len(downloadedData) != len(data) {
        t.Fail()
    }
    for i := range data {
        if data[i] != downloadedData[i] {
            t.Fail()
        }
    }
    node1.Close()
    node2.Close()
}

// If routing table bucket is full, ping the last contact, if it does not respond, add the contact.
func TestNetworkAddContactSuccess(t *testing.T) {
    node1 := NewNetwork("127.0.0.1", getTestPort(), getTestPort())
    //node2 := NewNetwork("127.0.0.1", getTestPort(), getTestPort())
    id, _ := hex.DecodeString("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF")
    var i int
    for i = 0; i < ReplicationFactor+1; i++ {
        // Decrease the new ID to one lower than previous
        newId := make([]byte, len(id))
        copy(newId, id)
        newId[len(newId)-1] = id[len(id)-1] - byte(i)
        // Add contact with this ID
        contact := NewContact(NewKademliaID(hex.EncodeToString(newId)), "127.0.0.1", getTestPort(), getTestPort())
        contactWasAdded, _ := node1.Routing.AddContact(contact, node1.SendPingMessage)
        // The last contact will be pinged to see if it is alive, since bucket is full.
        // Since it pings to a port no one listens to it will get no response, so the contact should be added.
        if !contactWasAdded {
            t.Fail()
        }
    }
    node1.Close()
}

// If routing table bucket is full, ping the last contact, if it does respond, do not add the contact.
func TestNetworkAddContactFail(t *testing.T) {
    var networks []*Network
    node1 := NewNetwork("127.0.0.1", getTestPort(), getTestPort())
    networks = append(networks, node1);
    //node2 := NewNetwork("127.0.0.1", getTestPort(), getTestPort())
    id, _ := hex.DecodeString("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF")
    var i int
    for i = 0; i < ReplicationFactor+1; i++ {
        // Decrease the new ID to one lower than previous
        newId := make([]byte, len(id))
        copy(newId, id)
        newId[len(newId)-1] = id[len(id)-1] - byte(i)
        // Add contact with this ID
        nodei := NewNetwork("127.0.0.1", getTestPort(), getTestPort())
        networks = append(networks, nodei);
        kademliaId := NewKademliaID(hex.EncodeToString(newId))
        nodei.Routing.Me.ID = kademliaId
        contactWasAdded, _ := node1.Routing.AddContact(nodei.Routing.Me, node1.SendPingMessage)
        // The last contact will be pinged to see if it is alive, since bucket is full.
        // Since the node will respond to ping, it will not be added
        if i == ReplicationFactor && contactWasAdded {
            t.Fail()
        }
    }
    for _, network := range networks {
        network.Close()
    }
}
