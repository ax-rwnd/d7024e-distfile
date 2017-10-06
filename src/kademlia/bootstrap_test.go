package kademlia

import (
    "testing"
    "rpc"
    "log"
)

func imin(a, b int) int {
    if a < b {
        return a
    }
    return b
}

func TestBootstrap(t *testing.T) {
    // Crete a star topology
    NewKademlia("127.0.0.1", 4000, 4001)
    nodes := 9

    k := make([]*Kademlia, nodes)
    for i := 0; i<len(k); i++ {
        k[i] = NewKademlia("127.0.0.1", getTestPort(), getTestPort())
    }

    for i := 0; i<len(k); i++ {
       k[i].Bootstrap("127.0.0.1", 4000, 4001)
    }

    // Test finding another node
    var cc = []chan []Contact{make(chan []Contact), make(chan []Contact),}

    go func() {
        cc[0] <- k[0].LookupContact(k[len(k)-1].Net.Routing.Me.ID)
    }()
    go func() {
        cc[1] <- k[len(k)-1].LookupContact(k[0].Net.Routing.Me.ID)
    }()

	// Check that k elements were returned, if k elements exist
	//TODO: one more contact is returned than expected
    con1 := <-cc[0]
	expected := imin(nodes, K-1)
	log.Println("k",K," nodes",nodes," expected",expected)
    if len(con1) != expected {
        log.Println("incorrect amount of contacts returned, got",len(con1),"expected",expected)
        t.Fail()
    }
    con2 := <-cc[1]
    if len(con2) != expected {
        log.Println("incorrect amount of contacts returned, got",len(con2),"expected",expected)
        t.Fail()
    }
    log.Println("\n\nRESULTS\nOUTGOING\n",con1)
    log.Println("\n\nRESULTS\nINCOMING\n",con2)

    // Test pinging various, presumably connected nodes
    log.Println("\n\nSending messages")
    for i := 0; i<len(k); i++ {
        // If the node is the same, skip
        if i == len(k)-1-i {
            continue
        } else {
            msg := &NetworkMessage{MsgType: rpc.PING_MSG, Origin: k[i].Net.Routing.Me, RpcID: *NewKademliaIDRandom()}
            response := k[i].Net.SendReceiveMessage(UDP, msg, &k[len(k)-1-i].Net.Routing.Me) //&center.Net.Routing.Me

            // Check the response
            if response.MsgType != rpc.PONG_MSG ||
                !response.RpcID.Equals(&msg.RpcID) ||
                !response.Origin.ID.Equals(k[len(k)-1-i].Net.Routing.Me.ID) {
                t.Fail()
            }
        }
    }

}
