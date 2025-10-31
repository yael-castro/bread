.PHONY: pprof mem_prof cpu_prof

pprof:
	go test -bench . -cpuprofile=cpu.pprof -memprofile=mem.pprof -run ^$

mem_prof:
	go tool pprof -http=:8080 mem.pprof

benchmark:
	go test . -o "bench_mem.test" -run ^# -bench . -benchmem -count=10 -timeout 60m -v

cpu_prof:
	go tool pprof -http=:8080 cpu.pprof

