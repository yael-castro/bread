package bread

import (
	"bufio"
	"context"
	"errors"
	"io"
	"sync"
)

// Default Bread parameters
const (
	DefaultDelimiter byte = '\n'
	DefaultWorkers        = 1
)

var (
	ErrNilReader         = errors.New("nil io reader")
	ErrMissingWorkerFunc = errors.New("missing worker function")
	ErrMissingBufferSize = errors.New("missing buffer size")
)

// Bread provides a way to read data line by line an io.Reader
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
	// BufferSize indicates how big will be the buffers instanced
	//
	// This member is required.
	BufferSize uint32
	// Delimiter delimits the end of a line/record, in order to avoid sending half-batches of information.
	//
	// This member is optional. Default value DefaultDelimiter
	Delimiter byte
}

// Eat
//
// NOTE: if the file does not have an end line the file will read completely
func (b Bread) Eat(ctx context.Context, reader io.Reader) (err error) {
	switch {
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

	// Worker settings
	wg := sync.WaitGroup{}
	workerCh := make(chan struct{}, b.Workers)

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

	r := bufio.NewReader(reader)
	n, complement := 0, make([]byte, 0)

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		buffer := pool.Get().(*[]byte)

		n, err = r.Read(*buffer)
		if err != nil {
			if err == io.EOF {
				break
			}

			return
		}

		*buffer = (*buffer)[:n]

		complement, err = r.ReadBytes(b.Delimiter)
		if err != nil && err != io.EOF {
			return err
		}

		*buffer = append(*buffer, complement...)

		workerCh <- struct{}{}
		wg.Add(1)

		go func() {
			defer wg.Done()

			b.WorkerFunc(ctx, buffer)

			pool.Put(buffer)
			<-workerCh
		}()
	}

	close(workerCh)
	wg.Wait()

	return nil
}
