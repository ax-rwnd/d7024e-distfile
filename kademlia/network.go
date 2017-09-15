package kademlia

import (
    "net"
    "time"
    "fmt"
    "log"
    "github.com/vmihailenco/msgpack"
    "strconv"
)

const ALPHA = 3
const CONNECTION_TIMEOUT = time.Second * 2
const CONNECTION_RETRY_DELAY = time.Second / 2
const (
    FIND_CONTACT_MSG = iota
    FIND_DATA_MSG
    STORE_DATA_MSG
    PING
    PONG
)
const RECEIVE_BUFFER_SIZE = 1 << 20

// Msgpack package requires public variables
type NetworkMessage struct {
    MsgType int
    Origin  Contact
    RpcID   KademliaID
    Data    []byte
}

type Network struct {
    // If not nil, Listen will channel messages here after processing them
    listenChannel chan NetworkMessage
    // Kademlia routing table
    Routing *RoutingTable
    // <Key, Value> store
    store *KVStore
}

func (msg *NetworkMessage) String() string {
    return fmt.Sprintf("MsgType=%v, Origin=%v, RpcID=%v, Data=%v", msg.MsgType, msg.Origin.String(), msg.RpcID.String(), msg.Data)
}

func min(a, b int) int {
    if a <= b {
        return a
    }
    return b
}

func NewNetwork(ip string, port int) *Network {
    network := new(Network)
    // Random ID on network start
    network.Routing = NewRoutingTable(NewContact(NewKademliaIDRandom(), ip+":"+strconv.Itoa(port)))
    // Key value store
    network.store = NewKVStore()
    // Start listening to UDP socket
    listening := make(chan bool)
    go network.Listen(&listening)
    <-listening
    return network
}

func (network *Network) handlePingMessage(connection *net.UDPConn, remote_addr *net.UDPAddr, message *NetworkMessage) {
    // Respond to the ping
    msg := NetworkMessage{MsgType: PONG, Origin: network.Routing.Me, RpcID: message.RpcID}
    go network.SendMessageToConnection(&msg, remote_addr, connection)
}

func (network *Network) handleFindContactMessage(connection *net.UDPConn, remote_addr *net.UDPAddr, message *NetworkMessage) {
    // Unmarshal the contact from data field. Then find the k closest neighbors to it.
    var findTarget KademliaID
    err := msgpack.Unmarshal(message.Data, &findTarget)
    if err != nil {
        log.Printf("%v malformed message from %v: %v\n", network.Routing.Me.Address, remote_addr, err)
        return
    }
    closestContacts := network.Routing.FindClosestContacts(&findTarget, bucketSize)
    // Marshal the closest contacts and send them in the response
    closestContactsMsg, err := msgpack.Marshal(closestContacts)
    if err != nil {
        fmt.Printf("%v failed to marshal contact list with %v\n", network.Routing.Me.Address, err)
        return
    }
    msg := NetworkMessage{MsgType: FIND_CONTACT_MSG, Origin: network.Routing.Me, RpcID: message.RpcID, Data: closestContactsMsg}
    go network.SendMessageToConnection(&msg, remote_addr, connection)
}

func (network *Network) handleStoreDataMessage(connection *net.UDPConn, remote_addr *net.UDPAddr, message *NetworkMessage) {
    // Store a non-marshalled kademlia id as key (file hash), and marshalled contacts as value (file owners)
    var key KademliaID
    err := msgpack.Unmarshal(message.Data, &key)
    if err != nil {
        log.Printf("%v malformed message from %v: %v\n", network.Routing.Me.Address, remote_addr, err)
        return
    }
    var owners []Contact
    // Check if we have it
    if value, err := network.store.Lookup(key); err == nil {
        err := msgpack.Unmarshal(value, &owners)
        if err != nil {
            // The content of this <key,value> is not a contact list, but a file. Do nothing.
            return
        }
    } else {
        owners = []Contact{}
    }
    owners = append(owners, message.Origin)
    fmt.Printf("%v\n", owners)
    marshaledOwners, err := msgpack.Marshal(owners)
    if err != nil {
        log.Printf("%v failed to marshal value from %v: %v\n", network.Routing.Me.Address, remote_addr, err)
        return
    }
    network.store.Insert(key, false, marshaledOwners)
    fmt.Printf("%v stored hash key %v from %v\n", network.Routing.Me.Address, key.String(), message.Origin.String())
}

