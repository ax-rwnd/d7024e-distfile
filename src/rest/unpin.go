package rest

import (
    "net/http"
    "github.com/gorilla/mux"
    "fmt"
    "kademlia"
)

func unpinHandler(k *kademlia.Kademlia, w http.ResponseWriter, r *http.Request) {
    req := mux.Vars(r)
    hash := req["hash"]
    if r.Method != "POST" {
        sendResponse(w, http.StatusBadRequest, "400 - Not a POST request")
        return
    }

    h := kademlia.NewKademliaID(hash)
    k.Net.Store.Unpin(*h)

    sendResponse(w, http.StatusOK, fmt.Sprintf("%s was unpinned.", hash))
}
