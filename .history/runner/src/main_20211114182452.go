package main

import (
	"log"
	"math/rand"

	lua "github.com/yuin/gopher-lua"
)

func walletBuy(L *lua.LState) int {
	n := L.ToInt(1)
	log.Printf("wallet_buy invoked with %d", n)
	return 0
}

type Wallet struct {
	pair           string
	balance_fiat   float64
	balance_crypto float64
}

func (w *Wallet) Buy(price float64) {
	w.balance_crypto = w.balance_fiat / price
	w.balance_fiat = 0
}

func (w *Wallet) Sell(price float64) {
	w.balance_fiat = w.balance_crypto * price
	w.balance_crypto = 0
}

func main() {
	var generators [100]*rand.Rand
	for i := 0; i < len(generators); i++ {
		generators[i] = rand.New(rand.NewSource(int64(i)))
	}

}
