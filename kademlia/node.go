/*
* Kadmelia node
* - at least contains ID 
*/

package kademlia

type Node struct {
    Id Id_t
    KVStore map[Id_t]string
}

func NewNode(random bool) (node_out Node) {
        node_out.Id = NewId(random)
        node_out.KVStore = make(map[Id_t]string)
    return
}

func Distance (a, b Node) (c Id_t) {
        for i := 0; i<20; i++ {
            c[i] = a.Id[i]^b.Id[i]
        }
    return
}
