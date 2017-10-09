package main

import (
    "strconv"
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
	BootAddr	string
	BootPort	int
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

    if config.BootAddr == "detect" {
        stdlog.Println("Detecting bootstrap address!")
		config.BootAddr = os.Getenv("BOOTIP")
        stdlog.Println("New bootstrap address", config.Address)
    }

    if config.BootPort == -1 {
        stdlog.Println("Detecting bootstrap port!")
        if newPort, err := strconv.Atoi(os.Getenv("BOOTPORT")); err == nil {
            config.BootPort = newPort
            stdlog.Println("New bootstrap port", config.BootPort)
        } else {
            errlog.Println("Failed setting new port", err)
        }
    }

    runDaemon(&config)
}
