NTRIP Caster / Client / Server implementation in Go

### Run a Caster 
Application in `cmd/ntripcaster/` configurable with `caster.yml`.

```
# Building
go build -o ntripcaster ./cmd/ntripcaster/ntripcaster.go

# Running 
./ntripcaster --config=caster.yml
```
