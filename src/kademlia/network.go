package kademlia

import (
    "net"
    "time"
    "fmt"
    "log"
    "github.com/vmihailenco/msgpack"
    "rpc"
    "strconv"
)

const (
    TCP = iota
    UDP
)
const CONNECTION_TIMEOUT = time.Second * 2
const CONNECTION_RETRY_DELAY = time.Second
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
    // <Key, Value> Store
    Store *KVStore
}

func (msg *NetworkMessage) String() string {
    dataMsg := fmt.Sprintf("%v", msg.Data)
    if len(msg.Data) > 1<<10 { // Not larger than 1 kB
        dataMsg = strconv.Itoa(len(msg.Data)) + " bytes"
    }
    return fmt.Sprintf("MsgType=%v, Origin=%v, RpcID=%v, Data=%v", rpc.EnumToString(msg.MsgType), msg.Origin.String(), msg.RpcID.String(), dataMsg)
}

func min(a, b int) int {
    if a <= b {
        return a
    }
    return b
}

// Create a new network and start listening to incoming TCP connections and UDP packets
func NewNetwork(ip string, tcpPort int, udpPort int) *Network {
    network := new(Network)
    // Random ID on network start
    network.Routing = NewRoutingTable(NewContact(NewKademliaIDRandom(), ip, tcpPort, udpPort))
    // Key value Store
    network.Store = NewKVStore()
    // Start listening to UDP socket
    listening := make(chan bool)
    go network.Listen(&listening)
    <-listening
    return network
}

// Someone sent a ping message, respond to it
func (network *Network) receivePingMessage(connection net.PacketConn, remote_addr net.Addr, message *NetworkMessage) {
    // Respond to the ping
    msg := NetworkMessage{MsgType: rpc.PONG_MSG, Origin: network.Routing.Me, RpcID: message.RpcID}
    go network.SendMessageToUdpConnection(&msg, remote_addr, connection)
}

// Someone wants to know k of our contacts closest to a kademlia ID
func (network *Network) receiveFindContactMessage(connection net.PacketConn, remote_addr net.Addr, message *NetworkMessage) {
    // Unmarshal the contact from data field. Then find the k closest neighbors to it.
    var findTarget KademliaID
    err := msgpack.Unmarshal(message.Data, &findTarget)
    if err != nil {
        log.Printf("%v malformed message from %v: %v\n", network.Routing.Me.Address, remote_addr, err)
        return
    }
    closestContacts := network.Routing.FindClosestContacts(&findTarget, K)
    // Marshal the closest contacts and send them in the response
    closestContactsMsg, err := msgpack.Marshal(closestContacts)
    if err != nil {
        fmt.Printf("%v failed to marshal contact list with %v\n", network.Routing.Me.Address, err)
        return
    }
    msg := NetworkMessage{MsgType: rpc.FIND_CONTACT_MSG, Origin: network.Routing.Me, RpcID: message.RpcID, Data: closestContactsMsg}
    go network.SendMessageToUdpConnection(&msg, remote_addr, connection)
}

// Someone wants us to Store a kademlia ID (file hash) along with their contact information in our <key,value> Store
func (network *Network) receiveStoreDataMessage(connection net.PacketConn, remote_addr net.Addr, message *NetworkMessage) {
    // Store a non-marshalled kademlia id as key (file hash), and marshalled contacts as value (file owners)
    var key KademliaID
    err := msgpack.Unmarshal(message.Data, &key)
    if err != nil {
        log.Printf("%v malformed message from %v: %v\n", network.Routing.Me.Address, remote_addr, err)
        return
    }
    var owners []Contact
    // Check if we have it
    if value, err := network.Store.Lookup(key); err == nil {
        if err := msgpack.Unmarshal(value, &owners); err != nil {
            // The content of this <key,value> is not a contact list, but a file. Do nothing.
            return
        }
    } else {
        owners = []Contact{}
    }
    owners = append(owners, message.Origin)
    fmt.Printf("%v has contacts %v for hash %vh\n", network.Routing.Me.Address, owners, key.String())
    marshaledOwners, err := msgpack.Marshal(owners)
    if err != nil {
        log.Printf("%v failed to marshal value from %v: %v\n", network.Routing.Me.Address, remote_addr, err)
        return
    }
    network.Store.Insert(key, false, marshaledOwners)
    fmt.Printf("%v stored hash key %v from %v\n", network.Routing.Me.Address, key.String(), message.Origin.String())
}

