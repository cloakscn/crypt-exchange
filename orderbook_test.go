package main

import "testing"

func TestOrderbook(t *testing.T) {

}

func TestLimit(t *testing.T) {
	l := NewLimit(10_000)
	buyOrder := NewOrder(true, 5)
}