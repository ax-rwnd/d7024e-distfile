package main

import(
	"os"
	"fmt"
	"net/http"
	"io/ioutil"
	"log"
)

func main() {

// Send a simple post request to port :8080
// Sending a request

resp, err := http.Get("http://localhost:8080/cat/testing")

if err != nil {
	// handle error
}

defer resp.Body.Close()
body, err := ioutil.ReadAll(resp.Body)

fmt.Println(string(body))

nextResp()

}

func nextResp() {

fileReader, err := os.Open("C:/Users/LiNn/d7024e-distfile/src/rest/client/test.txt")

if err != nil {
	log.Fatal(err)
}

resp, err := http.Post("http://localhost:8080/store", "application/octet-stream", fileReader)

if err != nil {
	log.Fatal(err)
}

defer resp.Body.Close()

body, err := ioutil.ReadAll(resp.Body)

fmt.Println(string(body))

}