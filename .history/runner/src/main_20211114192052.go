package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"sync"
	"time"

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

type DataPoint struct {
	Price     float64
	Timestamp int64
}

func main() {
	pair := "BTCUSDT"
	f, err := os.Open(fmt.Sprintf("/media/dan/My_Passport_4TB/ticker/data/%s/25062021.bin", pair))
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

	data := make([]DataPoint, numPairs)

	for i := int64(0); i < numPairs; i++ {
		offset := i * 16
		price := math.Float64frombits(binary.LittleEndian.Uint64(buf[offset : offset+8]))
		ts := int64(binary.LittleEndian.Uint64(buf[offset+8 : offset+16]))

		if ts%10 != 0 {
			continue
		}

		data[i] = DataPoint{
			Price:     price,
			Timestamp: ts,
		}
	}

	wallets := make([]Wallet, 100)
	for i := 0; i < len(wallets); i++ {
		wallets[i] = Wallet{
			Pair:        "BTCUSDT",
			BalanceFiat: 1000,
			Generator:   rand.New(rand.NewSource(int64(i))),
		}
	}

	start := time.Now()

	var wg sync.WaitGroup
	wg.Add(len(wallets))

	for i := 0; i < len(wallets); i++ {
		go func(i int) {
			defer wg.Done()

			wallet := wallets[i]

			for dataPointIndex := int64(0); dataPointIndex < numPairs; dataPointIndex++ {
				dataPoint := data[dataPointIndex]

				action := wallet.Generator.Intn(2) == 0
				if action {
					wallet.Swap(dataPoint.Price)
				}
			}
		}(i)

		wg.Wait()
	}

	fmt.Printf("took %d ms\n", time.Since(start).Milliseconds())
}