// Someone wants to query our <key,value> Store for a file hash and know which contacts it can be downloaded from
func (network *Network) receiveFindDataMessage(connection net.PacketConn, remote_addr net.Addr, message *NetworkMessage) {
    // Read the file hash (kvStore key) requested
    var hash KademliaID
    err := msgpack.Unmarshal(message.Data, &hash)
    if err != nil {
        log.Printf("%v malformed message from %v: %v\n", network.Routing.Me.Address, remote_addr, err)
        return
    }
    // Check if we have it
    value, err := network.Store.Lookup(hash)
    if err != nil {
        // Key not in Store, reply with empty message
        fmt.Printf("%v cannot find <key,value> for key=%v\n", network.Routing.Me.Address, hash.String())
        msg := NetworkMessage{MsgType: rpc.FIND_DATA_MSG, Origin: network.Routing.Me, RpcID: message.RpcID}
        go network.SendMessageToUdpConnection(&msg, remote_addr, connection)
        return
    }
    // <Key,Value> exists
    var owners []Contact
    err = msgpack.Unmarshal(value, &owners)
    if err != nil {
        owners = []Contact{network.Routing.Me}
    }
    // <Key,Value> exists, and value is the contacts of file owners
    fmt.Printf("%v sends to %v <key,value> pair <%v,%v>\n", network.Routing.Me.Address, remote_addr, hash.String(), owners)
    msg := NetworkMessage{MsgType: rpc.FIND_DATA_MSG, Origin: network.Routing.Me, RpcID: message.RpcID, Data: value}
    go network.SendMessageToUdpConnection(&msg, remote_addr, connection)

}

// Someone wants to download stored files from us
func (network *Network) receiveTransferDataMessage(connection net.Conn, message *NetworkMessage) {
    var hash KademliaID
    err := msgpack.Unmarshal(message.Data, &hash)
    if err != nil {
        log.Printf("%v invalid hash from %v: %v\n", network.Routing.Me.Address, connection.RemoteAddr().String(), err)
        return
    }
    data, err := network.Store.Lookup(hash)
    if err != nil {
        log.Printf("%v cannot find data for %v: %v\n", network.Routing.Me.Address, connection.RemoteAddr().String(), hash.String())
        return
    }
    response := NetworkMessage{MsgType: rpc.TRANSFER_DATA_MSG, Origin: network.Routing.Me, RpcID: message.RpcID, Data: data}
    marshaledResponse, err := msgpack.Marshal(response)
    if err != nil {
        log.Printf("%v invalid hash from %v: %v\n", network.Routing.Me.Address, connection.RemoteAddr().String(), err)
        return
    }
    fmt.Printf("%v responds to %v: %v\n", network.Routing.Me.Address, connection.RemoteAddr().String(), response.String())
    connection.Write(marshaledResponse)
}

// Someone initiated a TCP connection, check if they want to download data from us
func (network *Network) receiveTCP(connection net.Conn) {
    buffer := make([]byte, RECEIVE_BUFFER_SIZE)
    _, err := connection.Read(buffer)
    if err != nil {
        log.Printf("%v unreadable TCP message from %v: %v\n", network.Routing.Me.Address, connection.RemoteAddr().String(), err)
        return
    }
    var message NetworkMessage
    err = msgpack.Unmarshal(buffer, &message)
    if err != nil {
        log.Printf("%v malformed message from %v: %v\n", network.Routing.Me.Address, connection.RemoteAddr().String(), err)
        return
    }
    // Store the contact that just messaged the node
    network.Routing.AddContact(message.Origin, network.SendPingMessage)
    fmt.Printf("%v received from %v: %v \n", network.Routing.Me.Address, connection.RemoteAddr().String(), message.String())
    switch {
    case message.MsgType == rpc.TRANSFER_DATA_MSG:
        network.receiveTransferDataMessage(connection, &message)
    default:
        log.Printf("%v received unknown message from %v: %v\n", network.Routing.Me.Address, connection.RemoteAddr().String(), message)
    }
    connection.Close()
}

