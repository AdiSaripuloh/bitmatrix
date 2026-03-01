.PHONY: test test-race bench bench-core bench-sizes lint cover fuzz clean

test:
	go test -v -count=1 ./...

test-race:
	go test -race -count=1 ./...

bench:
	go test -bench=. -benchmem -benchtime=3s -run='^$$' ./...

bench-core:
	go test -bench='^Benchmark[^S]' -benchmem -benchtime=3s -run='^$$' ./...

bench-sizes:
	go test -bench='^BenchmarkSizes' -benchmem -benchtime=3s -run='^$$' ./...

lint:
	golangci-lint run ./...

cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out
	@echo ""
	@echo "To view HTML report: go tool cover -html=coverage.out"

fuzz:
	go test -fuzz=FuzzSetHasRoundTrip -fuzztime=10s ./...
	go test -fuzz=FuzzClearRoundTrip -fuzztime=10s ./...
	go test -fuzz=FuzzMultiBitSetGetRoundTrip -fuzztime=10s ./...

clean:
	rm -f coverage.out coverage.html
	rm -rf testdata/fuzz/
