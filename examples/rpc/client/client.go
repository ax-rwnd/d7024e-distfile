package main

import (
    "fmt"
    "log"
    "net/rpc"
    "../shared"
)

func main() {
    serverAddress := "localhost"
    client, err := rpc.DialHTTP("tcp", serverAddress + ":9090")
    if err != nil {
        log.Fatal("dialing:", err)
    }

    // Synchronous call
    args := &shared.Args{7,8}
    var reply int
    err = client.Call("Arith.Multiply", args, &reply)
    if err != nil {
        log.Fatal("arith error:", err)
    }
    fmt.Printf("Arith: %d*%d=%d", args.A, args.B, reply)

    // Asynchronous call
    /*
    quotient := new(shared.Quotient)
    divCall := client.Go("Arith.Divide", args, quotient, nil)
    replyCall := <-divCall.Done // will be equal to divCall
    */
    // check errors, print, etc.
}
