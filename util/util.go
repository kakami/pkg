package util

import (
    "fmt"
    "math"
    "math/rand"
    "os"
    "os/user"
    "path/filepath"
    "runtime"
)

var chars = []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789_")

func RandomString(length int) string {
    var out []byte
    // rand.Seed(time.Now().UnixNano())
    for i := 0; i < length; i++ {
        out = append(out, chars[rand.Int()%len(chars)])
    }
    return string(out)
}

func Prime(d int64) int64 {
    for i := d; ; i++ {
        if isPrime(i) {
            return i
        }
    }
}

func isPrime(d int64) bool {
    if d == 2 || d == 3 {
        return true
    }
    if d < 2 || d%2 == 0 || d%3 == 0 {
        return false
    }
    sd := int64(math.Sqrt(float64(d)))
    for i := int64(5); i <= sd; i += 2 {
        if d%i == 0 {
            return false
        }
    }
    return true
}

func ConfigDir(dirName string) (string, error) {
    if userConfigDir, err := os.UserConfigDir(); err == nil {
        return filepath.Join(userConfigDir, dirName), nil
    }

    if runtime.GOOS == "windows" {
        return filepath.Join(os.Getenv("APPDATA"), dirName), nil
    }

    if xdgConfigDir := os.Getenv("XDG_CONFIG_HOME"); xdgConfigDir != "" {
        return filepath.Join(xdgConfigDir, dirName), nil
    }

    homeDir := guessUnixHomeDir()
    if homeDir == "" {
        return "", fmt.Errorf("unable to get current user home directory: os/user lookup failed; $HOME is empty")
    }
    return filepath.Join(homeDir, ".config", dirName), nil
}

func guessUnixHomeDir() string {
    usr, err := user.Current()
    if err == nil {
        return usr.HomeDir
    }
    return os.Getenv("HOME")
}