// Someone sent us a UDP packet, check if it is an RPC message and handle it in that case
func (network *Network) receiveUDP(connection net.PacketConn) {
    buf := make([]byte, RECEIVE_BUFFER_SIZE)
    _, remote_address, err := connection.ReadFrom(buf)
    if err != nil {
        log.Printf("%v unreadable UDP packet from %v: %v\n", network.Routing.Me.Address, remote_address, err)
        return
    }
    var message NetworkMessage
    err = msgpack.Unmarshal(buf, &message)
    if err != nil {
        log.Printf("%v malformed message from %v: %v\n", network.Routing.Me.Address, remote_address, err)
        return
    }
    // Store the contact that just messaged the node
    network.Routing.AddContact(message.Origin, network.SendPingMessage)
    fmt.Printf("%v received from %v: %v \n", network.Routing.Me.Address, remote_address, message.String())
    switch {
    case message.MsgType == rpc.PING_MSG:
        network.receivePingMessage(connection, remote_address, &message)
    case message.MsgType == rpc.FIND_CONTACT_MSG:
        network.receiveFindContactMessage(connection, remote_address, &message)
    case message.MsgType == rpc.STORE_DATA_MSG:
        network.receiveStoreDataMessage(connection, remote_address, &message)
    case message.MsgType == rpc.FIND_DATA_MSG:
        network.receiveFindDataMessage(connection, remote_address, &message)
    default:
        log.Printf("%v received unknown message from %v: %v\n", network.Routing.Me.Address, remote_address, message)
    }
    if network.listenChannel != nil {
        network.listenChannel <- message
    }
}

// Listen for incoming TCP and UDP connections
func (network *Network) Listen(listening *chan bool) {
    tcpAddress := network.Routing.Me.Address.IP + ":" + strconv.Itoa(network.Routing.Me.Address.TcpPort)
    udpAddress := network.Routing.Me.Address.IP + ":" + strconv.Itoa(network.Routing.Me.Address.UdpPort)
    tcpChannel := make(chan bool)
    udpChannel := make(chan bool)

    // TCP connections
    tcpListen, err := net.Listen("tcp", tcpAddress)
    if err != nil {
        log.Fatal(err)
    }
    defer tcpListen.Close()
    go func(channel chan bool) {
        for {
            connection, err := tcpListen.Accept()
            if err != nil {
                log.Fatal(err)
            }
            channel <- true
            go network.receiveTCP(connection)
        }
    }(tcpChannel)

    // UDP packets listen
    udpListen, err := net.ListenPacket("udp", udpAddress)
    if err != nil {
        log.Fatal(err)
    }
    defer udpListen.Close()
    go func(channel chan bool) {
        for {
            channel <- true
            // Cannot call this in a go routine since UDP has no blocking accept
            network.receiveUDP(udpListen)
        }
    }(udpChannel)

    // Listen has been called for both UDP and TCP
    *listening <- true
    // Infinite loop listening to TCP and UDP sockets
    for {
        select {
        case <-tcpChannel:
            // TCP message received
        case <-udpChannel:
            // UDP message received
        }
    }
}

// Send a message over an established UDP connection
func (network *Network) SendMessageToUdpConnection(message *NetworkMessage, address net.Addr, conn net.PacketConn) {
    fmt.Printf("%v responds to %v: %v \n", network.Routing.Me.Address, address, message.String())
    msg, err := msgpack.Marshal(message)
    if err != nil {
        log.Printf("%v failed to marshal network message with %v\n", network.Routing.Me.Address, err)
    }
    _, err = conn.WriteTo(msg, address)
    if err != nil {
        log.Printf("%v UDP write failed with %v\n", network.Routing.Me.Address, err)
    }
}

