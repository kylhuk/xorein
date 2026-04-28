package main

import (
	"testing"
)

func TestTcpAddrToMultiaddr(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"127.0.0.1:9000", "/ip4/127.0.0.1/tcp/9000"},
		{"0.0.0.0:1337", "/ip4/0.0.0.0/tcp/1337"},
		{":9001", "/ip4/0.0.0.0/tcp/9001"},
		{"/ip4/127.0.0.1/tcp/0", "/ip4/127.0.0.1/tcp/0"}, // already a multiaddr
		{"192.168.1.5:4242", "/ip4/192.168.1.5/tcp/4242"},
	}
	for _, c := range cases {
		got := tcpAddrToMultiaddr(c.input)
		if got != c.want {
			t.Errorf("tcpAddrToMultiaddr(%q) = %q, want %q", c.input, got, c.want)
		}
	}
}

func TestTcpAddrToMultiaddrNoPort(t *testing.T) {
	got := tcpAddrToMultiaddr("127.0.0.1")
	want := "/ip4/127.0.0.1/tcp/0"
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
}
