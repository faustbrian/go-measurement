# Performance

Cost is dominated by arbitrary-precision coefficient alignment,
multiplication, and division. Finite SI conversions allocate proportionally to
decimal size. Non-terminating rounded conversions perform one quantized
division of the combined ratio, avoiding intermediate double rounding.

Prefer canonical units inside large homogeneous batches and convert only at
ingress or egress. Reuse immutable profiles and conversion contexts safely
across goroutines. Keep input limits proportional to actual carrier payloads.

`go test ./... -run '^$' -bench Benchmark -benchmem -benchtime=100ms` records
allocations for conversion and loading-metre paths. `make check-full` retains
the configured benchmark campaign; ordinary local and PR checks do not run it.
Benchmarks are regression signals, not cross-machine service-level guarantees.
No hidden cache trades memory for speed.
