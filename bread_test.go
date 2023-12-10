package bread

import (
	"bytes"
	"context"
	"io"
	"strconv"
	"testing"
)

const MB = 1 << 20

func TestBread_Eat(t *testing.T) {
	cases := [...]struct {
		ctx    context.Context
		bread  Bread
		reader io.Reader
	}{
		{
			ctx: context.TODO(),
			bread: Bread{
				Workers:    16,
				WorkerFunc: func(ctx context.Context, buffer *[]byte) {},
				BufferSize: 1 * MB,
			},
			reader: memoryReader(300 * MB),
		},
	}

	for i, v := range cases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			err := v.bread.Eat(v.ctx, v.reader)
			if err != nil {
				t.Fatal(err)
			}

			t.Log("Success!")
		})
	}
}

func BenchmarkBread_Eat(b *testing.B) {
	cases := [...]struct {
		ctx    context.Context
		bread  Bread
		reader io.Reader
	}{
		{
			ctx: context.TODO(),
			bread: Bread{
				Workers:    16,
				WorkerFunc: func(context.Context, *[]byte) {},
				BufferSize: MB,
				BufferSeed: 0,
			},
			reader: memoryReader(5_000 * MB),
		},
	}

	for i, c := range cases {
		b.Run(strconv.Itoa(i), func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				err := c.bread.Eat(c.ctx, c.reader)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func memoryReader(size int) io.Reader {
	data := make([]byte, size)

	for i := 0; i < len(data); i++ {
		if i%300 == 0 {
			data[i] = DefaultDelimiter
			continue
		}

		data[i] = 'O'
	}

	return bytes.NewBuffer(data)
}
