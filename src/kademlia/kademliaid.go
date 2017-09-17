package kademlia

import (
    "encoding/hex"
    "math/rand"
    "crypto/sha1"
)

const IDLength = 20

type KademliaID [IDLength]byte

func NewKademliaIDRandom() *KademliaID {
    random := make([]byte, IDLength)
    rand.Read(random)
    newKademliaID := KademliaID{}
    for i := 0; i < IDLength; i++ {
        newKademliaID[i] = random[i]
    }
    return &newKademliaID
}

func NewKademliaID(data string) *KademliaID {
    decoded, _ := hex.DecodeString(data)

    newKademliaID := KademliaID{}
    for i := 0; i < IDLength; i++ {
        newKademliaID[i] = decoded[i]
    }

    return &newKademliaID
}

func NewKademliaIDFromBytes(data []byte) *KademliaID {
    result := KademliaID{}
    hash := sha1.Sum(data)
    for i := IDLength - 1; i >= 0; i-- {
        // SHA1 sum is always 160 bits, IDLength might not be?
        result[i] = hash[i]
    }
    return &result
}

func NewRandomKademliaID() *KademliaID {
    newKademliaID := KademliaID{}
    for i := 0; i < IDLength; i++ {
        newKademliaID[i] = uint8(rand.Intn(256))
    }
    return &newKademliaID
}

func (kademliaID KademliaID) Less(otherKademliaID *KademliaID) bool {
    for i := 0; i < IDLength; i++ {
        if kademliaID[i] != otherKademliaID[i] {
            return kademliaID[i] < otherKademliaID[i]
        }
    }
    return false
}

func (kademliaID KademliaID) Equals(otherKademliaID *KademliaID) bool {
    for i := 0; i < IDLength; i++ {
        if kademliaID[i] != otherKademliaID[i] {
            return false
        }
    }
    return true
}

func (kademliaID KademliaID) CalcDistance(target *KademliaID) *KademliaID {
    result := KademliaID{}
    for i := 0; i < IDLength; i++ {
        result[i] = kademliaID[i] ^ target[i]
    }
    return &result
}

func (kademliaID *KademliaID) String() string {
    return hex.EncodeToString(kademliaID[0:IDLength])
}
