package rest

import (
    "net/http"
    "github.com/gorilla/mux"
    "fmt"
    "kademlia"
)

func pinHandler(k *kademlia.Kademlia, w http.ResponseWriter, r *http.Request) {
    req := mux.Vars(r)
    hash := req["hash"]

    if r.Method != "PUT" {
        sendResponse(w, http.StatusBadRequest, "400 - Not a PUT request")
        return
    }

    fmt.Println(hash)

    sendResponse(w, http.StatusOK, "200 - OK ")
}
