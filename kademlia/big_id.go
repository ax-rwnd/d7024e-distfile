/*
* 160-bit ID Storage type
*/

package kademlia

import "crypto/rand"

type Id_t [20]byte

func distance (a, b Id_t) (c Id_t) {
    for i := 0; i<20; i++ {
        c[i] = a[i]^b[i]
    }
    return
}

func NewId(random bool) (output Id_t) {
    if !random {
        output = *new(Id_t)
    } else {
        val := make([]byte, 20)
        rand.Read(val)
        for i := 0; i<20; i++ {
            output[i] = val[i]
        }
    }
    return
}
