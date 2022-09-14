//go:build linux
// +build linux

package elog

func RedirectStderr(path string) error {
    ff, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_SYNC|os.O_APPEND, 0644)
    if err != nil {
        return err
    }
    if err = syscall.Dup3(int(ff.Fd()), int(os.Stderr.Fd()), 0); err != nil {
        return err
    }
    return err
}
