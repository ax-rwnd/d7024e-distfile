package main

import (
    "time"
    "testing"
    "fmt"
    "os"
    "net/http"
    "io/ioutil"
    "github.com/gorilla/mux"
    "github.com/BurntSushi/toml"
)

var testConfig = "test_config.toml"
var config clientConfig

func init() {
    if _, err := toml.DecodeFile("test_config.toml", &config); err != nil {
        fmt.Println("Error while parsing", config_file, ":", err)
    }
}

func testHandleCat(w http.ResponseWriter, r *http.Request) {
    req := mux.Vars(r)
    hash := req["hash"]
    if r.Method == "GET" {
        w.Write([]byte(hash))
    } else {
        w.Write([]byte("NOT_GET"))
    }
}

func testHandleStore(w http.ResponseWriter, r *http.Request) {
    if r.Method == "POST" {
        b, _ := ioutil.ReadAll(r.Body)
        w.Write(b)
    } else {
        w.Write([]byte("NOT_POST"))
    }
}
func serveTestEndpoints(config *clientConfig) {
    router := mux.NewRouter()
    router.HandleFunc("/cat/{hash}", testHandleCat)
    router.HandleFunc("/store", testHandleStore)
    /*
    router.HandleFunc("/pin/{hash}", pinHandler)
    router.HandleFunc("/unpin/{hash}", unpinHandler)*/
    http.ListenAndServe(config.Address, router)
}

func TestCat (t *testing.T) {
    go serveTestEndpoints(&config)
    time.Sleep(1*time.Second)

    var args = []string{"DEADBEEFDEADBEEFDEADBEEFDEADBEEFDEADBEEF"}
    response, _ := handleCat(&config, args)
    if response == "NOT_GET" {
        fmt.Println("Cat should GET.")
        t.Fail()
    } else if args[0] != response {
        fmt.Println("Exptected",args[0], "got", response)
        t.Fail()
    }
}

func TestStore(t *testing.T) {
    go serveTestEndpoints(&config)
    time.Sleep(1*time.Second)

    var args = []string{"test.html"}
    response, _ := handleStore(&config, args)

    file, _ := os.Open("test.html")
    content, _ := ioutil.ReadAll(file)
    if response != string(content) {
        fmt.Println("Expected",content,"got",response)
        t.Fail()
    }
}
