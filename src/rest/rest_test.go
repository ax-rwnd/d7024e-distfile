package rest

import (
    "os"
    "net/http"
    "log"
    "testing"
    "kademlia"
    "time"
    "fmt"
    "io/ioutil"
)

func TestRestStoreCat(t *testing.T) {
    k := kademlia.NewKademlia("localhost", 9000, 9001)
    go Initialize(k, 9002)
    time.Sleep(time.Second)

    // Open a file to store in server
    fileReader, err := os.Open("test.txt")
    if err != nil {
        log.Fatal(err)
    }
    bytesToStore, err := ioutil.ReadAll(fileReader)
    if err != nil {
        log.Fatal(err)
    }
    fileReader.Seek(0, 0)
    id := kademlia.NewKademliaIDFromBytes(bytesToStore)
    fmt.Print("Storing: ", id.String(), ": ", string(bytesToStore), "\n")

    // Send a store RPC for the file content
    resp, err := http.Post("http://localhost:9002/store", "application/octet-stream", fileReader)
    if err != nil {
        log.Fatal(err)
    }
    resp.Body.Close()

    // Send a cat RPC for the stored file content
    resp, err = http.Get("http://localhost:9002/cat/" + id.String())
    if err != nil {
        log.Fatal(err)
    }
    bytesFromCat, err := ioutil.ReadAll(resp.Body)
    resp.Body.Close()

    // Check that we got back what we put in
    if len(bytesFromCat) != len(bytesToStore) {
        log.Println("Invalid length")
        t.Fail()
    }
    for i := range bytesFromCat {
        if bytesFromCat[i] != bytesToStore[i] {
            t.Fail()
        }
    }
}
