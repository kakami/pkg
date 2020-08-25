package net_test

import (
	"fmt"
	"testing"

	"pkg/net"
)

func Test_CurrentIP(t *testing.T) {
	ips, err := net.CurrentIP()
	fmt.Println(ips)
	fmt.Println(err)
}

func Test_Region(t *testing.T) {
	t.Log(net.RegionCode("45.32.65.1"))
	t.Log(net.RegionCode("8.8.8.8"))
}
