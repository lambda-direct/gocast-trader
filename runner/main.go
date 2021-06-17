package main

import (
	"encoding/binary"
	"log"
	"math"
	"os"

	lua "github.com/yuin/gopher-lua"
)

func walletBuy(L *lua.LState) int {
	n := L.ToInt(1)
	log.Printf("wallet_buy invoked with %d", n)
	return 0
}

func main() {
	L := lua.NewState()
	if err := L.DoFile("test.lua"); err != nil {
		panic(err)
	}

	f, err := os.Open("../ticker/data/BTCUSDT/17062021.bin")
	if err != nil {
		panic(err)
	}

	//stats, err := f.Stat()
	//if err != nil {
	//	panic(err)
	//}

	//pointCount := int(stats.Size() / 16)

	for i := 0; i < 10; i++ {
		b := make([]byte, 16)
		if _, err := f.Read(b); err != nil {
			panic(err)
		}

		price := math.Float64frombits(binary.LittleEndian.Uint64(b[:8]))
		ts := binary.LittleEndian.Uint64(b[8:])

		if err := L.CallByParam(lua.P{
			Fn:      L.GetGlobal("step"),
			NRet:    1,
			Protect: true,
		}, lua.LString("BTCUSDT"), lua.LNumber(price), lua.LNumber(ts)); err != nil {
			panic(err)
		}
	}
}
