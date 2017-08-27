/*
* Kadmelia node
* - at least contains ID 
*/

package kademlia

type node struct {
    Id Id_t
}

func NewNode(random bool) (node_out node) {
    node_out.Id = NewId(true)
    return
}
