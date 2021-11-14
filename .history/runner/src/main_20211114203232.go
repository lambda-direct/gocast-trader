package main

import (
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"os"
	"sort"
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
	// if w.BalanceFiat >= 1200 {
	// 	return
	// }
	w.BalanceCrypto = w.BalanceFiat / price * .999
	w.BalanceFiat = 0
}

func (w *Wallet) Sell(price float64) {
	w.BalanceFiat = w.BalanceCrypto * price * .999
	w.BalanceCrypto = 0
}

func (w *Wallet) Balance(price float64) float64 {
	return w.BalanceFiat + w.BalanceCrypto*price
}

type DataPoint struct {
	Price     float64
	Timestamp int64
}

type ResultStats struct {
	StrategyIndex int
	Balance       float64
}

const INITIAL_WALLET_BALANCE = 100

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
			BalanceFiat: INITIAL_WALLET_BALANCE,
			Generator:   rand.New(rand.NewSource(int64(i))),
		}
	}

	var data []DataPoint

	for fileIndex, fileInDir := range files {
		if fileIndex%2 == 0 {
			continue
		}

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

			if ts%(4*3600) != 0 {
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

		// fmt.Printf("File %s processed (%d/%d)\n", fileInDir.Name(), fileIndex+1, len(files))

		total := float64(0)

		for i := 0; i < len(wallets); i++ {
			balance := wallets[i].Balance(data[len(data)-1].Price)
			total += balance
		}

		initialBalance := float64(INITIAL_WALLET_BALANCE * len(wallets))
		fmt.Printf("balance percentage: %.2f%%\n", total/initialBalance*100)
	}

	total := float64(0)

	results := make([]ResultStats, len(wallets))

	for i := 0; i < len(wallets); i++ {
		balance := wallets[i].Balance(data[len(data)-1].Price)
		total += balance
		results[i] = ResultStats{
			StrategyIndex: i,
			Balance:       balance,
		}
	}

	sort.SliceStable(results, func(i, j int) bool {
		return results[i].Balance < results[j].Balance
	})

	// for i := range results {
	// 	fmt.Printf("%d\t%.2f\n", results[i].StrategyIndex, results[i].Balance)
	// }

	// initialBalance := float64(INITIAL_WALLET_BALANCE * len(wallets))
	// fmt.Printf("balance percentage: %.2f%%\n", total/initialBalance*100)
}
