# Golang Bindings for `bgp.he.net`

## Usage

```bash
go get https://github.com/WangYihang/bgphenet
```

```golang
package main

import (
    "fmt"
    "github.com/WangYihang/bgphenet"
)

func main() {
    s := bgphenet.Search("cloudflare")
    fmt.Println(s)
}
```