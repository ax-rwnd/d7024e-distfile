/*
* K-Bucket
* - One list for each bit of the ID
* - The bucket length is determined by the distance from itself, with 0<i<160
*       len = 2^i
* - Each list entry contains the data to find one node (ip, port, ID)
* - Nodes in the nth list have a differing nth bit from the node's ID, bits
*       before that must match that of the node
*/

package kademlia

//Global k-factor that determines the replicativity of the system
const k_repfactor = 20

//A list of the nodes that have a specific distance from some other node
type KBucket struct {
    hosts []Node
    elements int
    capacity int
}

//Create a KBucket with k capacity and a limited slice
func CreateKBucket(capacity int) (output KBucket) {
    output.hosts = make([]Node, k_repfactor)
    output.elements = 0
    output.capacity = capacity
    return
}

//Insert or update a node in the bucket by a O(k) search
func (t *KBucket) Insert(target *Node) error {
    for i, e := range t.hosts {
        if e.Id == target.Id {
            tmp := e
            j := i
            for ; j<t.elements-1; j++ {
                t.hosts[j] = t.hosts[j+1]
            }
            t.hosts[j] = tmp
            return nil
        }
    }

    //If the list is full already, try tp evict oldest
    if t.elements >= t.capacity || t.elements >= 20 {
        if err := t.Evict(); err != nil {
            return err
        }
    }

    //The element is not in the list, add it
    t.hosts[t.elements] = *target
    t.elements++
    return nil
}

//Try to evict the head of the list
func (t *KBucket) Evict() error {
    //TODO: not implemented
    return nil
}
