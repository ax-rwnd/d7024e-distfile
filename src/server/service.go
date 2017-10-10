package main

import (
    "os/signal"
    "github.com/takama/daemon"
    "os"
    "syscall"
    "kademlia"
    "rest"
)

var dependencies = []string{"dummy.service"}

type Service struct {
    daemon.Daemon
}

func runDaemon(config *daemonConfig) {
    srv, err := daemon.New("kademliad", "Kademlia Storage Daemon", dependencies...)
    if err != nil {
        errlog.Println("Error: ", err)
        os.Exit(1)
    }

    service := &Service{srv}
    stdlog.Println("Starting kademlia storage node.")
    if status, err := service.Manage(config); err != nil {
        errlog.Println("Error: ", err)
        os.Exit(1)
    } else {
        stdlog.Println(status)
    }
}

func (service *Service) Manage(config *daemonConfig) (string, error) { // Start upp rest server and rpc server here, send the config file
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

    if config.Alpha < 1 {
        panic("Invalid alpha value")
    }
    if config.ReplicationFactor < 1 {
        panic("Invalid replication factor")
    }
    if config.RepublishTime < 1 {
        panic("Invalid republish time")
    }
    if config.EvictionTime < 1 {
        panic("Invalid eviction time")
    }
    if config.ReceiveBufferSize < 1024 {
        panic("Receive buffer size too small")
    }
    if config.ConnectionTimeout < 0 {
        panic("Invalid connection timeout")
    }
    if config.ConnectionRetryDelay < 0 {
        panic("Invalid connection retry timeout")
    }
    if config.RestPort < 1 || config.UdpPort < 1 || config.TcpPort < 1 {
        panic("Invalid port setting")
    }
    if len(config.Address) == 0 || len(config.BootAddr) == 0 {
        panic("Invalid IP address setting")
    }
    kademlia.Alpha = config.Alpha
    kademlia.ReplicationFactor = config.ReplicationFactor
    kademlia.ConnectionTimeout = config.ConnectionTimeout
    kademlia.ConnectionRetryDelay = config.ConnectionRetryDelay
    kademlia.ReceiveBufferSize = config.ReceiveBufferSize
    kademlia.EvictionTime = config.EvictionTime
    kademlia.RepublishTime = config.RepublishTime

    k := kademlia.NewKademlia(config.Address, config.TcpPort, config.UdpPort)
    k.Bootstrap(config.BootAddr, config.TcpPort, config.BootPort)

    go rest.Initialize(k, config.RestPort)
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
