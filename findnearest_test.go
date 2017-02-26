package main

import (
	"testing"
)

func TestNewPointFromLatLngStrings1(t *testing.T) {
	p, _ := newPointFromLatLngStrings("51.92125", "6.57755")
	if p[1] != 51.92125 || p[0] != 6.57755 {
		t.Fail()
	}
}

func TestNewPointFromLatLngStrings2(t *testing.T) {
	_, err := newPointFromLatLngStrings("", "")
	// attempting to create a point from empty strings should give an error
	if err == nil {
		t.Fail()
	}
}

func TestNewPointFromLatLngStrings3(t *testing.T) {
	_, err := newPointFromLatLngStrings("NULL", "NULL") //NaN, NaN will succeed!
	// attempting to create a point from strings, apart from NaN, should give an error
	if err == nil {
		t.Fail()
	}
}

func TestMax1(t *testing.T) {
	m := max(1, 2)
	if m != 2 {
		t.Fail()
	}
}

func TestMax2(t *testing.T) {
	m := max(2, 1)
	if m != 2 {
		t.Fail()
	}
}
