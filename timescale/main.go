package main

import (
	"context"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"time"

	"github.com/jackc/pgx/v4"
)

type DataPoint struct {
	Price     float64
	Timestamp int64
}

func main() {
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, "postgres://codecast:codecast@localhost:5432/postgres")
	if err != nil {
		panic(err)
	}

	defer conn.Close(ctx)

	err = conn.Ping(ctx)
	if err != nil {
		panic(err)
	}

	pair := "BTCUSDT"
	files, err := ioutil.ReadDir(fmt.Sprintf("/media/dan/Data_SSD_2TB/ticker/data/%s", pair))
	if err != nil {
		panic(err)
	}

	var data []DataPoint

	for _, fileInDir := range files {
		fmt.Printf("Inserting from %s/%s\n", pair, fileInDir.Name())

		f, err := os.Open(fmt.Sprintf("/media/dan/Data_SSD_2TB/ticker/data/%s/%s", pair, fileInDir.Name()))
		if err != nil {
			panic(err)
		}

		stat, err := f.Stat()
		if err != nil {
			panic(err)
		}

		size := stat.Size()
		numPairs := size / 16

		buf := make([]byte, size)

		f.Read(buf)

		data = make([]DataPoint, numPairs)

		for i := int64(0); i < numPairs; i++ {
			offset := i * 16
			price := math.Float64frombits(binary.LittleEndian.Uint64(buf[offset : offset+8]))
			ts := int64(binary.LittleEndian.Uint64(buf[offset+8 : offset+16]))

			data[i] = DataPoint{
				Price:     price,
				Timestamp: ts,
			}
		}

		const chunkSize = 1000
		numChunks := int(math.Ceil(float64(len(data)) / chunkSize))

		for i := 0; i < numChunks; i++ {
			var vals [][]interface{}

			for k := i * chunkSize; k < int(math.Min(float64(len(data)), float64((i+1)*chunkSize))); k++ {
				vals = append(vals, []interface{}{"BTC", data[k].Price, time.Unix(data[k].Timestamp, 0)})
			}

			copyCount, err := conn.CopyFrom(
				ctx,
				pgx.Identifier{"tickers"},
				[]string{"base", "price", "time"},
				pgx.CopyFromRows(vals),
			)
			if err != nil {
				panic(err)
			}
			fmt.Printf("inserted %d rows\n", copyCount)
		}
	}
}
