package main

import (
    "errors"
    "log"
    "net"
    "net/http"
    "net/rpc"
    "../shared"
)

type Arith int

func (t *Arith) Multiply(args *shared.Args, reply *int) error {
    *reply = args.A * args.B
    return nil
}

func (t *Arith) Divide(args *shared.Args, quo *shared.Quotient) error {
    if args.B == 0 {
        return errors.New("zero division")
    }

    quo.Quo = args.A / args.B
    quo.Rem = args.A % args.B

    return nil
}

func main() {
    arith := new(Arith)
    rpc.Register(arith)
    rpc.HandleHTTP()
    l, e := net.Listen("tcp", ":9090")
    if e != nil {
        log.Fatal("listen error:", e)
    }

    http.Serve(l, nil)
}
