/*
* K-Bucket
* - One list for each bit of the ID
* - Each list entry contains the data to find one node (ip, port, ID)
* - Nodes in the nth list have a differing nth bit from the node's ID, bits
*       before that must match that of the node
*/

package kademlia
