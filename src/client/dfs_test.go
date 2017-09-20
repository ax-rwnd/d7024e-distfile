package main

import (
    "time"
    "testing"
    "fmt"
    "os"
    "bytes"
    "net/http"
    "io/ioutil"
    "crypto/sha1"
    "kademlia"
    "github.com/gorilla/mux"
    "github.com/BurntSushi/toml"
    "io"
    "encoding/hex"
)

var testConfig = "test_config.toml"
var config clientConfig

// Always begin by loading config
func init() {
    if _, err := toml.DecodeFile("test_config.toml", &config); err != nil {
        fmt.Println("Error while parsing", config_file, ":", err)
    }
}

// Handle CAT requests
func testHandleCat(w http.ResponseWriter, r *http.Request) {
    req := mux.Vars(r)
    hash := req["hash"]
    hex_hash, _ := hex.DecodeString(hash)
    expected, _ := hex.DecodeString("DEADBEEFDEADBEEFDEADBEEFDEADBEEFDEADBEEF")

    if r.Method == "GET" {
        if bytes.Equal(hex_hash, expected) {
            w.Write([]byte("CORRECT_HASH"))
        } else {
            w.Write([]byte("WRONG_HASH"))
        }
    } else {
        w.Write([]byte("NOT_GET"))
    }
}

// Handle STORE requests
func testHandleStore(w http.ResponseWriter, r *http.Request) {
    if r.Method == "POST" {
        b, _ := ioutil.ReadAll(r.Body)
        sum := sha1.Sum(b)
        w.Write(sum[:kademlia.IDLength])
    } else {
        w.Write([]byte("NOT_POST"))
    }
}

// Serve dummy endpoints for testing
func serveTestEndpoints(config *clientConfig) {
    router := mux.NewRouter()
    router.HandleFunc("/cat/{hash}", testHandleCat)
    router.HandleFunc("/store", testHandleStore)
    /*
    router.HandleFunc("/pin/{hash}", pinHandler)
    router.HandleFunc("/unpin/{hash}", unpinHandler)*/
    http.ListenAndServe(config.Address, router)
}

// Serve a fixed hash over network, make sure it's the requested one
func TestCat (t *testing.T) {
    // Start server
    go serveTestEndpoints(&config)
    time.Sleep(1*time.Second)

    // Request the fixed hash
    var args = []string{"DEADBEEFDEADBEEFDEADBEEFDEADBEEFDEADBEEF"}
    response, _ := handleCat(&config, args)

    // Test the returned value
    if response == "NOT_GET" {
        fmt.Println("Cat should GET.")
        t.Fail()
    } else if "CORRECT_HASH" != response {
        fmt.Println("Exptected",args[0], "got", response)
        t.Fail()
    }
}

// Serve a hash over network and make sure it's delivered properly
func TestStore(t *testing.T) {
    // Start server
    go serveTestEndpoints(&config)
    time.Sleep(1*time.Second)

    // Perform request
    var args = []string{"test.html"}
    response, _ := handleStore(&config, args)

    file, _ := os.Open("test.html")
    content, _ := ioutil.ReadAll(file)
    defer file.Close()

    // Dump content of file into sha1 hasher
    h := sha1.New()
    if _, err := io.Copy(h, file); err != nil {
        fmt.Println("Failed to copy from the file.")
        t.Fail()
    }
    io.WriteString(h, string(content))
    expected_hash := h.Sum(nil)
    hexString := hex.EncodeToString(expected_hash[:kademlia.IDLength])

    // Make sure that the response is the same as the hasher
    if response != hexString {
        fmt.Println("Expected", hexString,"got",response)
        t.Fail()
    }
}
