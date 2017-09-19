package main

import (
    "os"
    "io/ioutil"
    "strings"
    "fmt"
    "errors"
    "net/http"
    "github.com/BurntSushi/toml"
    "kademlia"
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

    if len(args) > 0 {
        if args[0] == "store" {
            var r string
            r, err = handleStore(&config, args[1:])
            fmt.Println("Your hash is:",r)
        } else if args[0] == "cat" {
            _, err = handleCat(&config, args[1:])
        } else if args[0] == "pin" {
            err = handlePin(&config, args[1:])
        } else if args[0] == "unpin" {
            err = handleUnpin(&config, args[1:])
        }
    }

    if len(args) <= 0 {
        fmt.Println("Too few args.")
        os.Exit(1)
    } else if err != nil {
        fmt.Println("Error:", err)
        os.Exit(1)
    }
}

func handleStore(config *clientConfig, args []string) (r string, err error) {
    if len(args) != 1 {
        err = ArgumentError
        return
    } else {
        fileName := args[0]
        var request = fmt.Sprintf("http://%s/store", config.Address)
        var fileReader *os.File
        var resp *http.Response


        // Send file data
        fileReader, err = os.Open(fileName)
        if err != nil {
            return
        }
        resp, err = http.Post(request, "application/octet-stream", fileReader)
        if err != nil {
            return
        } else {
            var body []byte
            body, err = ioutil.ReadAll(resp.Body)
            r = string(body)
            fmt.Println("Your file was stored: ", r)
            resp.Body.Close()
        }
        return
    }
}

func handleCat(config *clientConfig, args []string) (r string, err error) {
    if len(args) != 1 {
        err = ArgumentError
        return
    } else {
        if hash := args[0]; len(hash) != 2*kademlia.IDLength {
            err = HashError
            return
        } else {
            var response *http.Response

            var request = fmt.Sprintf("http://%s/cat/%s", config.Address, hash)
            response, err = http.Get(request)
            if err != nil {
                return
            }

            var body []byte
            body, err = ioutil.ReadAll(response.Body)
            response.Body.Close()
            r = string(body)

            return
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
