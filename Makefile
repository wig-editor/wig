
run:
	-pkill -f go-build
	go run cmd/main.go 2> /tmp/mcwig.panic.txt

test:
	go test -v ./... -count=1

build:
	go build cmd/main.go
	mv ./main ~/go/bin/mcwig
