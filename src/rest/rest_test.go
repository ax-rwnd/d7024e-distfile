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
    "strconv"
)

var testPort int = 7000

func getTestPort() int {
    testPort++
    return testPort
}

func TestRestBadRequest(t *testing.T) {
    k := kademlia.NewKademlia("127.0.0.1", getTestPort(), getTestPort())
    kRestPort := getTestPort()
    go Initialize(k, kRestPort)
    time.Sleep(time.Second)

    // This should be a POST
    resp, err := http.Get("http://localhost:" + strconv.Itoa(kRestPort) + "/store")
    if err != nil {
        log.Fatal(err)
        t.Fail()
    }
    if resp.StatusCode != http.StatusBadRequest {
        fmt.Println("Store accepted GET with ", strconv.Itoa(resp.StatusCode))
        t.Fail()
    }
    resp.Body.Close()

    // This should be a GET
    resp, err = http.Post("http://localhost:"+strconv.Itoa(kRestPort)+"/cat/0", "application/octet-stream", nil)
    if err != nil {
        log.Fatal(err)
        t.Fail()
    }
    if resp.StatusCode != http.StatusBadRequest {
        fmt.Println("Cat accepted POST with ", strconv.Itoa(resp.StatusCode))
        t.Fail()
    }
    resp.Body.Close()

    // This should be a GET
    resp, err = http.Post("http://localhost:"+strconv.Itoa(kRestPort)+"/dump", "application/octet-stream", nil)
    if err != nil {
        log.Fatal(err)
        t.Fail()
    }
    if resp.StatusCode != http.StatusBadRequest {
        fmt.Println("Cat accepted POST with ", strconv.Itoa(resp.StatusCode))
        t.Fail()
    }
    resp.Body.Close()

    // This should be a POST
    resp, err = http.Get("http://localhost:" + strconv.Itoa(kRestPort) + "/pin/0")
    if err != nil {
        log.Fatal(err)
        t.Fail()
    }
    if resp.StatusCode != http.StatusBadRequest {
        fmt.Println("Pin accepted GET with ", strconv.Itoa(resp.StatusCode))
        t.Fail()
    }
    resp.Body.Close()

    // This should be a POST
    resp, err = http.Get("http://localhost:" + strconv.Itoa(kRestPort) + "/unpin/0")
    if err != nil {
        log.Fatal(err)
        t.Fail()
    }
    if resp.StatusCode != http.StatusBadRequest {
        fmt.Println("Unpin accepted GET with ", strconv.Itoa(resp.StatusCode))
        t.Fail()
    }
    resp.Body.Close()
}

func TestRestPinUnpin(t *testing.T) {
    k := kademlia.NewKademlia("127.0.0.1", getTestPort(), getTestPort())
    kRestPort := getTestPort()
    go Initialize(k, kRestPort)
    time.Sleep(time.Second)

    // Open a file to store in server
    fileReader, err := os.Open("test.txt")
    if err != nil {
        log.Fatal(err)
        t.Fail()
    }
    bytesToStore, err := ioutil.ReadAll(fileReader)
    if err != nil {
        log.Fatal(err)
        t.Fail()
    }
    fileReader.Seek(0, 0)
    id := kademlia.NewKademliaIDFromBytes(bytesToStore)
    fmt.Print("Storing: ", id.String(), ": ", string(bytesToStore), "\n")

    // Send a store RPC for the file content
    resp, err := http.Post("http://localhost:"+strconv.Itoa(kRestPort)+"/store", "application/octet-stream", fileReader)
    if err != nil {
        log.Fatal(err)
        t.Fail()
    }
    fileReader.Close()
    resp.Body.Close()

    // Send a pin RPC for the stored file content
    resp, err = http.Post("http://localhost:"+strconv.Itoa(kRestPort)+"/pin/"+id.String(), "application/octet-stream", nil)

    if err != nil {
        log.Fatal(err)
        t.Fail()
    }
    if resp.StatusCode != http.StatusOK {
        fmt.Println(resp.StatusCode)
        t.Fail()
    }
    resp.Body.Close()

    // Send a unpin RPC for the stored file content
    resp, err = http.Post("http://localhost:"+strconv.Itoa(kRestPort)+"/unpin/"+id.String(), "application/octet-stream", nil)

    if err != nil {
        log.Fatal(err)
        t.Fail()
    }
    if resp.StatusCode != http.StatusOK {
        fmt.Println(resp.StatusCode)
        t.Fail()
    }
    resp.Body.Close()
    k.Net.Close()
}

