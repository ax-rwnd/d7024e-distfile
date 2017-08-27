package kademlia

import "testing"

func TestID(t *testing.T) {
    id := new(Id_t)
    for i := 0; i<20; i++ {
        if id[i] != 0 {
            t.Errorf("Initialized var is not zero.")
        }
    }
}
