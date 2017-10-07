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

    if len(contactsWithData) == 1 && contactsWithData[0].ID.Equals(k.Net.Routing.Me.ID) {
        // The data is in our KVStore
        data, _ = k.Net.Store.Lookup(*hashID)
        var content []kademlia.Contact
        if err := msgpack.Unmarshal(data, &content); err != nil {
            fmt.Println("Your data found locally:", string(data))
            sendResponse(w, http.StatusOK, string(data))
            return
        } else {
            panic("self-referential content in KVStore")
        }
    } else {
        // The data is elsewhere
        fmt.Println("Your data is in another castle")

            // We got a list of contacts
        for _, contact := range contactsWithData {
            fmt.Println("Candidate:", contact)
            downloadedData := k.Download(hashID, &contact)

            if len(downloadedData) > 0 {
                fmt.Println("Your data was downloaded remotely:", string(downloadedData))
                sendResponse(w, http.StatusOK, string(downloadedData))
                return
            }
        }
        fmt.Println("None of the contacts had the file.")
    }

    sendResponse(w, http.StatusNoContent, "")
}
