package kademlia

import (
    "fmt"
    "net"
    "errors"
    "../share"
)

type Network struct {
}

func Listen(msg chan<-share.Message, ip string, port int) error {
    fmt.Println("Opening network on",ip, port)
    addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d",port))
    if err != nil {
        fmt.Println("Failed to resolve",err)
        return errors.New("Failed to resolve")
    }

    conn, err := net.ListenUDP("udp", addr)
    if err != nil {
        fmt.Println("Failed to listen",err)
        return errors.New("Failed to listen")
    }

    fmt.Println("Listening...")
    buf := make([]byte, 1024)
    for {
        fmt.Println("here")
        if n, _, err := conn.ReadFromUDP(buf); err == nil {
            fmt.Println("Here!")
            msg <- share.Message{Argument:n}
        } else {
            return err
        }
    }
}

func (network *Network) SendPingMessage(contact *Contact) {
    // TODO
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
