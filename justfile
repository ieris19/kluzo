set default-list := true

[group('compiling')]
clean:
    rm -f update-link coverage.out

[group('compiling')]
build: clean
    go build -o update-link .

[group('test')]
test:
    go test -cover ./...

[group('test')]
cover: clean
    go test -coverprofile=coverage.out ./...
    go tool cover -html=coverage.out

[group('compiling')]
run: build
    ./update-link -config config/dev.toml

[group('housekeeping')]
format:
    go fmt ./...

[group('housekeeping')]
fix-deps:
    go mod tidy
