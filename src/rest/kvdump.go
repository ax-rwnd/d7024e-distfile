package rest

import (
    "kademlia"
    "net/http"
    "encoding/json"
)

func dumpStoreHandler(k *kademlia.Kademlia, w http.ResponseWriter, r *http.Request) {
    if r.Method != "GET" {
        sendResponse(w, http.StatusBadRequest, "400 - Not a GET request")
        return
    }

    data := k.Net.Store.DumpStore()
    if to_return, err := json.Marshal(data); err != nil {
        sendResponse(w, 500, "")
    } else {
        sendResponse(w, http.StatusOK, string(to_return))
    }
}
