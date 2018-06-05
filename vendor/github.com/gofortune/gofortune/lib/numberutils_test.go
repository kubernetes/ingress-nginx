package lib

import "testing"

func TestMin(t *testing.T) {
	if (Min(10, 20) != 10) || (Min(20, 10) != 10) {
		t.Error("Expected 10")
	}
}

func TestMax(t *testing.T) {
	if (Max(10, 20) != 20) || (Max(20, 10) != 20){
		t.Error("Expected 20")
	}
}
