package main

import (
    "../kademlia"
    "fmt"
    "html"
    "net/http"
    "../share"
)

func Translate(inmsg chan share.Message, config *daemonConfig) {
    for {
        select {
            case msg := <-inmsg:
                //TODO: execute
                stdlog.Println(msg)
                if msg.Function == nil {
                    errlog.Println("Warning: nil message", msg)
                } else {
                    go msg.Function(msg.Argument)
                }
        }
    }
}

func RESTServer(msg chan<- share.Message, config *daemonConfig) {
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        msg <- share.Message{Argument: 5}
        fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
        })
    if err := http.ListenAndServe(":8095", nil); err != nil {
        errlog.Println(err)
    }
}

func RPCServer(msg chan<- share.Message, config *daemonConfig) {
    //Spin up udp servers
    for i := 0; i<config.Alpha; i++ {
        go kademlia.Listen(msg, config.Address, 8096+i)
    }
}
