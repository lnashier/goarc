# Example CLI App

A minimal `cli.Service`. Registers one command, `echo`, which sleeps for 10 seconds before
printing its argument — sleeping via `xtime.SleepWithContext`, so `Ctrl-C` interrupts it early
and cleanly instead of waiting out the full 10 seconds. Try running it and pressing `Ctrl-C`
partway through to see the cancellation path.

# Build & Run

```shell
go build -o bin/mockcli .
./bin/mockcli echo hello
```

# Run

```shell
go run . echo Hello 
```
