package net

import (
    "fmt"
    "net"

    "github.com/kakami/pkg/types"
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

func LocalInterfaces() ([]string, error) {
    iis, err := net.Interfaces()
    if err != nil {
        return nil, err
    }
    ips := types.NewUnsafeSet()
    for idx := range iis {
        if iis[idx].Flags&net.FlagUp == 0 {
            continue
        }
        if iis[idx].Flags&net.FlagLoopback != 0 {
            continue
        }
        addrs, err := iis[idx].Addrs()
        if err != nil {
            return nil, err
        }

        for i := range addrs {
            ip := getIpFromAddr(addrs[i])
            if ip == nil {
                continue
            }
            _, ipn, err := net.ParseCIDR(addrs[i].String())
            if err != nil {
                return nil, err
            }
            m, _ := ipn.Mask.Size()
            if m == 32 {
                continue
            }
            ips.Add(ip.String())
        }
    }

    if ips.Length() < 1 {
        return nil, fmt.Errorf("no valid interface")
    }

    return ips.Values(), nil
}

func getIpFromAddr(addr net.Addr) net.IP {
    var ip net.IP
    switch v := addr.(type) {
    case *net.IPNet:
        ip = v.IP
    case *net.IPAddr:
        ip = v.IP
    }
    if ip == nil || ip.IsLoopback() {
        return nil
    }
    ip = ip.To4()
    if ip == nil {
        return nil // not an ipv4 address
    }
    return ip
}
