PKG = github.com/k1LoW/runn
COMMIT = $$(git describe --tags --always)
OSNAME=${shell uname -s}
DATE = $$(date '+%Y-%m-%d_%H:%M:%S%z')

export GO111MODULE=on

BUILD_LDFLAGS = -X $(PKG).commit=$(COMMIT) -X $(PKG).date=$(DATE)

default: test

ci: depsdev test-all

test: cert
	go test ./... -coverprofile=coverage.out -covermode=count

race: cert
	go test ./... -race

test-integration: cert
	chmod 600 testdata/sshd/id_rsa
	go test ./... -tags=integration -count=1

test-all: cert
	chmod 600 testdata/sshd/id_rsa
	go test ./... -tags=integration -coverprofile=coverage.out -covermode=count

benchmark: cert
	go test -bench . -benchmem -run Benchmark | octocov-go-test-bench --tee > custom_metrics_benchmark.json

lint:
	golangci-lint run ./...
	govulncheck ./...
	go vet -vettool=`which gostyle` -gostyle.config=$(PWD)/.gostyle.yml ./...

doc:
	go run ./scripts/fndoc.go

build:
	go build -ldflags="$(BUILD_LDFLAGS)" -o runn cmd/runn/main.go

depsdev:
	go install github.com/Songmu/ghch/cmd/ghch@latest
	go install github.com/Songmu/gocredits/cmd/gocredits@latest
	go install golang.org/x/vuln/cmd/govulncheck@latest
	go install github.com/k1LoW/octocov-go-test-bench/cmd/octocov-go-test-bench@latest
	go install github.com/k1LoW/gostyle@latest

cert:
	rm -f testdata/*.pem testdata/*.srl
	openssl req -x509 -newkey rsa:4096 -days 365 -nodes -sha256 -keyout testdata/cakey.pem -out testdata/cacert.pem -subj "/C=UK/ST=Test State/L=Test Location/O=Test Org/OU=Test Unit/CN=*.example.com/emailAddress=k1lowxb@gmail.com"
	openssl req -newkey rsa:4096 -nodes -keyout testdata/key.pem -out testdata/csr.pem -subj "/C=JP/ST=Test State/L=Test Location/O=Test Org/OU=Test Unit/CN=*.example.com/emailAddress=k1lowxb@gmail.com"
	openssl x509 -req -sha256 -in testdata/csr.pem -days 60 -CA testdata/cacert.pem -CAkey testdata/cakey.pem -CAcreateserial -out testdata/cert.pem -extfile testdata/openssl.cnf
	openssl verify -CAfile testdata/cacert.pem testdata/cert.pem

prerelease:
	git pull origin main --tag
	go mod tidy
	ghch -w -N ${VER}
	gocredits -skip-missing -w .
	cat _EXTRA_CREDITS >> CREDITS
	git add CHANGELOG.md CREDITS go.mod go.sum
	git commit -m'Bump up version number'
	git tag ${VER}

prerelease_for_tagpr:
	gocredits -skip-missing -w .
	cat _EXTRA_CREDITS >> CREDITS
	git add CHANGELOG.md CREDITS go.mod go.sum

release:
	git push origin main --tag
	goreleaser --clean

.PHONY: default test
