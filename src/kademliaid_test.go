package d7024e

import(
	//"fmt"
	"testing"
)

// Functions to test:
// func NewKademliaID(data string) *KademliaID
// func NewRandomKademliaID() *KademliaID
// func (kademliaID KademliaID) Less(otherKademliaID *KademliaID) bool
// func (kademliaID KademliaID) Equals(otherKademliaID *KademliaID) bool
// func (kademliaID KademliaID) CalcDistance(target *KademliaID) *KademliaID
// func (kademliaID *KademliaID) String()

func TestNewKademliaID(t *testing.T) {	
	id := NewKademliaID("FFFFFFFF00000000000000000000000000000000")
	if id == nil {
		t.Fail()
	}	
}

func TestNewRandomKademliaID(t *testing.T) {	id := NewRandomKademliaID()
	if id == nil {
		t.Fail()
	}	
}

func TestLess(t *testing.T) {
	id1 := NewKademliaID("FFFFFFFF00000000000000000000000000000000")
	id2 := NewKademliaID("FFFFFFFF00000000000000000000000000000001")
	
	b1 := id1.Less(id2) // should return true
	b2 := id2.Less(id1) // should return false
	
	if b1 != true {
		t.Fail()
	}
	
	if b2 != false {
		t.Fail()
	}
	
}

func TestEquals(t *testing.T) {

	id1 := NewKademliaID("FFFFFFFF00000000000000000000000000000000")
	id2 := id1
	id3 := NewKademliaID("FFFFFFFF00000000000000000000000000000001")
	
	b1 := id1.Equals(id2) // should return true
	b2 := id1.Equals(id3) // should return false
	
	if b1 != true {
		t.Fail()
	}
	
	if b2 != false {
		t.Fail()
	}
}

func TestCalcDistance(t *testing.T) {
	id1 := NewKademliaID("FFFFFFFF00000000000000000000000000000000")
	id2 := NewKademliaID("FFFFFFFF00000000000000000000000000000001")
	
	id3 := NewRandomKademliaID()
	id4 := NewRandomKademliaID()
	
	d1 := id1.CalcDistance(id2)
	d2 := id2.CalcDistance(id1)
	
	d3 := id3.CalcDistance(id4)
	d4 := id4.CalcDistance(id3)
	
	if *d1 != *d2 {
		t.Fail()
	}
	
	if *d3 != *d4 {
		t.Fail()
	}	
}
