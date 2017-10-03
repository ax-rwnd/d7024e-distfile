package main

import(
	"errors"
	"math/rand"
	"kademlia"
)

func Bootstrap(port int, addr string, netw *kademlia.Network) {

	tmpID := kademlia.NewKademliaID("0000000000000000000000000000000000000000") // dummy ID
	boot := kademlia.NewContact(tmpID, fmt.Sprintf("%s:%d", addr, port))
	
	k, bootID := netw.FindContactAndID(self, boot) // k should be a list of contacts returning, targetID to boot
	
	for _, contact := range k {
		netw.Routing.AddContact(contact, nil)	
	}	
	
	boot = kademlia.NewContact(bootID, fmt.Sprintf("%s:%d", addr, port)
	b, _ := netw.Routing.AddContact(boot, netw.SendPingMessage)
	
	if !b {
		errors.new("contact couldn't be added")
	}

	bootIndex := netw.Routing.getBucketIndex(boot) // all index from 0 to the bootIndex is further away from the nóde than the boot node
	for i := 0; i < bootIndex; i++ { // pick a random node in each bucket to send node lookup on
		bucket := netw.Routing.GetBucket(i) // get the bucket we're going to pick a random index from
        j := rand.Intn(bucket.Len())
		n := 0
		for e := bucket.list.Front(); e != nil; e = e.Next() {
			if j==n {
				contact := e.Value.(Contact) // this might not work
				nodeID := e.Value.(Contact).ID
				netw.SendFindContactMessage(nodeID, contact)
			}
			n++
		}	
    }

}