func TestRestStoreCatLocal(t *testing.T) {
    k := kademlia.NewKademlia("127.0.0.1", getTestPort(), getTestPort())
    kRestPort := getTestPort()
    go Initialize(k, kRestPort)
    time.Sleep(time.Second)

    // Open a file to store in server
    fileReader, err := os.Open("test.txt")
    if err != nil {
        log.Fatal(err)
        t.Fail()
    }
    bytesToStore, err := ioutil.ReadAll(fileReader)
    if err != nil {
        log.Fatal(err)
        t.Fail()
    }
    fileReader.Seek(0, 0)
    id := kademlia.NewKademliaIDFromBytes(bytesToStore)
    fmt.Print("Storing: ", id.String(), ": ", string(bytesToStore), "\n")

    // Send a store RPC for the file content
    resp, err := http.Post("http://localhost:"+strconv.Itoa(kRestPort)+"/store", "application/octet-stream", fileReader)
    if err != nil {
        log.Fatal(err)
        t.Fail()
    }
    fileReader.Close()
    resp.Body.Close()

    time.Sleep(time.Second)

    // Send a cat RPC for the stored file content
    resp, err = http.Get("http://localhost:" + strconv.Itoa(kRestPort) + "/cat/" + id.String())
    if err != nil {
        log.Fatal(err)
        t.Fail()
    }
    bytesFromCat, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        log.Fatal(err)
        t.Fail()
    }
    resp.Body.Close()

    // Check that we got back what we put in
    if len(bytesFromCat) != len(bytesToStore) || len(bytesFromCat) == 0 {
        log.Println("Invalid length")
        t.Fail()
    }
    for i := range bytesFromCat {
        if bytesFromCat[i] != bytesToStore[i] {
            t.Fail()
        }
    }
    k.Net.Close()
}

func TestRestStoreCatRemote(t *testing.T) {
    // Create two Kademlias
    k1 := kademlia.NewKademlia("127.0.0.1", getTestPort(), getTestPort())
    k1RestPort := getTestPort()
    go Initialize(k1, k1RestPort)
    k2 := kademlia.NewKademlia("127.0.0.1", getTestPort(), getTestPort())
    k2RestPort := getTestPort()
    go Initialize(k2, k2RestPort)

    // Bootstrap one to the other
    time.Sleep(time.Second)
    k2.Bootstrap("127.0.0.1", k1.Net.Routing.Me.Address.TcpPort, k1.Net.Routing.Me.Address.UdpPort)

    // Open a file to store
    fileReader, err := os.Open("test.txt")
    if err != nil {
        log.Fatal(err)
        t.Fail()
    }
    bytesToStore, err := ioutil.ReadAll(fileReader)
    if err != nil {
        log.Fatal(err)
        t.Fail()
    }
    fileReader.Seek(0, 0)
    id := kademlia.NewKademliaIDFromBytes(bytesToStore)
    fmt.Print("Storing: ", id.String(), ": ", string(bytesToStore), "\n")

    // Send a store RPC for the file content to one server
    resp, err := http.Post("http://localhost:"+strconv.Itoa(k1RestPort)+"/store", "application/octet-stream", fileReader)
    if err != nil {
        log.Fatal(err)
        t.Fail()
    }
    fileReader.Close()
    resp.Body.Close()

    time.Sleep(time.Second)

    // Send a cat RPC for the stored file content to the other server
    resp, err = http.Get("http://localhost:" + strconv.Itoa(k2RestPort) + "/cat/" + id.String())
    if err != nil {
        log.Fatal(err)
        t.Fail()
    }
    bytesFromCat, err := ioutil.ReadAll(resp.Body)
    resp.Body.Close()

    // Check that we got back what we put in
    if len(bytesFromCat) != len(bytesToStore) || len(bytesFromCat) == 0 {
        log.Println("Invalid length")
        t.Fail()
    }
    for i := range bytesFromCat {
        if bytesFromCat[i] != bytesToStore[i] {
            t.Fail()
        }
    }
    k1.Net.Close()
    k2.Net.Close()
}

func TestRestDump(t *testing.T) {
    k := kademlia.NewKademlia("127.0.0.1", getTestPort(), getTestPort())
    kRestPort := getTestPort()
    go Initialize(k, kRestPort)
    time.Sleep(time.Second)

    // Open a file to store in server
    fileReader, err := os.Open("test.txt")
    if err != nil {
        log.Fatal(err)
        t.Fail()
    }
    bytesToStore, err := ioutil.ReadAll(fileReader)
    if err != nil {
        log.Fatal(err)
        t.Fail()
    }
    fileReader.Seek(0, 0)
    id := kademlia.NewKademliaIDFromBytes(bytesToStore)
    fmt.Print("Storing: ", id.String(), ": ", string(bytesToStore), "\n")

    // Send a store RPC for the file content
    resp, err := http.Post("http://localhost:"+strconv.Itoa(kRestPort)+"/store", "application/octet-stream", fileReader)
    if err != nil {
        log.Fatal(err)
        t.Fail()
    }
    fileReader.Close()
    resp.Body.Close()

    // Send a pin RPC for the stored file content
    resp, err = http.Get("http://localhost:" + strconv.Itoa(kRestPort) + "/contacts")

    if err != nil {
        log.Fatal(err)
        t.Fail()
    }
    if resp.StatusCode != http.StatusOK {
        fmt.Println(resp.StatusCode)
        t.Fail()
    }
    resp.Body.Close()

    // Send a unpin RPC for the stored file content
    resp, err = http.Get("http://localhost:" + strconv.Itoa(kRestPort) + "/dump")

    if err != nil {
        log.Fatal(err)
        t.Fail()
    }
    if resp.StatusCode != http.StatusOK {
        fmt.Println(resp.StatusCode)
        t.Fail()
    }
    resp.Body.Close()
    k.Net.Close()
}
