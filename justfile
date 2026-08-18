set default-list := true

repo-root := justfile_directory()
go-cmd := repo-root / "cmd/update-link"
coverage-report := repo-root / "coverage.out"
compiled-binary := repo-root / "update-link"
development-config := repo-root / "config/dev.toml"

[group('compiling')]
clean:
    rm -f {{compiled-binary}} {{coverage-report}}

[group('compiling')]
build: clean
    go build -o {{compiled-binary}} {{go-cmd}}

[group('compiling')]
run: build
    {{compiled-binary}} -config {{development-config}}

[group('test')]
test:
    go test -cover {{repo-root}}/...

[group('test')]
cover: clean
    go test -coverprofile={{coverage-report}} {{repo-root}}/...
    go tool cover -html={{coverage-report}}

[group('housekeeping')]
format:
    go fmt {{repo-root}}/...

[group('housekeeping')]
fix-deps:
    go mod tidy
