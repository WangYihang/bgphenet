package main

import (
	"fmt"

	"github.com/WangYihang/bgp.he.net/pkg/util"
)

func main() {
	s := util.NewSearch("cloudflare")
	fmt.Println(s)
}
