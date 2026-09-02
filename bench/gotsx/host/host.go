// Package host exposes the product list to the gotsx bench page (import { products } from "host:data").
package host

import (
	"encoding/json"
	"os"

	gotsx "github.com/childrentime/gotsx/runtime"
)

type Product struct {
	ID     string   `json:"id"`
	Title  string   `json:"title"`
	Brand  string   `json:"brand"`
	Desc   string   `json:"desc"`
	Price  int      `json:"price"`
	Tags   []string `json:"tags"`
	Rating float64  `json:"rating"`
}

type DataModule struct{ items []Product }

func (d *DataModule) Products() []Product { return d.items }
func (d *DataModule) Money(cents float64) string {
	c := int(cents)
	return "$" + itoa(c/100) + "." + pad2(c%100)
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	var b [12]byte
	n := len(b)
	for i > 0 {
		n--
		b[n] = byte('0' + i%10)
		i /= 10
	}
	return string(b[n:])
}
func pad2(i int) string {
	if i < 10 {
		return "0" + itoa(i)
	}
	return itoa(i)
}

func load() []Product {
	var items []Product
	b, _ := os.ReadFile("../data/products.json")
	json.Unmarshal(b, &items)
	return items
}

var Data = &DataModule{items: load()}

var Registry = map[string]gotsx.HostModule{"data": {Value: Data, Go: "host.Data"}}
