package main

import (
    "fmt"
    "./kademlia"
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

//Post-config setup, inject settings and such
func post_config(config tomlConfig) {
    kademlia.Replication_factor = config.Kademlia.Replication_factor
}

func main() {
    var config tomlConfig
    if _, err := toml.DecodeFile("config.toml", &config); err != nil {
        fmt.Println("Error parsing config: ", err)
    }

    hostNode := kademlia.NewNode(true)
    fmt.Println("Running with node:", hostNode)
}
