package bgphenet

import (
	"fmt"
	"log/slog"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/anaskhan96/soup"
)

type ASN struct {
	Number       int      `json:"number" bson:"number"`
	Name         string   `json:"name" bson:"name"`
	Whois        string   `json:"whois" bson:"whois"`
	IPv4Networks []string `json:"ipv4_networks" bson:"ipv4_networks"`
	IPv6Networks []string `json:"ipv6_networks" bson:"ipv6_networks"`
	UpdatedAt    int64    `json:"updated_at" bson:"updated_at"`
}

func (a *ASN) LoadIPRanges() error {
	url := url.URL{
		Scheme: "https",
		Host:   "bgp.he.net",
		Path:   fmt.Sprintf("/AS%d", a.Number),
	}
	resp, err := soup.Get(url.String())
	if err != nil {
		slog.Error("Failed to fetch URL: %s", err)
		return err
	}
	doc := soup.HTMLParse(resp)
	if doc.Error != nil {
		slog.Error("Failed to parse HTML: %s", err)
		return err
	}
	// Find Name
	nameDiv := doc.Find("a", "rel", "bookmark")
	junk := fmt.Sprintf("AS%d ", a.Number)
	var name string
	if strings.HasPrefix(nameDiv.Text(), junk) {
		name = strings.TrimPrefix(nameDiv.Text(), junk)
	} else {
		name = nameDiv.Text()
	}
	a.Name = name
	// Find IPv4 prefixes
	ipv4Div := doc.Find("div", "id", "prefixes")
	if ipv4Div.Error == nil {
		ipv4Links := ipv4Div.FindAll("a")
		for _, link := range ipv4Links {
			a.IPv4Networks = append(a.IPv4Networks, link.Text())
		}
	}
	// Find IPv6 prefixes
	ipv6Div := doc.Find("div", "id", "prefixes6")
	if ipv6Div.Error == nil {
		ipv6Links := ipv6Div.FindAll("a")
		for _, link := range ipv6Links {
			a.IPv6Networks = append(a.IPv6Networks, link.Text())
		}
	}
	// Find Whois
	whoisDiv := doc.Find("div", "id", "whois")
	whoisPre := whoisDiv.Find("pre")
	a.Whois = whoisPre.Text()
	// Find Updated
	updatedDiv := doc.Find("div", "id", "footer")
	a.UpdatedAt = time.Time{}.UnixMilli()
	updatedStr := strings.TrimSpace(updatedDiv.Text())
	if strings.HasPrefix(updatedStr, "Updated ") {
		parts := strings.Split(updatedStr, " ©")
		if len(parts) > 0 {
			dateStr := parts[0]
			updatedAt, err := time.Parse("Updated 02 Jan 2006 15:04 MST", dateStr)
			if err != nil {
				slog.Error("error occured while parsing date", slog.String("date", dateStr), slog.String("error", err.Error()))
			} else {
				a.UpdatedAt = updatedAt.UnixMilli()
			}
		}
	}
	return nil
}

func (a *ASN) String() string {
	return fmt.Sprintf("AS%d, IPv4: %v, IPv6: %v", a.Number, a.IPv4Networks, a.IPv6Networks)
}

func NewASN(number int, load bool) (*ASN, error) {
	asn := &ASN{
		Number:       number,
		IPv4Networks: []string{},
		IPv6Networks: []string{},
	}
	if load {
		err := asn.LoadIPRanges()
		if err != nil {
			return nil, err
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

func NewSearch(keyword string) *Search {
	s := &Search{
		Name: keyword,
		ASNs: []*ASN{},
	}
	for asn := range search(s.Name) {
		s.ASNs = append(s.ASNs, asn)
	}
	return s
}

// SearchASN searches for an ASN by keyword and returns a channel of ASN numbers
func SearchASN(keyword string) chan int {
	return parseASN(buildUrl(keyword))
}

func search(keyword string) chan *ASN {
	out := make(chan *ASN)
	go func() {
		defer close(out)
		for asn := range parseASN(buildUrl(keyword)) {
			as, err := NewASN(asn, true)
			if err != nil {
				slog.Error("Failed to create ASN: %s", err)
				continue
			}
			out <- as
		}
	}()
	return out
}

func buildUrl(keyword string) string {
	url := url.URL{
		Scheme: "https",
		Host:   "bgp.he.net",
		Path:   "/search",
	}
	query := url.Query()
	query.Set("search[search]", keyword)
	query.Set("commit", "Search")
	url.RawQuery = query.Encode()
	return url.String()
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
