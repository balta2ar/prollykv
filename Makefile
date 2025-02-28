.PHONY: build test run

build:
	go build -o prollykv .

vet:
	go vet ./...

lint:
	golint ./...

test:
	go test ./... -v -p 1 -count=1 -timeout 5s

run:
	go run cmd/prollykv/main.go

dot: $(patsubst %.dot,%.dot.png,$(wildcard *.dot))

%.dot.png: %.dot
	dot -Kneato -Tpng $< -o $@