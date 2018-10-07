GO_VERSION:=$(shell go version)

.PHONY: bench profile clean test

all: install

bench:
	go test -count=5 -run=NONE -bench . -benchmem

profile: clean
	mkdir pprof
	mkdir bench
	go test -count=10 -run=NONE -bench . -benchmem -o pprof/test.bin -cpuprofile pprof/cpu.out -memprofile pprof/mem.out
	go tool pprof --svg pprof/test.bin pprof/mem.out > bench/mem.svg
	go tool pprof --svg pprof/test.bin pprof/cpu.out > bench/cpu.svg

clean:
	rm -rf bench
	rm -rf pprof
	rm -rf ./*.svg
	rm -rf ./*.log

test:
	GOCACHE=off go test --race -coverprofile=cover.out ./...
	go tool cover -html=cover.out -o cover.html
	rm -rf cover.out

dbuild:
	sudo docker build --pull=true --file=Dockerfile -t kpango/golang-server-template:latest .

dpush: dbuild
	sudo docker push kpango/golang-server-template:latest
