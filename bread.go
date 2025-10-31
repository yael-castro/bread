package bread

import (
	"bufio"
	"context"
	"errors"
	"io"
	"sync"
)

const (
	// B bytes
	B = 1 << (10 * iota)
	// KB kilobytes
	KB
	// MB megabytes
	MB
	// GB gigabytes
	GB
)

// Default Bread parameters
const (
	DefaultDelimiter byte = '\n'
	DefaultWorkers        = 1
)

var (
	ErrNilReader         = errors.New("nil io reader")
	ErrMissingContext    = errors.New("missing context")
	ErrMissingWorkerFunc = errors.New("missing worker function")
	ErrMissingBufferSize = errors.New("missing buffer size")
)

// Bread provides a way to read large amounts of data concurrently and
// in a highly optimized way (optimize performance and memory usage)
type Bread struct {
	// WorkerFunc is used to process each data batch from the io.Reader
	//
	// This member is required.
	WorkerFunc func(context.Context, *[]byte)
	// Workers number of concurrent workers
	//
	// This member is optional. Default value DefaultWorkers
	Workers uint32
	// BufferSeed indicates the initial reservation of available buffer instances in the object pool
	//
	// This member is optional. Default value 0
	BufferSeed uint32
	// BufferSize indicates the initial length and capacity of the byte batches (buffers)
	//
	// This member is required.
	BufferSize uint32
	// Delimiter delimits the end of a line/record, in order to avoid sending half-batches of information.
	// For example.
	//
	// If the BufferSize is equal to 3, the io.Reader contains "Hello|Hello|Hello" and the Delimiter is "|"
	// We have read until the "\n" and the outputs will be "Hello|", "Hello|", "Hello"
	//
	// This member is optional. Default value DefaultDelimiter
	Delimiter byte
	// NoDelimiter indicates if the delimiter should be ignored.
	//
	// If this value is true, the BufferSize indicates the maximum length
	// that byte batches can have.
	//
	// If this member is true, the BufferSize indicates the maximum length
	// that the batches of bytes can be.
	//
	// This member is optional. Default value is false
	NoDelimiter bool
}

// Eat helps to concurrently process the io.Reader in batches.
// Eat method ends until all workers finish or the context is cancelled
//
// WARNING The Eat method is not concurrently safe if io.Reader is not.
// WARNING If there is no Delimiter and NoDelimiter is false, the io.Reader will be read completely.
func (b Bread) Eat(ctx context.Context, reader io.Reader) error {
	switch {
	case ctx == nil:
		return ErrMissingContext
	case reader == nil:
		return ErrNilReader
	case b.WorkerFunc == nil:
		return ErrMissingWorkerFunc
	case b.BufferSize == 0:
		return ErrMissingBufferSize
	}

	if b.Delimiter == 0 {
		b.Delimiter = DefaultDelimiter
	}

	if b.Workers == 0 {
		b.Workers = DefaultWorkers
	}

	err := b.eat(ctx, reader)
	if err != nil && !errors.Is(err, io.EOF) {
		return err
	}

	return nil
}

func (b Bread) eat(ctx context.Context, reader io.Reader) (err error) {
	// Object pool in charge of handling buffers
	pool := sync.Pool{
		New: func() any {
			buffer := make([]byte, b.BufferSize, b.BufferSize*2)
			return &buffer
		},
	}

	// Initial reservation of available buffer instances in the object pool
	for seed := b.BufferSeed; seed > 0; seed-- {
		pool.Put(pool.New())
	}

	// Concurrency management
	wg := sync.WaitGroup{}

	workerChan := make(chan struct{}, b.Workers)
	defer close(workerChan)

	// Listening done signal
	doneChan := make(chan struct{}, 1)
	defer close(doneChan)

	go func() {
		wg.Wait()
		doneChan <- struct{}{}
	}()

	// Objects to manage the buffered data
	r := bufio.NewReader(reader)
	n, complement := 0, make([]byte, 0)

	for {
		select {
		case <-ctx.Done():
			break
		default:
		}

		buffer := pool.Get().(*[]byte)

		n, err = r.Read(*buffer)
		if err != nil {
			break
		}

		*buffer = (*buffer)[:n]

		if !b.NoDelimiter {
			complement, err = r.ReadBytes(b.Delimiter)
			if err != nil && !errors.Is(err, io.EOF) {
				break
			}

			*buffer = append(*buffer, complement...)
		}

		workerChan <- struct{}{}
		wg.Add(1)

		go func() {
			defer wg.Done()

			b.WorkerFunc(ctx, buffer)

			pool.Put(buffer)
			<-workerChan
		}()
	}

	// Waiting for batch processing to complete, an error, or context cancellation.
	select {
	case <-ctx.Done():
		<-doneChan
		return ctx.Err()
	case <-doneChan:
		return
	}
}
