package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"

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

func (w *Wallet) Buy(price float64) {
	w.BalanceCrypto = w.BalanceFiat / price
	w.BalanceFiat = 0
}

func (w *Wallet) Sell(price float64) {
	w.BalanceFiat = w.BalanceCrypto * price
	w.BalanceCrypto = 0
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
	// numPairs := size / 16

	buf := make([]byte, size)

	f.Read(buf)

	for i := 0; i < 100; i++ {
		offset := i * 16
		price := math.Float64frombits(binary.LittleEndian.Uint64(buf[offset : offset+8]))
		ts := int64(binary.LittleEndian.Uint64(buf[offset : offset+8]))
		fmt.Printf("price: %f; timestamp: %d\n", price, ts)
	}

	// var wallets [100]*Wallet
	// for i := 0; i < 100; i++ {
	// 	wallet := &Wallet{
	// 		Pair:        "BTCUSDT",
	// 		BalanceFiat: 1000,
	// 		Generator:   rand.New(rand.NewSource(int64(i))),
	// 	}

	// }
}
