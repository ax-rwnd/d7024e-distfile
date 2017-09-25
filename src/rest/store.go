package rest

import (
    "net/http"
    "kademlia"
    "io/ioutil"
    "log"
)

// Store data to KVStore and calculate the hash with SHA-1
func storeHandler(k *kademlia.Kademlia, w http.ResponseWriter, r *http.Request) {
    // Not a POST
    if r.Method != "POST" {
        sendResponse(w, http.StatusBadRequest, "400 - Not a POST request")
        return
    }

    // Read request body
    data, err := ioutil.ReadAll(r.Body) // here, get the data in byte array
    if err != nil {
        log.Fatal(err)
        sendResponse(w, http.StatusInternalServerError, "500 - Couldn't read body")
        return
    }
    defer r.Body.Close()

    // Store data
    hash := k.Store(data)
    // Respond with hash
    sendResponse(w, http.StatusOK, string(hash[:kademlia.IDLength]))
}
