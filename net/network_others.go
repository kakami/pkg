// +build !darwin

package net

import "fmt"

func getDefaultIP() (string, error) {
	return "", fmt.Errorf("not supported")
}

func setSystemProxy(addr, port string, enable bool) {
}