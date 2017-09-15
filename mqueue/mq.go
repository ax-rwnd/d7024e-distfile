package mqueue

import (
    "net/http"
    "sync"
    "fmt"
    "../kademlia"
    "github.com/vmihailenco/msgpack"
)

var Messages = make(chan Request, 100)
var table map[string]int
var route map[string]string

func Translate(wg *sync.WaitGroup) {
    defer wg.Done()

    for {
        select {
            case msg := <-Messages:
                msg.Fp(msg.Data, msg.ResponseChannel)
                fmt.Println(msg)
        }
    }
}

func SendPingInternal(args Args, r chan<- Response) {
    var parsed ArgsPing

    fmt.Println("in internal")
    if args.MType == ARGS_PING {
        msgpack.Unmarshal(args.Data, &parsed)
    } else {
        panic("Argument type mismatch!")
    }
    reply := parsed.Network.SendPingMessage(parsed.Contact)
    serialized, err := msgpack.Marshal(reply)
    if err != nil {
        panic("Failed to marshal!")
    }
    fmt.Println("writing response :-)")
    r<- Response{MType:RESPONSE_PING, Data:serialized}
}

func RestServer(hostNetwork *kademlia.Network, msg chan<- Request, address string) {
    http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
        if arg_data, err := msgpack.Marshal(
            ArgsPing{Network:hostNetwork,
                Contact:&kademlia.Contact{
                    ID:kademlia.NewKademliaIDRandom(),
                    Address:address}}); err != nil {
            panic("Failed to marshal!")
        } else {
            args := Args{MType:ARGS_PING, Data:arg_data}
            ret_value := make(chan Response, 1)
            msg <- Request{MType:REQUEST_PING, Fp:SendPingInternal, ResponseChannel:ret_value, Data:args}
            result := <-ret_value
            fmt.Println(result)
        }
        })
    if err := http.ListenAndServe(address, nil); err != nil {
        fmt.Println(err)
    }
}

func RPCServer(address string) {
}

//fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