// Send a one-way message
func (network *Network) SendMessage(protocol int, message *NetworkMessage, contact *Contact) (net.Conn, error) {
    var port int
    var protoStr string
    if protocol == UDP {
        port = contact.Address.UdpPort
        protoStr = "udp"
    } else {
        port = contact.Address.TcpPort
        protoStr = "tcp"
    }
    connection, err := net.Dial(protoStr, contact.Address.IP+":"+strconv.Itoa(port))
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
func (network *Network) SendReceiveMessage(protocol int, message *NetworkMessage, contact *Contact) *NetworkMessage {
    connection, err := network.SendMessage(protocol, message, contact)
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
                n := 0
                if protocol == UDP {
                    // For UDP connections, just read one datagram (should be an RPC)
                    n, err = connection.Read(buf)
                    if err != nil {
                        timeout := time.NewTimer(CONNECTION_RETRY_DELAY)
                        <-timeout.C
                        continue
                    }
                } else {
                    // For TCP file transfers, we can keep reading until there is no more left
                    for {
                        newBuf := make([]byte, RECEIVE_BUFFER_SIZE)
                        newN, err := connection.Read(newBuf)
                        if newN <= 0 {
                            break
                        }
                        timer.Reset(CONNECTION_TIMEOUT)
                        if err != nil {
                            timeout := time.NewTimer(CONNECTION_RETRY_DELAY)
                            <-timeout.C
                            continue
                        }
                        buf = append(buf[:n], newBuf[:newN]...)
                        n = n + newN
                    }
                }
                timer.Stop()
                // Unmarshal the message and return it
                var responseMsg NetworkMessage
                err = msgpack.Unmarshal(buf[:n], &responseMsg)
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
    msg := &NetworkMessage{MsgType: rpc.PING_MSG, Origin: network.Routing.Me, RpcID: *NewKademliaIDRandom()}
    response := network.SendReceiveMessage(UDP, msg, contact)
    if response == nil {
        return false
    }
    fmt.Printf("%v received from %v: %v\n", network.Routing.Me.Address, contact.Address, response.String())
    if response.MsgType == rpc.PONG_MSG && response.RpcID.Equals(&msg.RpcID) {
        // Node responded to ping, so add it to routing table
        //network.Routing.AddContact(response.Origin)
        return true
    }
    return false
}

// Send a Find Node message over UDP. Blocks until response or timeout.
func (network *Network) SendFindContactMessage(findTarget *KademliaID, receiver *Contact) ([]Contact) {
    // Marshal the contact and Store it in Data byte array later
    findTargetMsg, err := msgpack.Marshal(*findTarget)
    if err != nil {
        log.Printf("%v could not marshal contact: %v\n", network.Routing.Me, err)
        return []Contact{}
    }
    // Unique id for this RPC
    rpcID := *NewKademliaIDRandom()
    msg := NetworkMessage{MsgType: rpc.FIND_CONTACT_MSG, Origin: network.Routing.Me, RpcID: rpcID, Data: findTargetMsg}
    // Blocks until response
    response := network.SendReceiveMessage(UDP, &msg, receiver)
    // Validate the response
    if response != nil && response.MsgType == rpc.FIND_CONTACT_MSG {
        if !response.RpcID.Equals(&rpcID) {
            log.Printf("%v wrong RPC ID from %v: %v should be %v\n", network.Routing.Me.Address, response.Origin.Address, response.RpcID.String(), rpcID)
        }
        fmt.Printf("%v received from %v: %v \n", network.Routing.Me.Address, response.Origin.Address, response.String())
        // Unmarshal the contacts we got back
        var newContacts []Contact
        err := msgpack.Unmarshal(response.Data, &newContacts)
        if err != nil {
            log.Printf("%v could not unmarshal contact array: %v\n", network.Routing.Me, err)
        }
        return newContacts
    } else if response != nil {
        log.Printf("%v received unknown message %v: %v \n", network.Routing.Me.Address, response.Origin.Address, response.String())
    }
    return []Contact{}
}

// Added this function so can retrieve node ID for the boot node when bootstraping
func (network *Network) FindContactAndID(findTarget *KademliaID, receiver *Contact) ([]Contact, KademliaID) {
	// Marshal the contact and Store it in Data byte array later
    findTargetMsg, err := msgpack.Marshal(*findTarget)
    if err != nil {
        log.Printf("%v could not marshal contact: %v\n", network.Routing.Me, err)
        return []Contact{}, KademliaID{}
    }
    // Unique id for this RPC
    rpcID := *NewKademliaIDRandom()
    msg := NetworkMessage{MsgType: rpc.FIND_CONTACT_MSG, Origin: network.Routing.Me, RpcID: rpcID, Data: findTargetMsg}
    // Blocks until response
    response := network.SendReceiveMessage(UDP, &msg, receiver)
    // Validate the response
    if response != nil && response.MsgType == rpc.FIND_CONTACT_MSG {
        if !response.RpcID.Equals(&rpcID) {
            log.Printf("%v wrong RPC ID from %v: %v should be %v\n", network.Routing.Me.Address, response.Origin.Address, response.RpcID.String(), rpcID)
        }
        fmt.Printf("%v received from %v: %v \n", network.Routing.Me.Address, response.Origin.Address, response.String())
        // Unmarshal the contacts we got back
        var newContacts []Contact
        err := msgpack.Unmarshal(response.Data, &newContacts)
        if err != nil {
            log.Printf("%v could not unmarshal contact array: %v\n", network.Routing.Me, err)
        }
        return newContacts, *response.Origin.ID
    } else if response != nil {
        log.Printf("%v received unknown message %v: %v \n", network.Routing.Me.Address, response.Origin.Address, response.String())
    }
    return []Contact{}, KademliaID{}
}

// Search for owners of a particular file, using its hash
func (network *Network) SendFindDataMessage(hash *KademliaID, receiver *Contact) []Contact {
    // Marshal the contact and Store it in Data byte array later
    hashMsg, err := msgpack.Marshal(*hash)
    if err != nil {
        log.Printf("%v could not marshal kademlia id: %v\n", network.Routing.Me, err)
        return []Contact{}
    }
    message := NetworkMessage{MsgType: rpc.FIND_DATA_MSG, Origin: network.Routing.Me, RpcID: *NewKademliaIDRandom(), Data: hashMsg}
    // Blocks until response
    response := network.SendReceiveMessage(UDP, &message, receiver)
    // Validate the response
    if response != nil && response.MsgType == rpc.FIND_DATA_MSG {
        if !response.RpcID.Equals(&message.RpcID) {
            log.Printf("%v wrong RPC ID from %v: %v should be %v\n", network.Routing.Me.Address, response.Origin.Address, response.RpcID.String(), message.RpcID.String())
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

// Tell another node to Store <hash,me> as <key,value>
func (network *Network) SendStoreMessage(hash *KademliaID, receiver *Contact) {
    hashMsg, err := msgpack.Marshal(hash)
    if err != nil {
        log.Printf("%v could not marshal kademlia ID %v\n", network.Routing.Me, hash)
    }
    message := NetworkMessage{MsgType: rpc.STORE_DATA_MSG, Origin: network.Routing.Me, RpcID: *NewKademliaIDRandom(), Data: hashMsg}
    network.SendMessage(UDP, &message, receiver)
}

// Request a file transfer from message receiver
func (network *Network) SendDownloadMessage(hash *KademliaID, receiver *Contact) []byte {
    hashMsg, err := msgpack.Marshal(hash)
    if err != nil {
        log.Printf("%v could not marshal kademlia ID %v\n", network.Routing.Me, hash)
    }
    message := NetworkMessage{MsgType: rpc.TRANSFER_DATA_MSG, Origin: network.Routing.Me, RpcID: *NewKademliaIDRandom(), Data: hashMsg}
    response := network.SendReceiveMessage(TCP, &message, receiver)
    fmt.Printf("%v downloaded from %v: %v\n", network.Routing.Me.Address, response.Origin.Address, response.String())
    if response != nil && response.MsgType == rpc.TRANSFER_DATA_MSG && response.RpcID.Equals(&message.RpcID) {
        return response.Data
    }
    return []byte{}
}
