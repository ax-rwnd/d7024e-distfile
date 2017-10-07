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
    if err := k.Net.Store.Unpin(*h); err == kademlia.NotFoundError {
        sendResponse(w, http.StatusNotFound, fmt.Sprintf("%s could not be found.", hash))
    } else if err != nil {
        sendResponse(w, 500, fmt.Sprintf("%s could't be unpinned: %s.", hash, err))
    } else {
        sendResponse(w, http.StatusOK, fmt.Sprintf("%s was unpinned.", hash))
    }
}
