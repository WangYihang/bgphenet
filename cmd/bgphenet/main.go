package main

import (
	"encoding/json"

	"github.com/WangYihang/bgphenet"
)

func main() {
	asn, err := bgphenet.NewASN(15169, false)
	if err != nil {
		panic(err)
	}
	asn.LoadIPRanges()
	data, err := json.Marshal(asn)
	if err != nil {
		panic(err)
	}
	println(string(data))
}
