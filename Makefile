.PHONY: all

run:
	go run cmd/main.go > /tmp/wig.panic.txt 2>&1

test:
	go test -v ./... -count=1

build:
	go build cmd/main.go
	mv ./main ~/go/bin/wig
	
build-run:
	go build cmd/main.go
	mv ./main ~/go/bin/wig
	wig > /tmp/wig.panic.txt 2>&1

setup-runtime:
	mkdir -p ~/.config/wig
	cp -r ./runtime/* ~/.config/wig/
