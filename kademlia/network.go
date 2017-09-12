package kademlia

import (
    "net"
    "time"
    "fmt"
    "log"
    "github.com/vmihailenco/msgpack"
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
    routing   *RoutingTable
    myAddress net.UDPAddr
    // Listening connection
    connection *net.UDPConn
    // Channel for telling when node started listening
    networkChannel *chan *Network
}

func (msg *NetworkMessage) String() string {
    return fmt.Sprintf("MsgType=%v, Origin=%v, RpcID=%v, Data=%v", msg.MsgType, msg.Origin.String(), msg.RpcID, msg.Data)
}

func min(a, b int) int {
    if a <= b {
        return a
    }
    return b
}

func NewNetwork(networkChannel *chan *Network, ip string, port int) {
    network := new(Network)
    network.myAddress.IP = net.ParseIP(ip)
    if network.myAddress.IP == nil {
        log.Fatal(fmt.Errorf("Unresolvable network address: %v\n", ip))
    }
    network.myAddress.Port = port
    // Random ID on network start
    network.routing = NewRoutingTable(NewContact(NewKademliaIDRandom(), network.myAddress.String()))
    network.networkChannel = networkChannel
    // Start listening to UDP socket
    go network.Listen()
}

// Listen for incoming UDP connections, until network.networkChannel is closed
func (network *Network) Listen() {
    var err error
    network.connection, err = net.ListenUDP("udp", &network.myAddress)
    if err != nil {
        // Fail if we cannot listen on that address
        log.Fatal(err)
    }
    // Message that node is now listening
    *network.networkChannel <- network

    defer network.connection.Close()

    buf := make([]byte, RECEIVE_BUFFER_SIZE)
    for {
        // Block until message is available, then unmarshal the package
        _, remote_addr, err := network.connection.ReadFromUDP(buf)
        var message NetworkMessage
        err = msgpack.Unmarshal(buf, &message)
        if err != nil {
            log.Printf("%v malformed message from %v: %v\n", network.routing.me.Address, remote_addr, err)
            continue
        }
        // Store the contact that just messaged the node TODO mutex lock
        //contact := NewContact(message.Origin.ID, remote_addr.String())
        //network.routing.AddContact(contact)

        fmt.Printf("%v received from %v: %v \n", network.myAddress.String(), remote_addr, message.String())

        switch {
        case message.MsgType == PING:
            // Respond to the ping
            msg := NetworkMessage{MsgType: PONG, Origin: network.routing.me, RpcID: message.RpcID}
            go network.SendMessageToConnection(&msg, remote_addr, network.connection)

        case message.MsgType == FIND_CONTACT_MSG:
            // Unmarshal the contact from data field. Then find the k closest neighbors to it.
            var contactToFind Contact
            err = msgpack.Unmarshal(message.Data, &contactToFind)
            if err != nil {
                log.Printf("%v malformed message from %v: %v\n", network.routing.me.Address, remote_addr, err)
                continue
            }
            closestContacts := network.routing.FindClosestContacts(contactToFind.ID, bucketSize)
            // Marshal the closest contacts and send them in the response
            closestContactsMsg, err := msgpack.Marshal(closestContacts)
            if err != nil {
                fmt.Printf("%v failed to marshal contact list with %v\n", network.routing.me.Address, err)
                continue
            }
            msg := NetworkMessage{MsgType: FIND_CONTACT_MSG, Origin: network.routing.me, RpcID: message.RpcID, Data: closestContactsMsg}
            go network.SendMessageToConnection(&msg, remote_addr, network.connection)

        case message.MsgType == FIND_DATA_MSG:

        case message.MsgType == STORE_DATA_MSG:

        default:
            log.Printf("%v received unknown message from %v: %v\n", network.routing.me.Address, remote_addr, err)
        }
    }
}

// Send a message over an established UDP connection
func (network *Network) SendMessageToConnection(message *NetworkMessage, address *net.UDPAddr, conn *net.UDPConn) {
    fmt.Printf("%v responds to %v: %v \n", network.myAddress.String(), address, message.String())
    msg, err := msgpack.Marshal(message)
    if err != nil {
        log.Printf("%v failed to marshal network message with %v\n", network.routing.me.Address, err)
    }
    _, err = conn.WriteToUDP(msg, address)
    if err != nil {
        log.Printf("%v UDP write failed with %v\n", network.routing.me.Address, err)
    }
}

// Send a one-way message
func (network *Network) SendMessage(message *NetworkMessage, contact *Contact) (net.Conn, error) {
    connection, err := net.Dial("udp", contact.Address)
    if err != nil {
        log.Printf("%v connection to %v failed with %v\n", network.routing.me.Address, contact.Address, err)
        return nil, err
    }
    fmt.Printf("%v sends to %v: %v\n", network.myAddress.String(), contact.Address, message.String())
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
                    log.Printf("%v malformed message from %v: %v\n", network.routing.me.Address, contact.Address, err)
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
        log.Printf("%v connection timeout to %v\n", network.routing.me.Address, contact.Address)
        return nil
    }
}

// Ping another node with a UDP packet
func (network *Network) SendPingMessage(contact *Contact) bool {
    if contact.Address == network.myAddress.String() {
        // Node pinged itself
        return true
    }
    msg := &NetworkMessage{MsgType: PING, Origin: network.routing.me, RpcID: *NewKademliaIDRandom()}
    response := network.SendReceiveMessage(msg, contact)
    if response == nil {
        return false
    }
    fmt.Printf("%v received from %v: %v\n", network.myAddress.String(), contact.Address, response.String())
    if response.MsgType == PONG && response.RpcID.Equals(&msg.RpcID) {
        // Node responded to ping, so add it to routing table
        network.routing.AddContact(NewContact(response.Origin.ID, contact.Address))
        return true
    }
    return false
}

// Send a Find Node message over UDP. Blocks until response or timeout.
func (network *Network) SendFindContactMessage(findTarget *Contact, receiver *Contact) ([]Contact) {
    // Marshal the contact and store it in Data byte array later
    contactToFindMsg, err := msgpack.Marshal(*findTarget)
    if err != nil {
        log.Printf("%v Could not marshal contact: %v\n", network.routing.me, err)
        return nil
    }
    // Unique id for this RPC
    rpcID := *NewKademliaIDRandom()
    msg := NetworkMessage{MsgType: FIND_CONTACT_MSG, Origin: network.routing.me, RpcID: rpcID, Data: contactToFindMsg}
    // Blocks until response
    response := network.SendReceiveMessage(&msg, receiver)
    // Validate the response
    if response != nil && response.MsgType == FIND_CONTACT_MSG {
        if !response.RpcID.Equals(&rpcID) {
            log.Printf("%v wrong RPC header from %v: %v should be %v\n", network.routing.me.Address, response.Origin.Address, response.RpcID.String(), rpcID)
        }
        fmt.Printf("%v received from %v: %v \n", network.routing.me.Address, response.Origin.Address, response.String())
        // Unmarshal the contacts we got back
        var newContacts []Contact
        err := msgpack.Unmarshal(response.Data, &newContacts)
        if err != nil {
            log.Printf("%v Could not unmarshal contact array: %v\n", network.routing.me, err)
        }
        return newContacts
    } else {
        log.Printf("%v received UNKNOWN %v: %v \n", network.routing.me.Address, response.Origin.Address, response.String())
    }
    return []Contact{}
}

func (network *Network) SendFindDataMessage(hash string) {
    // TODO
}

func (network *Network) SendStoreMessage(data []byte) {
    // TODO
}
