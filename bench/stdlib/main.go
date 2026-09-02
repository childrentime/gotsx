// stdlib: net/http + html/template rendering the same page as the other contenders.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
)

type Product struct {
	ID, Title, Brand, Desc string
	Price                  int
	Tags                   []string
	Rating                 float64
}

var page = template.Must(template.New("p").Funcs(template.FuncMap{"money": money}).Parse(`<!DOCTYPE html><html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1"><title>Bench · stdlib</title><style>{{.CSS}}</style></head><body><header><strong>bench · stdlib</strong><button id="c">count: <span id="n">0</span></button></header><main><h1>Products</h1><p class="muted">{{len .Items}} products</p><div class="grid">{{range .Items}}<article class="card"><h2>{{.Title}}</h2><p class="muted">{{.Brand}} · {{.Desc}}</p><div class="row">{{range .Tags}}<span class="chip">{{.}}</span>{{end}}<span class="price">{{money .Price}}</span></div></article>{{end}}</div></main><footer class="muted">rendered by html/template</footer><script>var n=0,el=document.getElementById('n');document.getElementById('c').onclick=function(){el.textContent=++n}</script></body></html>`))

func money(c int) string { return fmt.Sprintf("$%d.%02d", c/100, c%100) }

func main() {
	addr := flag.String("addr", ":3001", "")
	flag.Parse()
	b, err := os.ReadFile("../data/products.json")
	if err != nil {
		log.Fatal(err)
	}
	var items []Product
	json.Unmarshal(b, &items)
	css, _ := os.ReadFile("../shared.css")
	data := map[string]any{"Items": items, "CSS": template.CSS(css)}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		page.Execute(w, data)
	})
	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("ok")) })
	log.Fatal(http.ListenAndServe(*addr, nil))
}
