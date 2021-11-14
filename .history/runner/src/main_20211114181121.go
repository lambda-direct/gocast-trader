package main

import (
	"log"
	"math/rand"
	"os"

	lua "github.com/yuin/gopher-lua"
)

func walletBuy(L *lua.LState) int {
	n := L.ToInt(1)
	log.Printf("wallet_buy invoked with %d", n)
	return 0
}

func main() {
	for i := 0; i < 100; i++ {

	}
	generator := rand.New(rand.NewSource(123))

	f, err := os.Open("../ticker/data/BTCUSDT/17062021.bin")
	if err != nil {
		panic(err)
	}
}
