# Bread
Go library created to manage large amounts of data (Gigabytes) to be processed in batches optimizing performance and memory usage.
### Installation
```shell
go get github.com/yael-castro/bread@latest
```
### How to use
```go
package main

import (
    "bytes"
    "context"
    "github.com/yael-castro/bread"
    "log"
    "runtime"
)

func main() {
    reader := bytes.NewBuffer(make([]byte, bread.GB))

    b := bread.Bread{
        Workers: uint32(runtime.NumCPU()),
        WorkerFunc: func(ctx context.Context, buffer *[]byte) {
        },
        BufferSeed: 5,
        BufferSize: bread.MB,
    }

    err := b.Eat(context.TODO(), reader)
    if err != nil {
        log.Fatal(err)
    }
}
```
### Performance
Run the benchmarks and build the runtime profile
```shell
make pprof
```
See the memory profile
```shell
make mem_prof
```
See the cpu profile
```shell
make cpu_prof
```