# Lightweight systemd journal package for Go

This package provides a lightweight logging object that writes to systemd's
[journal](https://www.freedesktop.org/software/systemd/man/systemd-journald.service.html).

It implements the `lwlog.Logger` interface from the
[lwlog](https://github.com/lwithers/lwlog) package.

A trivial example:

```go
func BuildLogger() lwlog.Logger {
	lg, err := lwjournal.New()
	if err != nil {
		return log.NewStd() // fallback
	}
	lg.AddVariable("FOO", "bar")
}

func main() {
	lg := BuildLogger()
	lg.Infof("an integer: %d", 123)
}
```

Which would show up in the journal as:

```
lwithers@ruby demo $ journalctl -n 1 -o verbose _CMDLINE=./demo
-- Logs begin at Tue 2016-02-16 22:29:42 GMT, end at Sun 2016-09-18 19:48:49 BST. --
Sun 2016-09-18 19:47:55.345984 BST [s=…]
    _UID=1000
    _GID=1000
    _AUDIT_SESSION=1
    _AUDIT_LOGINUID=1000
    _SYSTEMD_CGROUP=/user.slice/user-1000.slice/session-1.scope
    _SYSTEMD_SESSION=1
    _SYSTEMD_OWNER_UID=1000
    _SYSTEMD_UNIT=session-1.scope
    _SYSTEMD_SLICE=user-1000.slice
    _BOOT_ID=…
    _MACHINE_ID=…
    _HOSTNAME=ruby
    PRIORITY=6
    _TRANSPORT=journal
    _CAP_EFFECTIVE=0
    CODE_FILE=/home/lwithers/go/src/github.com/lwithers/lwjournal/cmd/demo/journal.go
    FOO=bar
    _COMM=demo
    _EXE=/home/lwithers/go/src/github.com/lwithers/lwjournal/cmd/demo/demo
    _CMDLINE=./demo
    CODE_FUNC=main.runTest
    MESSAGE=an integer: 123
    CODE_LINE=31
    _PID=6287
    _SOURCE_REALTIME_TIMESTAMP=1474224475345984
```
