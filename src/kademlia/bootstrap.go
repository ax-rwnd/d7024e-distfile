package kademlia

import(
	"fmt"
	"errors"
	"math/rand"
)

func Bootstrap(bootPort int, bootAddr string, tcpPort int, udpPort int, netw *Network) {

	tmpID := NewKademliaID("0000000000000000000000000000000000000000") // dummy ID
	boot := NewContact(tmpID, bootAddr, tcpPort, udpPort)

	// k should be a list of contacts returning, targetID to boot
	k, bootID := netw.FindContactAndID(netw.Routing.Me.ID, &boot)
	if newt.Routing.Me.ID == bootID {
		return
	}

	for _, contact := range k {
		netw.Routing.AddContact(contact, nil)		
	}

	boot = NewContact(&bootID, bootAddr, tcpPort, udpPort)
	
	b, _ := netw.Routing.AddContact(boot, netw.SendPingMessage)

	if !b {
		errors.New("contact couldn't be added")
	}

	defer func() {
		if r := recover(); r!=nil {
			fmt.Println("Recovering...")
		}
	}()

	// all index from 0 to the bootIndex is further away from the nóde than the boot node
	bootIndex := netw.Routing.getBucketIndex(boot.ID)
	// pick a random node in each bucket to send node lookup on
	for i := 0; i < bootIndex; i++ {
        // get the bucket we're going to pick a random index from
		if bucket := netw.Routing.GetBucket(i); bucket.Len() > 0 {
            j := rand.Intn(bucket.Len())
            n := 0
            for e := bucket.list.Front(); e != nil; e = e.Next() {
                if j==n {
                    contact := e.Value.(Contact) // this might not work
                    nodeID := e.Value.(Contact).ID
                    netw.SendFindContactMessage(nodeID, &contact)
                }
                n++
            }
        } else {
            fmt.Println("No buckets to ping.")
        }
    }
	
}
