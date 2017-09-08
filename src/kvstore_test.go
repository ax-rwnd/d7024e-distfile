package d7024e

import (
    "testing"
    "fmt"
    )

func TestInit(t *testing.T) {
    if err := KVSInit(); err != nil {
        fmt.Println("Initialization of KVStore failed.")
        t.Fail()
    }

    if err := KVSInit(); err == nil {
        fmt.Println("KVStore was silently reinitialized!")
        t.Fail()
    }
}
