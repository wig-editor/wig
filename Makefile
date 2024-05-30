
run:
	-pkill -f go-build
	go run cmd/main.go 2> /tmp/mcwig.panic.txt

test:
	go test -v ./... -count=1
