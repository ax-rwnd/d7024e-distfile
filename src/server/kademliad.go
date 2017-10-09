package main

import (
    "log"
    "os"
    "github.com/BurntSushi/toml"
    "time"
)

const config_file = "kademliad.toml"

var stdlog, errlog *log.Logger

// Pre-main initialization
func init() {
    stdlog = log.New(os.Stdout, "", 0)
    errlog = log.New(os.Stderr, "", 0)
}

type daemonConfig struct {
    Address              string
    TcpPort              int
    UdpPort              int
    RestPort             int
    Alpha                int
    ReplicationFactor    int
    BootAddr             string
    BootPort             int
    EvictionTime         time.Duration
    RepublishTime        time.Duration
    ConnectionTimeout    time.Duration
    ConnectionRetryDelay time.Duration
    ReceiveBufferSize    int
}

func main() {
    var config daemonConfig
    if _, err := toml.DecodeFile(config_file, &config); err != nil {
        errlog.Println("Error while parsing", config_file, ":", err)
        os.Exit(1)
    }

    // Grab ip from environment
    if config.Address == "detect" {
        stdlog.Println("Detecting address!")
        config.Address = os.Getenv("KADIP")
        stdlog.Println("New address", config.Address)
    }

    runDaemon(&config)
}
