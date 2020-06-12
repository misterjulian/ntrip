NTRIP Caster / Client / Server implementation in Go

[![Go Report Card](https://goreportcard.com/badge/github.com/go-gnss/ntrip)](https://goreportcard.com/report/github.com/go-gnss/ntrip)

### Run a Caster 
Application in `cmd/ntripcaster/` configurable with `caster.yml`.

```
# Building
go build -o ntripcaster ./cmd/ntripcaster/ntripcaster.go

# Running 
./ntripcaster --config=caster.yml
```
