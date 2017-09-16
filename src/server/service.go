package main

import (
    "os/signal"
    "github.com/takama/daemon"
    "os"
    "syscall"
)

var dependencies = []string{"dummy.service"}

type Service struct {
    daemon.Daemon
}

func runDaemon() {
    srv, err := daemon.New("kademliad", "Kademlia Storage Daemon", dependencies...)
    if err != nil {
        errlog.Println("Error: ", err)
        os.Exit(1)
    }

    service := &Service{srv}
    stdlog.Println("Starting kademlia storage node.")
    if status, err := service.Manage(); err != nil {
        errlog.Println("Error: ", err)
        os.Exit(1)
    } else {
        stdlog.Println(status)
    }
}

func (service *Service) Manage() (string, error) {
    if len(os.Args) > 1 {
        command := os.Args[1]
        switch command {
        case "install":
            return service.Install()
        case "remove":
            return service.Remove()
        case "start":
            return service.Start()
        case "stop":
            return service.Stop()
        case "status":
            return service.Status()
        default:
            return "Usage: kademliad install | remove | start | stop | status", nil
        }
    }

    interrupt := make(chan os.Signal, 1)
    signal.Notify(interrupt, os.Interrupt, os.Kill, syscall.SIGTERM)

    for {
        select {
        //TODO: add case(s) to actually do stuff
        case signal := <-interrupt:
            stdlog.Println("Got signal:", signal)
            if signal == os.Interrupt {
                return "Daemon was interrupted by system signal", nil
            }
            return "Daemon was killed.", nil
        }
    }
}
