package mqueue

import "../kademlia"

const IDLength = 20
type KademliaID [IDLength]byte

type Contact struct {
    id KademliaID
    address string
    port int
}

const (
    REQUEST_PING uint8 = iota
    RESPONSE_PING
    ARGS_PING

    REQUEST_KVS_LOOKUP
    RESPONSE_KVS_LOOKUP
    ARGS_KVS_LOOKUP
    REQUEST_KVS_INSERT
    RESPONSE_KVS_INSERT
    ARGS_KVS_INSERT

    REQUEST_ROUTE_STORE
    RESPONSE_ROUTE_STORE
    ARGS_ROUTE_STORE
    REQUEST_ROUTE_CLOSEST
    RESPONSE_ROUTE_CLOSEST
    ARGS_ROUTE_CLOSEST
)

type Args struct {
    MType   uint8
    Data    []byte
}

type Response struct {
    MType   uint8
    Data    []byte
}

type Request struct {
    MType   uint8
    ResponseChannel chan Response
    Fp      func(Args, chan<- Response)
    Data    Args
}

type RequestPing struct {
    target Contact
    }
type ResponsePing struct {
    target Contact
}
type ArgsPing struct {
    Contact *kademlia.Contact
    Network *kademlia.Network
}
