package kademlia

import (
    "net"
    "time"
    "fmt"
    "log"
    "github.com/vmihailenco/msgpack"
    "reflect"
)

const ALPHA = 3
const CONNECTION_TIMEOUT = time.Second * 2
const (
    NET_STATUS_LISTENING = iota
)
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
    statusChannel *chan int
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

func NewNetwork(ip string, port int, statusChannel *chan int) *Network {
    network := new(Network)
    network.myAddress.IP = net.ParseIP(ip)
    if network.myAddress.IP == nil {
        log.Fatal(fmt.Errorf("Unresolvable network address: %v\n", ip))
    }
    network.myAddress.Port = port
    // Random ID on network start
    network.routing = NewRoutingTable(NewContact(NewKademliaIDRandom(), network.myAddress.String()))
    network.statusChannel = statusChannel
    // Start listening to UDP socket
    go network.Listen()
    return network
}

func (network *Network) Listen() {
    var err error
    network.connection, err = net.ListenUDP("udp", &network.myAddress)
    if err != nil {
        // Fail if we cannot listen on that address
        log.Fatal(err)
    }
    // Message that node is now listening
    *network.statusChannel <- NET_STATUS_LISTENING

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
        // Store the contact that just messaged the node
        contact := NewContact(message.Origin.ID, remote_addr.String())
        network.routing.AddContact(contact) // TODO mutex lock
        fmt.Printf("%v received from %v: %v \n", network.myAddress.String(), remote_addr, message.String())

        switch {
        case message.MsgType == PONG:
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

// Send a message over an established connection
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

// Send a message over a new connection
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

// Send and block until reply
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
            n, err := connection.Read(buf)
            if err != nil {
                timer.Stop()
                log.Printf("%v read UDP failed from %v: %v\n", network.routing.me.Address, contact.Address, err)
                m <- nil
                return
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

func (network *Network) SendFindContactMessage(contactToFind *Contact) ([]Contact) {
    // Marshal the Contact to find for later sending
    contactToFindMsg, err := msgpack.Marshal(*contactToFind)
    if err != nil {
        log.Printf("%v Could not marshal contact: %v\n", network.routing.me, err)
        return nil
    }
    // Find the alpha closest nodes
    closestContacts := network.routing.FindClosestContacts(contactToFind.ID, ALPHA)
    // How many nodes we have queried so far
    numNodesVisited := 0
    // This holds the nodes we have already queried
    nodesVisited := NewRoutingTable(network.routing.me)
    for _, contact := range closestContacts {
        nodesVisited.AddContact(contact)
        numNodesVisited++
    }
    // Mutex http://www.golangpatterns.info/concurrency/semaphores
    mut := make(chan struct{}, 1)

    var lookup func(closestContacts []Contact) []Contact
    lookup = func(closestContacts []Contact) []Contact {
        // Channels for sending/receiving network messages
        rpcChannels := []chan *NetworkMessage{}
        rpcIDs := []KademliaID{}
        for i := 0; i < len(closestContacts); i++ {
            rpcChannels = append(rpcChannels, make(chan *NetworkMessage))
            rpcIDs = append(rpcIDs, *NewKademliaIDRandom())
        }
        // Send go routine RPCs to the closest contacts, connect to channels
        for i := range closestContacts {
            msg := NetworkMessage{MsgType: FIND_CONTACT_MSG, Origin: network.routing.me, RpcID: rpcIDs[i], Data: contactToFindMsg}
            go func(msg *NetworkMessage, contact *Contact, chanIndex chan *NetworkMessage) int {
                set := []reflect.SelectCase{reflect.SelectCase{
                    Dir:  reflect.SelectSend,
                    Chan: reflect.ValueOf(chanIndex),
                    Send: reflect.ValueOf(network.SendReceiveMessage(msg, contact)),
                }}
                to, _, _ := reflect.Select(set)
                return to
            }(&msg, &closestContacts[i], rpcChannels[i])
        }
        // Channels to recursive lookup calls
        lookupChannels := []chan []Contact{}
        // There are as many RPC channels as closest contacts
        for i := 0; i < len(closestContacts); i++ {
            // Block until we get one or more responses from RPCs
            set := []reflect.SelectCase{}
            for _, ch := range rpcChannels {
                set = append(set, reflect.SelectCase{
                    Dir:  reflect.SelectRecv,
                    Chan: reflect.ValueOf(ch),
                })
            }
            chIdx, valValue, _ := reflect.Select(set)
            response := valValue.Interface().(*NetworkMessage)

            // Check if the response is valid
            if response == nil {
                // UDP connection/read/write failed, or timed out
                continue

            } else if response.MsgType == FIND_CONTACT_MSG {
                if !response.RpcID.Equals(&rpcIDs[chIdx]) {
                    log.Printf("%v wrong RPC header from %v: %v should be %v\n", network.myAddress.String(), response.Origin.Address, response.RpcID.String(), rpcIDs[chIdx])
                }
                fmt.Printf("%v received from %v: %v \n", network.myAddress.String(), response.Origin.Address, response.String())
                // Unmarshal the contacts we got back
                var newContacts []Contact
                err := msgpack.Unmarshal(response.Data, &newContacts)
                if err != nil {
                    log.Printf("%v Could not unmarshal contact array: %v\n", network.routing.me, err)
                    continue
                }
                nodesToVisit := []Contact{}
                // Check if we have already visited these contacts. If not, queue them for future visits.
                for i := 0; i < min(ALPHA, len(newContacts)); i++ {
                    newContact := newContacts[i]
                    // Mutex lock here to synchronize shared variables
                    mut <- struct{}{}
                    if !nodesVisited.Contains(newContact) && !network.routing.me.ID.Equals(newContact.ID) && numNodesVisited < bucketSize {
                        nodesToVisit = append(nodesToVisit, newContact)
                        nodesVisited.AddContact(newContact)
                        numNodesVisited++
                        fmt.Printf("%v new contact from %v: %v\n", network.myAddress.String(), response.Origin.Address, newContact.String())
                    }
                    <-mut
                }
                // If there were any new nodes, visit them now
                if len(nodesToVisit) > 0 {
                    callChannel := make(chan []Contact)
                    lookupChannels = append(lookupChannels, callChannel)

                    // Make new recursive lookup calls and store the channels
                    go func(fun func(c []Contact) []Contact, input []Contact, ch chan []Contact) int {
                        set := []reflect.SelectCase{}
                        set = append(set, reflect.SelectCase{
                            Dir:  reflect.SelectSend,
                            Chan: reflect.ValueOf(ch),
                            Send: reflect.ValueOf(fun(input)),
                        })
                        to, _, _ := reflect.Select(set)
                        return to
                    }(lookup, nodesToVisit, callChannel)
                }
            } else {
                log.Printf("%v received UNKNOWN %v: %v \n", network.myAddress.String(), response.Origin.Address, response.String())
            }
        }
        // Gather results from all the recursive calls we made and return them
        allContacts := []Contact{}
        for i := 0; i < len(lookupChannels); i++ {
            set := []reflect.SelectCase{}
            for _, ch := range lookupChannels {
                set = append(set, reflect.SelectCase{
                    Dir:  reflect.SelectRecv,
                    Chan: reflect.ValueOf(ch),
                })
            }
            _, valValue, _ := reflect.Select(set)
            for _, newContact := range valValue.Interface().([]Contact) {
                allContacts = append(allContacts, newContact)
            }
        }
        return allContacts
    }
    // Block on the initial call to the recursive lookup
    lookup(closestContacts)
    // The temporary routing table will contain the closest contacts found during lookup
    return nodesVisited.FindClosestContacts(contactToFind.ID, bucketSize)
}

func (network *Network) SendFindDataMessage(hash string) {
    // TODO
}

func (network *Network) SendStoreMessage(data []byte) {
    // TODO
}