func (network *Network) handleFindDataMessage(connection *net.UDPConn, remote_addr *net.UDPAddr, message *NetworkMessage) {
    // Read the file hash (kvStore key) requested
    var hash KademliaID
    err := msgpack.Unmarshal(message.Data, &hash)
    if err != nil {
        log.Printf("%v malformed message from %v: %v\n", network.Routing.Me.Address, remote_addr, err)
        return
    }
    // Check if we have it
    value, err := network.store.Lookup(hash)
    if err != nil {
        // Key not in store, reply with empty message
        msg := NetworkMessage{MsgType: FIND_DATA_MSG, Origin: network.Routing.Me, RpcID: message.RpcID}
        go network.SendMessageToConnection(&msg, remote_addr, connection)
        return
    }
    // <Key,Value> exists
    var owners []Contact
    err = msgpack.Unmarshal(value, &owners)
    if err == nil {
        // <Key,Value> exists, and value is the contacts of file owners
        fmt.Printf("%v sends to %v <key,value> pair <%v,%v>\n", network.Routing.Me.Address, remote_addr, hash.String(), owners)
        msg := NetworkMessage{MsgType: FIND_DATA_MSG, Origin: network.Routing.Me, RpcID: message.RpcID, Data: value}
        go network.SendMessageToConnection(&msg, remote_addr, connection)
    } else {
        // <Key,Value> exists, and is a file. TODO: use TCP
        fmt.Printf("%v sends TCP file transfer to %v\n", network.Routing.Me.Address, remote_addr)
        msg := NetworkMessage{MsgType: FIND_DATA_MSG, Origin: network.Routing.Me, RpcID: message.RpcID}
        go network.SendMessageToConnection(&msg, remote_addr, connection)
    }
}

func (network *Network) handleMessage(connection *net.UDPConn, remote_addr *net.UDPAddr, message *NetworkMessage) {
    // Store the contact that just messaged the node
    contact := NewContact(message.Origin.ID, remote_addr.String())
    network.Routing.AddContact(contact)
    fmt.Printf("%v received from %v: %v \n", network.Routing.Me.Address, remote_addr, message.String())
    switch {
    case message.MsgType == PING:
        network.handlePingMessage(connection, remote_addr, message)
    case message.MsgType == FIND_CONTACT_MSG:
        network.handleFindContactMessage(connection, remote_addr, message)
    case message.MsgType == STORE_DATA_MSG:
        network.handleStoreDataMessage(connection, remote_addr, message)
    case message.MsgType == FIND_DATA_MSG:
        network.handleFindDataMessage(connection, remote_addr, message)
    default:
        log.Printf("%v received unknown message from %v: %v\n", network.Routing.Me.Address, remote_addr)
    }
}

// Listen for incoming UDP connections, until network.networkChannel is closed
func (network *Network) Listen(listening *chan bool) {
    var err error
    ip, port, err := network.Routing.Me.ParseAddress()
    udpAddr := net.UDPAddr{IP: net.ParseIP(ip), Port: port}
    connection, err := net.ListenUDP("udp", &udpAddr)
    if err != nil {
        // Fail if we cannot listen on that address
        log.Fatal(err)
    }
    // Message that node is now listening
    *listening <- true

    defer connection.Close()

    buf := make([]byte, RECEIVE_BUFFER_SIZE)
    for {
        // Block until message is available, then unmarshal the package
        _, remote_addr, err := connection.ReadFromUDP(buf)
        var message NetworkMessage
        err = msgpack.Unmarshal(buf, &message)
        if err != nil {
            log.Printf("%v malformed message from %v: %v\n", network.Routing.Me.Address, remote_addr, err)
            continue
        }
        network.handleMessage(connection, remote_addr, &message)
        if network.listenChannel != nil {
            network.listenChannel <- message
        }
    }
}

// Send a message over an established UDP connection
func (network *Network) SendMessageToConnection(message *NetworkMessage, address *net.UDPAddr, conn *net.UDPConn) {
    fmt.Printf("%v responds to %v: %v \n", network.Routing.Me.Address, address, message.String())
    msg, err := msgpack.Marshal(message)
    if err != nil {
        log.Printf("%v failed to marshal network message with %v\n", network.Routing.Me.Address, err)
    }
    _, err = conn.WriteToUDP(msg, address)
    if err != nil {
        log.Printf("%v UDP write failed with %v\n", network.Routing.Me.Address, err)
    }
}

// Send a one-way message
func (network *Network) SendMessage(message *NetworkMessage, contact *Contact) (net.Conn, error) {
    connection, err := net.Dial("udp", contact.Address)
    if err != nil {
        log.Printf("%v connection to %v failed with %v\n", network.Routing.Me.Address, contact.Address, err)
        return nil, err
    }
    fmt.Printf("%v sends to %v: %v\n", network.Routing.Me.Address, contact.Address, message.String())
    msg, err := msgpack.Marshal(message)
    connection.Write(msg)
    return connection, nil
}

// Send over UDP, then block until response or timeout
func (network *Network) SendReceiveMessage(message *NetworkMessage, contact *Contact) *NetworkMessage {
    connection, err := network.SendMessage(message, contact)
    if err != nil {
        return nil
    }
    defer connection.Close()
    timer := time.NewTimer(CONNECTION_TIMEOUT)
    channel := make(chan *NetworkMessage)
    go func(m chan *NetworkMessage) {
        for {
            buf := make([]byte, RECEIVE_BUFFER_SIZE)
            for {
                n, err := connection.Read(buf)
                if err != nil {
                    timeout := time.NewTimer(CONNECTION_RETRY_DELAY)
                    <-timeout.C
                    continue
                }
                var responseMsg NetworkMessage
                err = msgpack.Unmarshal(buf[:n], &responseMsg)
                timer.Stop()
                if err != nil {
                    log.Printf("%v malformed message from %v: %v\n", network.Routing.Me.Address, contact.Address, err)
                    m <- nil
                    return
                }
                m <- &responseMsg
                return
            }
        }
    }(channel)
    select {
    case msg := <-channel:
        return msg
    case <-timer.C:
        log.Printf("%v connection timeout to %v\n", network.Routing.Me.Address, contact.Address)
        return nil
    }
}

