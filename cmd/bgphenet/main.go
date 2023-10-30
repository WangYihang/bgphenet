package main

import (
	"encoding/json"
	"os"

	"github.com/WangYihang/bgphenet"
	"github.com/jessevdk/go-flags"
)

type Options struct {
	Keyword string `short:"k" long:"keyword" description:"keyword to search" required:"true"`
}

var opts Options

func init() {
	_, err := flags.Parse(&opts)
	if err != nil {
		os.Exit(1)
	}
}

func main() {
	search := bgphenet.NewSearch(opts.Keyword)
	data, err := json.Marshal(search)
	if err != nil {
		panic(err)
	}
	println(string(data))
}
