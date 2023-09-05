package file

type Option func(*options)

type options struct {
    suffixes []string
}

func WithSuffix(s ...string) Option {
    return func(o *options) {
        o.suffixes = s
    }
}
