package main

import (
    "./kademlia"
    "fmt"
    "github.com/BurntSushi/toml"
)

type tomlConfig struct {
    Kademlia kademliaConf
    Host    hostConf
}

type kademliaConf struct {
   Replication_factor int
}

type hostConf struct {
    Address string
    Port int
}

func main() {
    var config tomlConfig
    if _, err := toml.DecodeFile("config.toml", &config); err != nil {
        fmt.Println("Error parsing config: ", err)
    }

    a := kademlia.NewNode(true)
    fmt.Println(a)
}
