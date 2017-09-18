package main

import (
    "os"
    "strings"
    "fmt"
    "errors"
    "net/http"
    "github.com/BurntSushi/toml"
    "../kademlia"
)

var ArgumentError = errors.New("invalid arguments")
var HashError = errors.New("hash wrong length")

const config_file = "dfs.toml"

type clientConfig struct {
    Address     string
}

func printHelp(args []string) {
    fmt.Println(args, "\nUsage: ...")
}

func main () {
    args := os.Args[1:]

    var config clientConfig
    if _, err := toml.DecodeFile(config_file, &config); err != nil {
        fmt.Println("Error while parsing", config_file, ":", err)
        os.Exit(1)
    }

    var err error

    if len(args) < 1 {
        printHelp(args)
    } else if args[0] == "store" {
        err = handleStore(&config, args[1:])
    } else if args[0] == "cat" {
        err = handleCat(&config, args[1:])
    } else if args[0] == "pin" {
        err = handlePin(&config, args[1:])
    } else if args[0] == "unpin" {
        err = handleUnpin(&config, args[1:])
    }

    if err != nil {
        fmt.Println(err)
    }
}

func handleStore(config *clientConfig, args []string) error {
    if len(args) != 1 {
        return ArgumentError
    } else {
        fileName := args[0]
        var request = fmt.Sprintf("%s/store", config.Address)

        // Send file data
        if fileReader, err := os.Open(fileName); err != nil {
            return err
        } else if resp, err := http.Post(request, "application/octet-stream", fileReader); err != nil {
            return err
        } else {
            defer resp.Body.Close()
            fmt.Println("Your file was stored: ", resp.Body)
        }

        return nil
    }
}

func handleCat(config *clientConfig, args []string) error {
    if len(args) != 1 {
        return ArgumentError
    } else {
        if hash := args[0]; len(hash) != 2*kademlia.IDLength {
            return HashError
        } else {
            var request = fmt.Sprintf("%s/cat?hash=%s", config.Address, hash)
            if response, err := http.Get(request); err != nil {
                fmt.Println("Response from server:", response)
            }

            return nil
        }
    }
}

func handlePin(config *clientConfig, args []string) error {
    if len(args) != 1 {
        return ArgumentError
    } else {
        if hash := args[0]; len(hash) != 2*kademlia.IDLength {
            return HashError
        } else {
            var request = fmt.Sprintf("%s/pin", config.Address)
            http.Post(request, "text/plain", strings.NewReader(hash))
            return nil
        }
    }
}

func handleUnpin(config *clientConfig, args []string) error {
    if len(args) != 1 {
        return ArgumentError
    } else {
        if hash := args[0]; len(hash) != 2*kademlia.IDLength {
            return HashError
        } else {
            var request = fmt.Sprintf("%s/unpin", config.Address, hash)
            http.Post(request, "text/plain", strings.NewReader(hash))
            return nil
        }
    }
}
