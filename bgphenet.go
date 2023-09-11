package bgphenet

import (
	"fmt"
	"log/slog"
	"net/url"
	"strconv"
	"strings"

	"github.com/anaskhan96/soup"
)

type ASN struct {
	Number       int      `json:"number" bson:"number"`
	IPv4Networks []string `json:"ipv4_networks" bson:"ipv4_networks"`
	IPv6Networks []string `json:"ipv6_networks" bson:"ipv6_networks"`
}

func (a *ASN) String() string {
	return fmt.Sprintf("AS%d, IPv4: %v, IPv6: %v", a.Number, a.IPv4Networks, a.IPv6Networks)
}

func NewASN(number int) (*ASN, error) {
	asn := &ASN{
		Number:       number,
		IPv4Networks: []string{},
		IPv6Networks: []string{},
	}
	url := url.URL{
		Scheme: "https",
		Host:   "bgp.he.net",
		Path:   fmt.Sprintf("/AS%d", number),
	}
	resp, err := soup.Get(url.String())
	if err != nil {
		slog.Error("Failed to fetch URL: %s", err)
		return nil, err
	}
	doc := soup.HTMLParse(resp)
	if doc.Error != nil {
		slog.Error("Failed to parse HTML: %s", err)
		return nil, err
	}
	// Find IPv4 prefixes
	ipv4div := doc.Find("div", "id", "prefixes")
	if ipv4div.Error == nil {
		ipv4links := ipv4div.FindAll("a")
		for _, link := range ipv4links {
			asn.IPv4Networks = append(asn.IPv4Networks, link.Text())
		}
	}
	// Find IPv6 prefixes
	ipv6div := doc.Find("div", "id", "prefixes6")
	if ipv6div.Error == nil {
		ipv6links := ipv6div.FindAll("a")
		for _, link := range ipv6links {
			asn.IPv6Networks = append(asn.IPv6Networks, link.Text())
		}
	}
	return asn, nil
}

type Search struct {
	Name string `json:"name" bson:"name"`
	ASNs []*ASN `json:"asns" bson:"asns"`
}

func (c *Search) String() string {
	return fmt.Sprintf("%s: %v", c.Name, c.ASNs)
}

func NewSearch(name string) *Search {
	s := &Search{
		Name: name,
		ASNs: []*ASN{},
	}
	for asn := range search(s.Name) {
		s.ASNs = append(s.ASNs, asn)
	}
	return s
}

func search(keyword string) chan *ASN {
	out := make(chan *ASN)
	go func() {
		defer close(out)
		url := url.URL{
			Scheme: "https",
			Host:   "bgp.he.net",
			Path:   "/search",
		}
		query := url.Query()
		query.Set("search[search]", keyword)
		query.Set("commit", "Search")
		url.RawQuery = query.Encode()
		for asn := range parseASN(url.String()) {
			as, err := NewASN(asn)
			if err != nil {
				slog.Error("Failed to create ASN: %s", err)
				continue
			}
			out <- as
		}
	}()
	return out
}

func parseASN(url string) chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		resp, err := soup.Get(url)
		if err != nil {
			slog.Error("Failed to fetch URL: %s", err)
			return
		}
		doc := soup.HTMLParse(resp)
		if doc.Error != nil {
			slog.Error("Failed to parse HTML: %s", err)
			return
		}
		searchdiv := doc.Find("div", "id", "search")
		if searchdiv.Error != nil {
			slog.Error("Failed to find search div: %s", err)
			return
		}
		links := searchdiv.FindAll("a")
		for _, link := range links {
			if strings.HasPrefix(link.Text(), "AS") {
				rawNumber := link.Attrs()["href"][3:]
				number, err := strconv.Atoi(rawNumber)
				if err != nil {
					slog.Error("Failed to parse ASN number: %s", err)
					continue
				}
				out <- number
			}
		}
	}()
	return out
}
