
run:
	-pkill -f go-build
	go run cmd/main.go

test:
	go test -v ./... -count=1
