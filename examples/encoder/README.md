# Example CLI App

# Build & Run

```
go build -o bin/ .
./bin/encoder base64encode str1 | xargs -L1 ./bin/encoder base64decode
```

# Run

```
go run . base64encode str1 
```
