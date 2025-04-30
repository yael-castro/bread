package bread

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strconv"
	"testing"
)

func TestBread_Eat(t *testing.T) {
	cases := [...]struct {
		ctx         context.Context
		bread       Bread
		reader      io.Reader
		expectedErr error
	}{
		// Test case: nil context
		{
			ctx:         nil,
			bread:       Bread{},
			expectedErr: ErrMissingContext,
		},
		// Test case: nil reader
		{
			ctx:         context.TODO(),
			bread:       Bread{},
			expectedErr: ErrNilReader,
		},
		// Test case: missing worker
		{
			ctx:         context.TODO(),
			bread:       Bread{},
			expectedErr: ErrNilReader,
		},
		// Test case: success!
		{
			ctx: context.TODO(),
			bread: Bread{
				Workers:    16,
				WorkerFunc: func(ctx context.Context, buffer *[]byte) {},
				BufferSeed: 5,
				BufferSize: 1 * MB,
			},
			reader: buffer(300 * MB),
		},
	}

	for i, v := range cases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			err := v.bread.Eat(v.ctx, v.reader)
			if !errors.Is(err, v.expectedErr) {
				t.Fatalf("expected error %v, got %v", v.expectedErr, err)
			}
		})
	}
}

func BenchmarkBread_Eat(b *testing.B) {
	cases := [...]struct {
		ctx     context.Context
		myBread Bread
		reader  io.Reader
	}{
		// Scenario: Consuming 5 GB using 16 MB as maximum
		{
			ctx: context.TODO(),
			myBread: Bread{
				Workers:    16,
				WorkerFunc: func(context.Context, *[]byte) {},
				BufferSeed: 0,
				BufferSize: MB,
				Delimiter:  'X',
			},
			reader: buffer(5 * GB),
		},
		// Scenario: Consuming 1 GB using 3 MB of memory
		{
			ctx: context.TODO(),
			myBread: Bread{
				Workers:    3,
				WorkerFunc: func(context.Context, *[]byte) {},
				BufferSeed: 3,
				BufferSize: MB,
				Delimiter:  'X',
			},
			reader: buffer(GB),
		},
	}

	b.ResetTimer()

	for i, c := range cases {
		b.Run(strconv.Itoa(i), func(b *testing.B) {
			b.ResetTimer()

			for n := 0; n < b.N; n++ {
				err := c.myBread.Eat(c.ctx, c.reader)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func buffer(size int) io.Reader {
	data := make([]byte, size)

	for i := 0; i < len(data); i++ {
		if i%300 == 0 {
			data[i] = 'X'
			continue
		}

		data[i] = 'O'
	}

	return bytes.NewBuffer(data)
}
