package rest

import (
    "net/http"
    "github.com/gorilla/mux"
    "fmt"
    "kademlia"
)

func catHandler(k *kademlia.Kademlia, w http.ResponseWriter, r *http.Request) {
    req := mux.Vars(r)
    hash := req["hash"]

    if r.Method != "GET" {
        sendResponse(w, http.StatusBadRequest, "400 - Not a GET request")
        return
    }

    fmt.Println(hash)

    sendResponse(w, http.StatusOK, "200 - OK ")
}
