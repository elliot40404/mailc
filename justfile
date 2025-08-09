set windows-shell := ["pwsh.exe", "-NoLogo", "-Command"]

default: build_no_lint

build_cmd := if os() == "windows" { "go build -o ./bin/mailc.exe ./cmd/mailc/"} else { "go build -o ./bin/mailc ./cmd/mailc/" }
rm_cmd := if os() == "windows" { "mkdir ./bin -Force; Remove-Item -Recurse -Force ./bin" } else { "rm -rf ./bin" }
exec_cmd := if os() == "windows" { "./bin/mailc.exe" } else { "./bin/mailc" }

build: clean lint test
    {{build_cmd}}

build_no_lint: clean test
    {{build_cmd}}

install:
    go install ./cmd/mailc

exec:
    {{exec_cmd}}

build_run: build exec

build_run_nl: build_no_lint exec

clean:
    {{rm_cmd}}

lint:
    golangci-lint run

lint-fix:
    golangci-lint run --fix

test:
    go test ./...

testv:
    go test -v ./...

vendor:
    go mod tidy
    go mod vendor
    go mod tidy

release:
    goreleaser release --snapshot --clean

gen-examples: build
    ./bin/mailc generate -input ./examples/templates -output ./examples/generated
