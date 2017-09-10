package kademlia

import (
    "net"
    "fmt"
    "log"
    "github.com/vmihailenco/msgpack"
)

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

// msgpack package requires struct variables to use upper case it seems, or (un)marshalling fails
type NetworkMessage struct {
    MsgType int
    Data []byte
}

type Network struct {
    routing       *RoutingTable
    myAddress     net.UDPAddr
    statusChannel *chan int
    connection    *net.UDPConn
}

func NewNetwork(ip string, port int, statusChannel *chan int) *Network {
    network := new(Network)
    network.myAddress.IP = net.ParseIP(ip)
    if network.myAddress.IP == nil {
        log.Fatal(fmt.Errorf("network: Unresolvable address: %v\n", ip))
    }
    network.myAddress.Port = port
    network.routing = NewRoutingTable(NewContact(NewKademliaIDRandom(), network.myAddress.String()))
    network.statusChannel = statusChannel

    go network.Listen()

    return network
}

func (network *Network) Listen() {
    var err error
    network.connection, err = net.ListenUDP("udp", &network.myAddress)
    if err != nil {
        log.Fatal(err)
    }
    *network.statusChannel <- NET_STATUS_LISTENING

    defer network.connection.Close()

    buf := make([]byte, 2048)
    for {
        n, remote_addr, err := network.connection.ReadFromUDP(buf)
        var msg NetworkMessage
        err = msgpack.Unmarshal(buf, &msg)
        if err != nil {
            log.Fatal(err)
        }
        switch {
        case msg.MsgType == PING:
            fmt.Printf("%v received %v from %v\n", network.myAddress.String(), "PING", remote_addr)
            network.routing.AddContact(NewContact(NewKademliaIDRandom(), remote_addr.String())) // TODO ID
            network.RespondPingMessage(network.connection, remote_addr)
        case n > 0:
            fmt.Printf("%v received %v from %v\n", network.myAddress.String(), buf[:n], remote_addr)
        case err != nil:
            log.Fatal(err)
        }
    }
}

func (network *Network) RespondPingMessage(conn *net.UDPConn, address *net.UDPAddr) {
    fmt.Printf("%v sends %v to %v\n", network.myAddress.String(), "PONG", address)
    msg, err := msgpack.Marshal(&NetworkMessage{MsgType: PONG})
    _, err = conn.WriteToUDP(msg, address)
    if err != nil {
        fmt.Printf("Error %v sending PONG to %v", err, address)
    }
}

func (network *Network) SendPingMessage(contact *Contact) bool {
    if contact.Address == network.myAddress.String() {
        // Node pinged itself
        return true
    }
    connection, err := net.Dial("udp", contact.Address)
    if err != nil {
        fmt.Printf("Error %v dialing contact %v\n", err, *contact)
        return false
    }
    defer connection.Close()
    fmt.Printf("%v sends %v to %v\n", network.myAddress.String(), "PING", contact.Address)
    msg, err := msgpack.Marshal(&NetworkMessage{MsgType: PING})
    connection.Write(msg)
    // TODO timeout
    for {
        buf := make([]byte, 2048)
        n, err := connection.Read(buf)
        if err != nil {
            fmt.Printf("Error %v listening to %v\n", err, *contact)
            return false
        }
        var response NetworkMessage
        err = msgpack.Unmarshal(buf[:n], &response)
        if err != nil {
            log.Fatal(err)
        }
        if response.MsgType == PONG {
            fmt.Printf("%v received %v from %v\n", network.myAddress.String(), "PONG", contact.Address)
            return true
        }
    }
    return false
}

func (network *Network) SendFindContactMessage(contact *Contact) {
    // TODO
}

func (network *Network) SendFindDataMessage(hash string) {
    // TODO
}

func (network *Network) SendStoreMessage(data []byte) {
    // TODO
}
