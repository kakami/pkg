package net

import (
	"net"
)

func CurrentIP() ([]string, error) {
	var out []string
	var err error

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return out, err
	}
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				out = append(out, ipnet.IP.String())
			}
		}
	}
	return out, err
}

func DefaultIP() (string, error) {
	return getDefaultIP()
}

func SetProxy(addr, port string, enable bool) {
	setSystemProxy(addr, port, enable)
}
