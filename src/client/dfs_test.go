package main

import (
    "time"
    "github.com/BurntSushi/toml"
    "testing"
    "fmt"
    "net/http"
    "github.com/gorilla/mux"
)

var testConfig = "test_config.toml"

func testHandleCat(w http.ResponseWriter, r *http.Request) {
    req := mux.Vars(r)
    hash := req["hash"]
    if r.Method == "GET" {
        w.Write([]byte(hash))
    } else {
        w.Write([]byte("NOT_GET"))
    }
}

func serveTestEndpoints(config *clientConfig) {
    router := mux.NewRouter()
    router.HandleFunc("/cat/{hash}", testHandleCat)
    /*
    router.HandleFunc("/store/{file}", storeHandler)
    router.HandleFunc("/pin/{hash}", pinHandler)
    router.HandleFunc("/unpin/{hash}", unpinHandler)*/
    http.ListenAndServe(config.Address, router)
}

func TestCat (t *testing.T) {
    var config clientConfig
    if _, err := toml.DecodeFile("test_config.toml", &config); err != nil {
        fmt.Println("Error while parsing", config_file, ":", err)
        t.Fail()
    }

    go serveTestEndpoints(&config)
    time.Sleep(1*time.Second)
    var args = []string{"DEADBEEFDEADBEEFDEADBEEFDEADBEEFDEADBEEF"}
    response, _ := handleCat(&config, args)
    if args[0] != response {
        fmt.Println("Exptected",args[0], "got", response)
        t.Fail()
    }
}
