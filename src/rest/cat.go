package rest

import (
    "net/http"
    "github.com/gorilla/mux"
    "github.com/vmihailenco/msgpack"
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
        if len(contactsWithData) == 1 && contactsWithData[0].ID.Equals(k.Net.Routing.Me.ID) {
            // We have the file locally
            data, _ = k.Net.Store.Lookup(*hashID)
            var content []kademlia.Contact
            if err := msgpack.Unmarshal(data, &content); err != nil {
                fmt.Println("Your data found locally:", string(data))
                sendResponse(w, http.StatusOK, string(data))
                return
            } else {
                fmt.Println("Your data is in another castle", content)
                for _, contact := range content {
                    downloadedData := k.Download(hashID, &contact)
                    if len(data) > 0 {
                        fmt.Println("Your data was downloaded remotely:", string(downloadedData))
                        sendResponse(w, http.StatusOK, string(downloadedData))
                        return
                    }
                }
            }
        }
    }
    sendResponse(w, http.StatusNoContent, "")
}
