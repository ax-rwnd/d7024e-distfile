package kademlia

import (
    "net"
    "fmt"
    "log"
    "bufio"
)

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
    *network.statusChannel <- 1

    defer network.connection.Close()

    buf := make([]byte, 2048)
    for {
        n, remote_addr, err := network.connection.ReadFromUDP(buf)
        message := string(buf[:n])

        switch {
        case message == "ping":
            fmt.Printf("%v received %v from %v\n", network.myAddress.String(), message, remote_addr)
            network.routing.AddContact(NewContact(NewKademliaIDRandom(), remote_addr.String())) // TODO ID
            network.RespondPingMessage(network.connection, remote_addr)
        case n > 0:
            fmt.Printf("%v received %v from %v\n", network.myAddress.String(), message, remote_addr)
        case err != nil:
            log.Fatal(err)
        }
    }
}

func (network *Network) RespondPingMessage(conn *net.UDPConn, address *net.UDPAddr) {
    fmt.Printf("%v sends %v to %v\n", network.myAddress.String(), "pong", address)
    _, err := conn.WriteToUDP([]byte("pong"), address)
    if err != nil {
        fmt.Printf("Couldn't send response %v to %v", err, address)
    }
}

func (network *Network) SendPingMessage(contact *Contact) {
    conn, err := net.Dial("udp", contact.Address)
    if err != nil {
        fmt.Printf("Error %v dialing contact %v\n", err, *contact)
        return
    }
    fmt.Printf("%v sends %v to %v\n", network.myAddress.String(), "ping", contact.Address)
    fmt.Fprintf(conn, "ping")
    buf := make([]byte, 2048)
    n, err := bufio.NewReader(conn).Read(buf)
    message := string(buf[:n])

    fmt.Printf("%v received %v from %v\n", network.myAddress.String(), message, contact.Address)

    conn.Close()
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
