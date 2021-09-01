// +build darwin

package net

import (
	"fmt"
	"net"
	"os/exec"
	"strings"
)

func getDefaultIP() (string, error) {
	cmd := exec.Command("/sbin/route", "-n", "get", "default")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}

	var interfaceName string

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[0] == "interface:" {
			interfaceName = fields[1]
			break
		}
	}

	if interfaceName == "" {
		return "", fmt.Errorf("no valid interface")
	}

	device, err := net.InterfaceByName(interfaceName)
	if err != nil {
		return "", err
	}

	addrs, err := device.Addrs()
	if err != nil {
		return "", err
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), err
			}
		}
	}

	return "", fmt.Errorf("no valid addr")
}

func setSystemProxy(addr, port string, enable bool) {
		setHttpProxy(addr, port, enable)
		setHttpsProxy(addr, port, enable)
		setSocksProxy(addr, port, enable)
}

func setHttpProxy(addr, port string, enable bool) {
	if enable {
		cmd := exec.Command("networksetup", "-setwebproxy", "Wi-Fi", addr, port)
		cmd.Start()
		cmd = exec.Command("networksetup", "-setwebproxystate", "Wi-Fi", "on")
		cmd.Start()
	} else {
		cmd := exec.Command("networksetup", "-setwebproxystate", "Wi-Fi", "off")
		cmd.Start()
	}
}

func setHttpsProxy(addr, port string, enable bool) {
	if enable {
		cmd := exec.Command("networksetup", "-setsecurewebproxy", "Wi-Fi", addr, port)
		cmd.Start()
		cmd = exec.Command("networksetup", "-setsecurewebproxystate", "Wi-Fi", "on")
		cmd.Start()
	} else {
		cmd := exec.Command("networksetup", "-setsecurewebproxystate", "Wi-Fi", "off")
		cmd.Start()
	}
}

func setSocksProxy(addr, port string, enable bool) {
	if enable {
		cmd := exec.Command("networksetup", "-setsocksfirewallproxy", "Wi-Fi", addr, port)
		cmd.Start()
		cmd = exec.Command("networksetup", "-setsocksfirewallproxystate", "Wi-Fi", "on")
		cmd.Start()
	} else {
		cmd := exec.Command("networksetup", "-setsocksfirewallproxystate", "Wi-Fi", "off")
		cmd.Start()
	}
}