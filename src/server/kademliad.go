package main

import (
    "log"
    "os"
    "github.com/BurntSushi/toml"
)

const config_file = "kademliad.toml"

var stdlog, errlog *log.Logger

// Pre-main initialization
func init() {
    stdlog = log.New(os.Stdout, "", 0)
    errlog = log.New(os.Stderr, "", 0)
}

type daemonConfig struct {
    Address     string
    TcpPort     int
    UdpPort     int
    RestPort    int
    Alpha       int
    Replication int
}

func main() {
    var config daemonConfig
    if _, err := toml.DecodeFile(config_file, &config); err != nil {
        errlog.Println("Error while parsing", config_file, ":", err)
        os.Exit(1)
    }

    runDaemon(&config)
}
