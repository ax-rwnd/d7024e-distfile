package rest

import (
    "net/http"
    "github.com/gorilla/mux"
    "kademlia"
    "strconv"
)

func Initialize(k *kademlia.Kademlia, restPort int) {
    router := mux.NewRouter()
    router.HandleFunc("/cat/{hash}", func(w http.ResponseWriter, r *http.Request) { catHandler(k, w, r) })     // cat.go
    router.HandleFunc("/store", func(w http.ResponseWriter, r *http.Request) { storeHandler(k, w, r) })        // store.go
    router.HandleFunc("/contacts", func(w http.ResponseWriter, r *http.Request) { contactsHandler(k, w, r) })        // store.go
    router.HandleFunc("/dump", func(w http.ResponseWriter, r *http.Request) { dumpStoreHandler(k, w, r) })        // store.go
    router.HandleFunc("/pin/{hash}", func(w http.ResponseWriter, r *http.Request) { pinHandler(k, w, r) })     // pin.go
    router.HandleFunc("/unpin/{hash}", func(w http.ResponseWriter, r *http.Request) { unpinHandler(k, w, r) }) // unpin.go
    http.ListenAndServe(":"+strconv.Itoa(restPort), router)                                                    // fix so take port from config file
    // could use log.Fatal here, prints the error but then uses os.exit
}
