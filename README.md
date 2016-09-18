# Lightweight systemd journal package for Go

This package provides a lightweight logging object that writes to systemd's
[journal](https://www.freedesktop.org/software/systemd/man/systemd-journald.service.html).

It implements the `lwlog.Logger` interface from the
[lwlog](https://github.com/lwithers/lwlog) package.

A trivial example:

```go
lg := lwjournal.New("my-app-name")
lg.AddVariable("foo", "bar")
lg.Infof("an integer: %d", 123")
```

Which would show up in the journal as:

TODO
