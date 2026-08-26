set default-list := true

repo-root := justfile_directory()
project-name := "kluzo"
go-cmd := repo-root / "cmd" / project-name
coverage-report := repo-root / "coverage.out"
compiled-binary := repo-root / project-name
development-config := repo-root / "config/dev.toml"

# Delete build artifacts from the repository
[group: 'housekeeping']
clean:
    rm -f {{compiled-binary}} {{coverage-report}}

# Reconcile go.mod and go.sum with the code's actual imports
[group: 'housekeeping']
fix-deps:
    go mod tidy

# Build the app into an executable
[group: 'compiling']
build: clean
    go build -o {{compiled-binary}} {{go-cmd}}

# Execute the app using the development configuration
[group: 'compiling']
run: build
    {{compiled-binary}} -config {{development-config}}

[group: 'compiling']
install: build
    mkdir -p $HOME/.local/bin
    mkdir -p {{"${XDG_CONFIG_HOME:-$HOME/.config}" / project-name}}
    install --mode 0750 {{compiled-binary}} $HOME/.local/bin
    install --mode 0640 {{development-config}} {{"${XDG_CONFIG_HOME:-$HOME/.config}" / project-name / "config.toml"}}

# Run the test suite with coverage
[group: 'test']
test:
    go test -cover {{repo-root}}/...

# Run the tests and open an HTML coverage report
[group: 'test']
cover: clean
    go test -coverprofile={{coverage-report}} {{repo-root}}/...
    go tool cover -html={{coverage-report}}

# Apply standard Go formatting to all files in the repo
[group: 'quality']
format:
    go fmt {{repo-root}}/...

# Check whether all files in the repo are correctly formatted
[group: 'quality']
format-check:
    #!/usr/bin/env bash
    set -euxo pipefail
    # gofmt does not fail, so we read the output and define our own fail-state
    unformatted="$(gofmt -l {{repo-root}})"
    if [ -n "$unformatted" ]; then
      echo "The following files are not gofmt'd:"
      echo "$unformatted"
      exit 1
    fi

# Run static analysis across the whole repo
[group: 'quality']
static-check:
    go vet {{repo-root}}/...

# Run the continuous integration pipeline
[group: 'ci']
ci: format-check static-check test

# If a forgejo-runner is locally available, check that the pipeline is correctly configured
[group: 'ci']
ci-check:
    #!/usr/bin/env bash
    set -euxo pipefail
    if command -v forgejo-runner > /dev/null 2>&1; then
      forgejo-runner exec -i docker.io/library/golang:1.26-trixie --detect-event --dryrun
    else
      echo "No local CI runner"
      exit 1
    fi
