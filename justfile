set default-list := true

coverage-report := justfile_directory() / "coverage.out"
compiled-binary := justfile_directory() / "update-link"
development-config := justfile_directory() / "config/dev.toml"

[group('compiling')]
clean:
    rm -f {{compiled-binary}} {{coverage-report}}

[group('compiling')]
build: clean
    go build -o {{compiled-binary}} .

[group('compiling')]
run: build
    {{compiled-binary}} -config {{development-config}}

[group('test')]
test:
    go test -cover ./...

[group('test')]
cover: clean
    go test -coverprofile={{coverage-report}} ./...
    go tool cover -html={{coverage-report}}

[group('housekeeping')]
format:
    go fmt ./...

[group('housekeeping')]
fix-deps:
    go mod tidy
