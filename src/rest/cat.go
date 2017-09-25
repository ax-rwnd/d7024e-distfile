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

    hashID := kademlia.NewKademliaID(hash)
    var data []byte

    contactsWithData := *k.LookupData(hashID)
    if len(contactsWithData) > 0 {
        if contactsWithData[0].ID.Equals(k.Net.Routing.Me.ID) {
            // We have the file locally
            data, _ = k.Net.Store.Lookup(*hashID)
        } else {
            // Someone else has the file
            for _, contact := range contactsWithData {
                data := k.Download(hashID, &contact)
                if len(data) > 0 {
                    break;
                }
            }
        }
    }
    sendResponse(w, http.StatusOK, string(data))
}
