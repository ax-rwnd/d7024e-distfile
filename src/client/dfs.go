package main

import (
    "os"
    "encoding/hex"
    "io/ioutil"
    "strings"
    "fmt"
    "errors"
    "net/http"
    "github.com/BurntSushi/toml"
    "kademlia"
    "log"
)

// Standard errors
var ArgumentError = errors.New("invalid arguments")
var HashError = errors.New("hash wrong length")

// Handle errors
func check (err error) bool {
    if err != nil {
        log.Fatal(err)
        return false
    }
    return true
}

// Config through TOML
const config_file = "dfs.toml"
var cConfig clientConfig

type clientConfig struct {
    Address     string
}

func init() {
    if _, err := toml.DecodeFile(config_file, &cConfig); err != nil {
        log.Fatal("Error while parsing", config_file, ":", err)
        os.Exit(1)
    }
}

func main () {
    args := os.Args[1:]

    if len(args) > 0 {
        if args[0] == "store" {
            r, err := handleStore(&cConfig, args[1:])
            check(err)
            print("Your hash is:",r)
        } else if args[0] == "cat" {
            content, err := handleCat(&cConfig, args[1:])
            print(content)
            check(err)
        } else if args[0] == "pin" {
            err := handlePin(&cConfig, args[1:])
            check(err)
        } else if args[0] == "unpin" {
            err := handleUnpin(&cConfig, args[1:])
            check(err)
        }
    } else {
        log.Fatal("Usage: dsf (store|cat|pin|unpin)")
    }
}

// Store data on client
func handleStore(config *clientConfig, args []string) (string, error) {
    if len(args) != 1 {
       check(ArgumentError)
    }

    // Load file
    fileReader, fileErr := os.Open(args[0])
    check(fileErr)

    // Perform request
    request := fmt.Sprintf("http://%s/store", config.Address)
    resp, requestErr := http.Post(request, "application/octet-stream", fileReader)
    check(requestErr)
    defer resp.Body.Close()

    // Read response body
    body, readErr := ioutil.ReadAll(resp.Body)
    check(readErr)

    // Report status to user
    if len(body) == kademlia.IDLength {
        return hex.EncodeToString(body), nil
    } else {
        return "", HashError
    }
}

func handleCat(config *clientConfig, args []string) (string, error) {
    if len(args) != 1 {
        check(ArgumentError)
    }

    if hash := args[0]; len(hash) != 2*kademlia.IDLength {
        return "", HashError
    } else {
        // Perform request
        request := fmt.Sprintf("http://%s/cat/%s", config.Address, hash)
        response, requestErr := http.Get(request)
        check(requestErr)

        // Read response
        body, readErr := ioutil.ReadAll(response.Body)
        check(readErr)
        defer response.Body.Close()
        return string(body), nil
    }
}

func handlePin(config *clientConfig, args []string) error {
    if len(args) != 1 {
        return ArgumentError
    } else {
        if hash := args[0]; len(hash) != 2*kademlia.IDLength {
            return HashError
        } else {
            request := fmt.Sprintf("%s/pin", config.Address)
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
            request := fmt.Sprintf("%s/unpin", config.Address, hash)
            http.Post(request, "text/plain", strings.NewReader(hash))
            return nil
        }
    }
}
