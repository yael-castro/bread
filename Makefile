.PHONY: pprof mem_prof cpu_prof

pprof:
	go test -bench . -cpuprofile=cpu.pprof -memprofile=mem.pprof -run ^$

mem_prof:
	go tool pprof -http=:8080 mem.pprof

cpu_prof:
	go tool pprof -http=:8080 cpu.pprof

