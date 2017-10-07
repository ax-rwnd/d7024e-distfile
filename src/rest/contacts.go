package rest

import (
    "net/http"
    "kademlia"
    "encoding/json"
)

func contactsHandler(k *kademlia.Kademlia, w http.ResponseWriter, r *http.Request) {
    var contacts []kademlia.Contact
    for i := 0; i<kademlia.IDLength*8; i++ {
        in_bucket := k.Net.Routing.GetBucket(i)
        from_bucket := in_bucket.DumpContacts()
        contacts = append(contacts, from_bucket...)
    }

    if to_return, err := json.Marshal(contacts); err != nil {
        sendResponse(w, 500, "")
    } else {
        sendResponse(w, http.StatusOK, string(to_return))
    }
}
