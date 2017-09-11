package kademlia

import (
    "net"
    "fmt"
    "log"
    "github.com/vmihailenco/msgpack"
    "reflect"
)

const ALPHA = 3
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

// msgpack package requires public variables
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
        log.Fatal(fmt.Errorf("network: Unresolvable address: %v\n", ip))
    }
    network.myAddress.Port = port
    // TODO: Should we grab a random ID on network start?
    network.routing = NewRoutingTable(NewContact(NewKademliaIDRandom(), network.myAddress.String()))
    network.statusChannel = statusChannel
    // Start listening
    go network.Listen()

    return network
}

func (network *Network) Listen() {
    var err error
    network.connection, err = net.ListenUDP("udp", &network.myAddress)
    if err != nil {
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
            log.Fatal(err)
        }
        // Store the contact that just messaged the node
        contact := NewContact(message.Origin.ID, remote_addr.String())
        network.routing.AddContact(contact)
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
                log.Fatal(err)
            }
            closestContacts := network.routing.FindClosestContacts(contactToFind.ID, bucketSize)
            // Marshal the closest contacts and send them in the response
            closestContactsMsg, err := msgpack.Marshal(closestContacts)
            if err != nil {
                log.Fatal(err)
            }
            msg := NetworkMessage{MsgType: FIND_CONTACT_MSG, Origin: network.routing.me, RpcID: message.RpcID, Data: closestContactsMsg}
            go network.SendMessageToConnection(&msg, remote_addr, network.connection)
        case err != nil:
            log.Fatal(err)
        }
    }
}

// Send a message over an established connection
func (network *Network) SendMessageToConnection(message *NetworkMessage, address *net.UDPAddr, conn *net.UDPConn) {
    fmt.Printf("%v responds to %v: %v \n", network.myAddress.String(), address, message.String())
    msg, err := msgpack.Marshal(message)
    _, err = conn.WriteToUDP(msg, address)
    if err != nil {
        fmt.Printf("%v WriteToUDP failed with %v\n", network.routing.me.Address, err)
    }
}

// Send a message over a new connection
func (network *Network) SendMessage(message *NetworkMessage, contact *Contact) net.Conn {
    connection, err := net.Dial("udp", contact.Address)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("%v sends to %v: %v\n", network.myAddress.String(), contact.Address, message.String())
    msg, err := msgpack.Marshal(message)
    connection.Write(msg)
    return connection
}