// Ping another node with a UDP packet
func (network *Network) SendPingMessage(contact *Contact) bool {
    if contact.Address == network.Routing.Me.Address {
        // Node pinged itself
        return true
    }
    msg := &NetworkMessage{MsgType: PING, Origin: network.Routing.Me, RpcID: *NewKademliaIDRandom()}
    response := network.SendReceiveMessage(msg, contact)
    if response == nil {
        return false
    }
    fmt.Printf("%v received from %v: %v\n", network.Routing.Me.Address, contact.Address, response.String())
    if response.MsgType == PONG && response.RpcID.Equals(&msg.RpcID) {
        // Node responded to ping, so add it to routing table
        network.Routing.AddContact(NewContact(response.Origin.ID, contact.Address))
        return true
    }
    return false
}

// Send a Find Node message over UDP. Blocks until response or timeout.
func (network *Network) SendFindContactMessage(findTarget *KademliaID, receiver *Contact) ([]Contact) {
    // Marshal the contact and store it in Data byte array later
    findTargetMsg, err := msgpack.Marshal(*findTarget)
    if err != nil {
        log.Printf("%v Could not marshal contact: %v\n", network.Routing.Me, err)
        return []Contact{}
    }
    // Unique id for this RPC
    rpcID := *NewKademliaIDRandom()
    msg := NetworkMessage{MsgType: FIND_CONTACT_MSG, Origin: network.Routing.Me, RpcID: rpcID, Data: findTargetMsg}
    // Blocks until response
    response := network.SendReceiveMessage(&msg, receiver)
    // Validate the response
    if response != nil && response.MsgType == FIND_CONTACT_MSG {
        if !response.RpcID.Equals(&rpcID) {
            log.Printf("%v wrong RPC ID from %v: %v should be %v\n", network.Routing.Me.Address, response.Origin.Address, response.RpcID.String(), rpcID)
        }
        fmt.Printf("%v received from %v: %v \n", network.Routing.Me.Address, response.Origin.Address, response.String())
        // Unmarshal the contacts we got back
        var newContacts []Contact
        err := msgpack.Unmarshal(response.Data, &newContacts)
        if err != nil {
            log.Printf("%v Could not unmarshal contact array: %v\n", network.Routing.Me, err)
        }
        return newContacts
    } else if response != nil {
        log.Printf("%v received unknown message %v: %v \n", network.Routing.Me.Address, response.Origin.Address, response.String())
    }
    return []Contact{}
}

func (network *Network) SendFindDataMessage(hash *KademliaID, receiver *Contact) []Contact {
    // Marshal the contact and store it in Data byte array later
    hashMsg, err := msgpack.Marshal(*hash)
    if err != nil {
        log.Printf("%v Could not marshal kademlia id: %v\n", network.Routing.Me, err)
        return []Contact{}
    }
    // Unique id for this RPC
    rpcID := *NewKademliaIDRandom()
    msg := NetworkMessage{MsgType: FIND_DATA_MSG, Origin: network.Routing.Me, RpcID: rpcID, Data: hashMsg}
    // Blocks until response
    response := network.SendReceiveMessage(&msg, receiver)
    // Validate the response
    if response != nil && response.MsgType == FIND_DATA_MSG {
        if !response.RpcID.Equals(&rpcID) {
            log.Printf("%v wrong RPC ID from %v: %v should be %v\n", network.Routing.Me.Address, response.Origin.Address, response.RpcID.String(), rpcID)
        }
        fmt.Printf("%v received from %v: %v \n", network.Routing.Me.Address, response.Origin.Address, response.String())
        // Unmarshal the contacts we got back, if any
        var newContacts []Contact
        err := msgpack.Unmarshal(response.Data, &newContacts)
        if err != nil {
            return []Contact{}
        }
        return newContacts
    } else if response != nil {
        log.Printf("%v received unknown message %v: %v \n", network.Routing.Me.Address, response.Origin.Address, response.String())
    }
    return []Contact{}
}

func (network *Network) SendStoreMessage(hash *KademliaID, receiver *Contact) {
    hashMsg, err := msgpack.Marshal(hash)
    if err != nil {
        log.Printf("%v Could not marshal kademlia ID %v\n", network.Routing.Me, hash)
    }
    message := NetworkMessage{MsgType: STORE_DATA_MSG, Origin: network.Routing.Me, RpcID: *NewKademliaIDRandom(), Data: hashMsg}
    network.SendMessage(&message, receiver)
}
