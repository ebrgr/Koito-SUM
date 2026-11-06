package main

import "testing"

func TestLoopbackOnly(t *testing.T) {
	for _, addr := range []string{"127.0.0.1:14111", "[::1]:14111"} {
		if err := loopback(addr); err != nil {
			t.Fatal(err)
		}
	}
	for _, addr := range []string{":14111", "0.0.0.0:14111", "192.168.1.1:14111", "localhost:14111", "bad"} {
		if loopback(addr) == nil {
			t.Fatalf("must reject %s", addr)
		}
	}
}
