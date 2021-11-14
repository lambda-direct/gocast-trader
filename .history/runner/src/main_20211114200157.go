package main

import (
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"os"
	"sync"

	lua "github.com/yuin/gopher-lua"
)

func walletBuy(L *lua.LState) int {
	n := L.ToInt(1)
	log.Printf("wallet_buy invoked with %d", n)
	return 0
}

type Wallet struct {
	Pair          string
	BalanceFiat   float64
	BalanceCrypto float64
	Generator     *rand.Rand
}

func (w *Wallet) Swap(price float64) {
	if w.BalanceCrypto == 0 {
		w.Buy(price)
	} else {
		w.Sell(price)
	}
}

func (w *Wallet) Buy(price float64) {
	w.BalanceCrypto = w.BalanceFiat / price
	w.BalanceFiat = 0
}

func (w *Wallet) Sell(price float64) {
	w.BalanceFiat = w.BalanceCrypto * price
	w.BalanceCrypto = 0
}

func (w *Wallet) Stats(price float64) float64 {
	return w.BalanceFiat + w.BalanceCrypto*price
}

type DataPoint struct {
	Price     float64
	Timestamp int64
}

func main() {
	pair := "BTCUSDT"
	files, err := ioutil.ReadDir(fmt.Sprintf("/media/dan/My_Passport_4TB/ticker/data/%s", pair))
	if err != nil {
		panic(err)
	}

	wallets := make([]*Wallet, 100)
	for i := 0; i < len(wallets); i++ {
		wallets[i] = &Wallet{
			Pair:        "BTCUSDT",
			BalanceFiat: 1000,
			Generator:   rand.New(rand.NewSource(int64(i))),
		}
	}

	var data []DataPoint

	for fileIndex, fileInDir := range files {
		f, err := os.Open(fmt.Sprintf("/media/dan/My_Passport_4TB/ticker/data/%s/%s", pair, fileInDir.Name()))
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

		data = make([]DataPoint, 0)

		for i := int64(0); i < numPairs; i++ {
			offset := i * 16
			price := math.Float64frombits(binary.LittleEndian.Uint64(buf[offset : offset+8]))
			ts := int64(binary.LittleEndian.Uint64(buf[offset+8 : offset+16]))

			if ts%10 != 0 {
				continue
			}

			data = append(data, DataPoint{
				Price:     price,
				Timestamp: ts,
			})
		}

		var wg sync.WaitGroup
		wg.Add(len(wallets))

		for i := 0; i < len(wallets); i++ {
			go func(i int) {
				defer wg.Done()

				wallet := wallets[i]

				for dataPointIndex := 0; dataPointIndex < len(data); dataPointIndex++ {
					dataPoint := data[dataPointIndex]

					action := wallet.Generator.Intn(2) == 0
					if action {
						wallet.Swap(dataPoint.Price)
					}
				}
			}(i)
		}

		wg.Wait()

		fmt.Printf("File %s processed (%d/%d)\n", fileInDir.Name(), fileIndex+1, len(files))
	}

	for i := 0; i < len(wallets); i++ {
		stats := wallets[i].Stats(data[len(data)-1].Price)
		fmt.Printf("#%d - %.2f\n", i, stats)
	}

	// fmt.Printf("balance percentage: %.2f%%\n", (balance-float64(1000*len(wallets)))/balance*100)
}
