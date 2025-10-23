# golang-microservices-sandbox
some microservices


## Pre-Commit
### golangci-lint
Install:
```
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b $(go env GOPATH)/bin v2.5.0
```

Run:
```
make lint

make lint-fix
```