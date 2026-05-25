# Makefile for bootstrap_sim
# Requires Go 1.22+

BINARY   := bootstrap
CMD      := ./cmd/bootstrap
GOFLAGS  := -mod=vendor

.PHONY: all build run test bench profile-cpu profile-mem clean

all: build

## build: compile the simulation binary
build:
	go build $(GOFLAGS) -o $(BINARY) $(CMD)

## run: run the simulation with default parameters
run: build
	./$(BINARY)

## run-debug: run with debug logging enabled
run-debug: build
	./$(BINARY) -debug

## run-json: run with JSON-formatted log output
run-json: build
	./$(BINARY) -json 2>&1 | head -40

## test: run all unit tests
test:
	go test $(GOFLAGS) -v ./...

## bench: run all benchmarks (1 second each)
bench:
	go test $(GOFLAGS) -bench=. -benchmem -benchtime=1s ./...

## bench-bootstrap: benchmark only the bootstrap package
bench-bootstrap:
	go test $(GOFLAGS) -bench=. -benchmem ./internal/bootstrap/...

## profile-cpu: run with CPU profiling, then open pprof interactive shell
profile-cpu: build
	./$(BINARY) -B 200 -iter 100 -cpuprofile cpu.prof
	go tool pprof cpu.prof

## profile-mem: run with memory profiling, then open pprof interactive shell
profile-mem: build
	./$(BINARY) -B 200 -iter 100 -memprofile mem.prof
	go tool pprof mem.prof

## profile-both: capture both CPU and memory profiles non-interactively
profile-both: build
	./$(BINARY) -B 200 -iter 100 -cpuprofile cpu.prof -memprofile mem.prof
	@echo "Profiles written: cpu.prof  mem.prof"
	@echo "Inspect with:"
	@echo "  go tool pprof cpu.prof"
	@echo "  go tool pprof -alloc_space mem.prof"

## svg-cpu: generate a flame-graph SVG from a CPU profile
svg-cpu: cpu.prof
	go tool pprof -svg cpu.prof > cpu_profile.svg
	@echo "SVG written to cpu_profile.svg"

## clean: remove binary and profile files
clean:
	rm -f $(BINARY) cpu.prof mem.prof cpu_profile.svg
