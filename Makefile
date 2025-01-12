
run:
	go run cmd/main.go > /tmp/mcwig.panic.txt 2>&1

test:
	go test -v ./... -count=1

build:
	go build cmd/main.go
	mv ./main ~/go/bin/mcwig
