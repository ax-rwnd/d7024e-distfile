package main

import (
    "os"
    "strings"
    "fmt"
    "errors"
    "kademlia"
    "log"
    "encoding/hex"
    "net/http"
    "io/ioutil"
    "github.com/BurntSushi/toml"
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
            r := handleStore(&cConfig, args[1:])
            println("Your hash is:",r)
        } else if args[0] == "cat" {
            content := handleCat(&cConfig, args[1:])
            print(content)
        } else if args[0] == "pin" {
            status := handlePin(&cConfig, args[1:])
            if status == "SUCCESS" {
                print("The content was pinned.")
            } else {
                print("The operation failed.")
            }
        } else if args[0] == "unpin" {
            status := handleUnpin(&cConfig, args[1:])
            if status == "SUCCESS" {
                print("The content is no longer pinned.")
            } else {
                print("The operation failed.")
            }
        }
    } else {
        log.Fatal("Usage: dsf (store filename|cat hex-hash|pin hex-hash|unpin hex-hash)")
    }
}

// Store data on client
func handleStore(config *clientConfig, args []string) string {
    if len(args) != 1 {
       check(ArgumentError)
    }

    // Load file
    fileReader, fileErr := os.Open(args[0])
    check(fileErr)

    // Perform request
    request := fmt.Sprintf("http://%s/store", config.Address)
    response, requestErr := http.Post(request, "application/octet-stream", fileReader)
    check(requestErr)
    defer response.Body.Close()

    // Read response body
    body, readErr := ioutil.ReadAll(response.Body)
    check(readErr)

    // Fail if length mismatch
    if len(body) != kademlia.IDLength {
        print("body was:", body)
        check(HashError)
    }

    // Report status to user
    return hex.EncodeToString(body)
}

// Read data from network
func handleCat(config *clientConfig, args []string) string {
    if len(args) != 1 {
        check(ArgumentError)
    }

    // Fail if length mismatch
    hash := args[0]
    if len(hash) != 2*kademlia.IDLength {
        check(HashError)
    }

    // Perform request
    request := fmt.Sprintf("http://%s/cat/%s", config.Address, hash)
    response, requestErr := http.Get(request)
    check(requestErr)
    defer response.Body.Close()

    // Read response
    body, readErr := ioutil.ReadAll(response.Body)
    check(readErr)
    return string(body)
}

func handlePin(config *clientConfig, args []string) string {
    if len(args) != 1 {
        check(ArgumentError)
    }

    hash := args[0]
    if len(hash) != 2*kademlia.IDLength {
        check(HashError)
    }

    // Perform request
    request := fmt.Sprintf("http://%s/pin/%s", config.Address, hash)
    response, requestErr := http.Post(request, "text/plain", strings.NewReader("PIN"))
    check(requestErr)
    defer response.Body.Close()

    // Return response
    body, readErr := ioutil.ReadAll(response.Body)
    check(readErr)
    return string(body)
}

func handleUnpin(config *clientConfig, args []string) string {
    if len(args) != 1 {
        check(ArgumentError)
    }

    hash := args[0]
    if len(hash) != 2*kademlia.IDLength {
        check(HashError)
    }

    // Perform request
    request := fmt.Sprintf("http://%s/unpin/%s", config.Address, hash)
    response, requestErr := http.Post(request, "text/plain", strings.NewReader("UNPIN"))
    check(requestErr)
    defer response.Body.Close()

    // Return response
    body, readErr := ioutil.ReadAll(response.Body)
    check(readErr)
    return string(body)
}
