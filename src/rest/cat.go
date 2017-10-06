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
    fmt.Println("Contacts with data:", contactsWithData)
    if len(contactsWithData) > 0 {
        if contactsWithData[0].ID.Equals(k.Net.Routing.Me.ID) {
            // We have the file locally
            data, _ = k.Net.Store.Lookup(*hashID)
            fmt.Println("Your data found locally:", string(data))
        } else {
            // Someone else has the file
            for _, contact := range contactsWithData {
                data := k.Download(hashID, &contact)
                fmt.Println("Your data was downloaded remotely:", string(data))
                if len(data) > 0 {
                    break;
                }
            }
        }

        sendResponse(w, http.StatusOK, string(data))
    } else {
        sendResponse(w, http.StatusNoContent,"")
    }
}
