package printer

import (
	"encoding/binary"
	"fmt"
	"github.com/lambda-direct/gocast-trader/fetcher"
	"log"
	"math"
	"os"
	"time"
)

type Client struct {
}

func New() *Client {
	return new(Client)
}

func (c *Client) Loop(errc chan<- error) {
	time.Sleep(2 * time.Second)

	for {
		start := time.Now()

		dirs, err := os.ReadDir("data")
		if err != nil {
			errc <- fmt.Errorf("unable to read data dir: %w", err)
			return
		}

		for _, entry := range dirs[:1] {
			if !entry.IsDir() {
				continue
			}

			symbol := entry.Name()

			tickerFiles, err := os.ReadDir(fmt.Sprintf("data/%s", entry.Name()))

			if err != nil {
				errc <- fmt.Errorf("unable to list files in ticker dir")
				return
			}

			for _, tickerFilesEntry := range tickerFiles {
				if tickerFilesEntry.IsDir() {
					continue
				}

				f, err := os.Open(fmt.Sprintf("data/%s/%s", entry.Name(), tickerFilesEntry.Name()))
				if err != nil {
					errc <- fmt.Errorf("unable to open data file: %w", err)
					return
				}

				fileStat, err := f.Stat()
				if err != nil {
					errc <- fmt.Errorf("unable to get data file stats: %w", err)
					return
				}

				fileSize := fileStat.Size()
				pointsCount := fileSize / 16

				points := make([]*fetcher.BinanceResponseItem, pointsCount)
				bytes := make([]byte, 16)
				sum := 0.0

				for i := range points {
					if _, err := f.Read(bytes); err != nil {
						errc <- fmt.Errorf("unable to read from data file: %w", err)
						return
					}

					price := math.Float64frombits(binary.LittleEndian.Uint64(bytes[:8]))
					// ts := int64(binary.LittleEndian.Uint64(bytes[8:]))

					points[i] = &fetcher.BinanceResponseItem{
						Symbol: symbol,
						Price:  price,
					}

					sum += price
				}

				open := points[0].Price
				close := points[len(points)-1].Price
				avg := sum / float64(len(points))

				log.Printf("%s %.6f %.6f %.6f\n", symbol, open, close, avg)
			}
		}

		time.Sleep(1*time.Second - time.Since(start))
	}
}
