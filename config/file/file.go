package file

import (
    "fmt"
    "io"
    "os"
    "path/filepath"
    "strings"

    "github.com/kakami/pkg/config"
)

var _ config.Source = (*file)(nil)

type file struct {
    opts options
    path string
}

// NewSource new a file source.
func NewSource(path string, opts ...Option) config.Source {
    o := options{
        suffixes: []string{".yaml", ".yml", ".json", ".xml"},
    }
    for _, opt := range opts {
        opt(&o)
    }
    return &file{
        opts: o,
        path: path,
    }
}

func (f *file) loadFile(path string) (*config.KeyValue, error) {
    valid := false
    for idx := range f.opts.suffixes {
        if strings.HasSuffix(path, f.opts.suffixes[idx]) {
            valid = true
            break
        }
    }
    if !valid {
        return nil, fmt.Errorf("invalid suffix: %s", path)
    }
    file, err := os.Open(path)
    if err != nil {
        return nil, err
    }
    defer file.Close()
    data, err := io.ReadAll(file)
    if err != nil {
        return nil, err
    }
    info, err := file.Stat()
    if err != nil {
        return nil, err
    }
    return &config.KeyValue{
        Key:    info.Name(),
        Format: format(info.Name()),
        Value:  data,
    }, nil
}

func (f *file) loadDir(path string) (kvs []*config.KeyValue, err error) {
    files, err := os.ReadDir(path)
    if err != nil {
        return nil, err
    }
    for _, file := range files {
        // ignore hidden files
        if file.IsDir() || strings.HasPrefix(file.Name(), ".") {
            continue
        }
        // ignore suffixes
        if len(f.opts.suffixes) > 0 {
            var goon bool
            for i := range f.opts.suffixes {
                if strings.HasSuffix(file.Name(), f.opts.suffixes[i]) {
                    goon = true
                    break
                }
            }
            if !goon {
                continue
            }
        }
        kv, err := f.loadFile(filepath.Join(path, file.Name()))
        if err != nil {
            return nil, err
        }
        kvs = append(kvs, kv)
    }
    return
}

func (f *file) Load() (kvs []*config.KeyValue, err error) {
    fi, err := os.Stat(f.path)
    if err != nil {
        return nil, err
    }
    if fi.IsDir() {
        return f.loadDir(f.path)
    }
    kv, err := f.loadFile(f.path)
    if err != nil {
        return nil, err
    }
    return []*config.KeyValue{kv}, nil
}

func (f *file) Watch() (config.Watcher, error) {
    return newWatcher(f)
}
