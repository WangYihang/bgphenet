# GoLang bindings for `bgp.he.net`

## Usage

```bash
go get github.com/WangYihang/bgphenet
```

```golang
package main

import (
	"fmt"

	"github.com/WangYihang/bgphenet"
)

func main() {
	// Search
	s := bgphenet.NewSearch("cloudflare")
	fmt.Println(s.ASNs)

	// Get IPv4 and IPv6 prefixes
	asn, err := bgphenet.NewASN(395747)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(asn.IPv4Networks)
	fmt.Println(asn.IPv6Networks)
}
```
