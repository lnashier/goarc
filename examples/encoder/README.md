# Example CLI App

# Build & Run

```shell
go build -o bin/ .
./bin/encoder base64encode str1 | xargs -L1 ./bin/encoder base64decode
```

# Run

```shell
go run . base64encode str1 
```