// Send and block until reply
func (network *Network) SendReceiveMessage(message *NetworkMessage, contact *Contact) *NetworkMessage {
    connection := network.SendMessage(message, contact)
    defer connection.Close()
    // TODO timeout
    for {
        buf := make([]byte, RECEIVE_BUFFER_SIZE)
        n, err := connection.Read(buf)
        if err != nil {
            log.Fatal(err)
        }
        var responseMsg NetworkMessage
        err = msgpack.Unmarshal(buf[:n], &responseMsg)
        if err != nil {
            log.Fatal(err)
        }
        return &responseMsg
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

    msg := NetworkMessage{MsgType: FIND_CONTACT_MSG, Origin: network.routing.me, RpcID: *NewKademliaIDRandom(), Data: contactToFindMsg}

    // Find the alpha closest node
    closestContacts := network.routing.FindClosestContacts(contactToFind.ID, ALPHA)

    // Channels for sending/receiving network messages
    sendChannels := []chan *NetworkMessage{}
    for i := 0; i < bucketSize; i++ {
        sendChannels = append(sendChannels, make(chan *NetworkMessage))
    }
    chanIndex := 0
    // This holds the nodes we have already queried
    nodesVisited := NewRoutingTable(network.routing.me)
    // How many nodes we have queried so far
    numNodesVisited := 0
    // Mutex http://www.golangpatterns.info/concurrency/semaphores
    m1 := make(chan struct{}, 1)
    m2 := make(chan struct{}, 1)

    var lookup func(closestContacts []Contact) []Contact
    lookup = func(closestContacts []Contact) []Contact {
        // Send queries to the closest contacts
        m1 <- struct{}{}
        for i := range closestContacts {
            //go network.callSendReceiveOnChannel(&msg, &closestContacts[i], sendChannels[chanIndex])
            go func(msg *NetworkMessage, contact *Contact, ch chan *NetworkMessage) int {
                set := []reflect.SelectCase{}
                set = append(set, reflect.SelectCase{
                    Dir:  reflect.SelectSend,
                    Chan: reflect.ValueOf(ch),
                    Send: reflect.ValueOf(network.SendReceiveMessage(msg, contact)),
                })
                to, _, _ := reflect.Select(set)
                return to
            }(&msg, &closestContacts[i], sendChannels[chanIndex])
            chanIndex++
        }
        <-m1
        // Channels to recursive calls
        callChannels := []chan []Contact{}
        // There are as many query channels as closest contacts
        for i := 0; i < len(closestContacts); i++ {
            // Block until we get one or more responses from RPCs
            //response, _ := readSendReceiveOnChannels(sendChannels)
            set := []reflect.SelectCase{}
            for _, ch := range sendChannels {
                set = append(set, reflect.SelectCase{
                    Dir:  reflect.SelectRecv,
                    Chan: reflect.ValueOf(ch),
                })
            }
            _, valValue, _ := reflect.Select(set)
            response := valValue.Interface().(*NetworkMessage)

            if response.MsgType == FIND_CONTACT_MSG && response.RpcID.Equals(&msg.RpcID) {
                fmt.Printf("%v received from %v: %v \n", network.myAddress.String(), response.Origin.Address, response.String())
                // Unmarshal the contacts we got back
                var newContacts []Contact
                err := msgpack.Unmarshal(response.Data, &newContacts)
                if err != nil {
                    log.Printf("%v Could not unmarshal contact array: %v\n", network.routing.me, err)
                    continue
                }
                toVisit := []Contact{}
                // Check if we have already visited these contacts. If not, queue them for future visits. Populate the routing table also.
                m2 <- struct{}{}
                for i := 0; i < min(ALPHA, len(newContacts)); i++ {
                    newContact := newContacts[i]
                    if !nodesVisited.Contains(newContact) && !network.routing.me.ID.Equals(newContact.ID) && numNodesVisited < bucketSize {
                        nodesVisited.AddContact(newContact)
                        toVisit = append(toVisit, newContact)
                        numNodesVisited++
                        fmt.Printf("%v new contact from %v: %v\n", network.myAddress.String(), response.Origin.Address, newContact.String())
                    }
                }
                // If there were any new nodes, visit them now
                if len(toVisit) > 0 {
                    callChannel := make(chan []Contact)
                    callChannels = append(callChannels, callChannel)

                    go func(fun func(c []Contact) []Contact, input []Contact, ch chan []Contact) int {
                        set := []reflect.SelectCase{}
                        set = append(set, reflect.SelectCase{
                            Dir:  reflect.SelectSend,
                            Chan: reflect.ValueOf(ch),
                            Send: reflect.ValueOf(fun(input)),
                        })
                        to, _, _ := reflect.Select(set)
                        return to
                    }(lookup, toVisit, callChannel)

                }
                <-m2
            }
        }
        // Gather results from all the recursive calls we made and return them
        allContacts := []Contact{}
        for i := 0; i < len(callChannels); i++ {
            set := []reflect.SelectCase{}
            for _, ch := range callChannels {
                set = append(set, reflect.SelectCase{
                    Dir:  reflect.SelectRecv,
                    Chan: reflect.ValueOf(ch),
                })
            }
            _, valValue, _ := reflect.Select(set)
            newContacts := valValue.Interface().([]Contact)

            for _, newContact := range newContacts {
                allContacts = append(allContacts, newContact)
            }
        }
        return allContacts
    }
    for _, contact := range closestContacts {
        nodesVisited.AddContact(contact)
        numNodesVisited++
    }
    // Block on the initial call to the recursive lookup
    lookup(closestContacts)
    // The temporary routing table will contain the closest contacts found during lookup
    closestContacts = nodesVisited.FindClosestContacts(contactToFind.ID, bucketSize)
    for _, channel := range sendChannels {
        close(channel)
    }
    return closestContacts
}

func (network *Network) SendFindDataMessage(hash string) {
    // TODO
}

func (network *Network) SendStoreMessage(data []byte) {
    // TODO
}
