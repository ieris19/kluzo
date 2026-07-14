_default:
    just --list

clean:
    rm -f update-link coverage.out

build: clean
    go -C ./src build -o ../update-link .

test:
    go -C ./src test -cover ./...

cover: clean
    go -C ./src test -coverprofile=../coverage.out ./...
    go -C ./src tool cover -html=../coverage.out

run: build
    ./update-link -config dev.toml

format:
    go -C ./src fmt ./...

fix-deps:
    go -C ./src mod tidy
