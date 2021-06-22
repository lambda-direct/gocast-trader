package persister

import (
	"context"
	"encoding/binary"
	"fmt"
	"github.com/lambda-direct/gocast-trader/ticker/src/lock"
	"math"
	"os"
	"sync"
	"time"

	"github.com/lambda-direct/gocast-trader/common/src/env"
	"github.com/lambda-direct/gocast-trader/ticker/src/fetcher"
)

type Client struct {
	ctx     context.Context
	fetcher *fetcher.Client
	s       *env.Spec
	l       *lock.Lock
}

func New(ctx context.Context, f *fetcher.Client, s *env.Spec, l *lock.Lock) *Client {
	return &Client{ctx, f, s, l}
}

func (c *Client) Watch(errc chan<- error) {
	c.l.Acquire(errc)

	for {
		select {
		case <-c.ctx.Done():
			return
		default:
		}

		start := time.Now()
		latest := c.fetcher.Latest
		var wg sync.WaitGroup

		wg.Add(len(latest))

		for _, pair := range latest {
			go func(pair fetcher.BinanceResponseItem) {
				defer wg.Done()

				var err error
				now := time.Now()

				dirPath := fmt.Sprintf("%s/%s", c.s.DataDir, pair.Symbol)
				fileName := fmt.Sprintf("%s/%s.bin", dirPath, now.Format("02012006"))

				if err := os.MkdirAll(dirPath, os.ModePerm); err != nil {
					errc <- fmt.Errorf("unable to create directories: %w", err)
					return
				}

				f, err := os.OpenFile(fileName, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)

				if err != nil {
					errc <- fmt.Errorf("unable to open data file: %w", err)
					return
				}
				defer f.Close()

				priceBytes := make([]byte, 16)
				binary.LittleEndian.PutUint64(priceBytes[:8], math.Float64bits(pair.Price))
				binary.LittleEndian.PutUint64(priceBytes[8:], uint64(now.Unix()))

				if _, err := f.Write(priceBytes); err != nil {
					errc <- fmt.Errorf("unable to write bytes to file: %w", err)
					return
				}
			}(pair)
		}

		wg.Wait()

		//log.Printf("persister iteration took %.3fms and stored %d pair(s)", float64(time.Since(start).Microseconds())/1000, len(latest))
		time.Sleep(1*time.Second - time.Since(start))
	}
}
