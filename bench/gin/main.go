// gin: the same html/template page served through Gin (release mode, no logger).
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

type Product struct {
	ID, Title, Brand, Desc string
	Price                  int
	Tags                   []string
	Rating                 float64
}

const tmpl = `<!DOCTYPE html><html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1"><title>Bench · gin</title><style>{{.CSS}}</style></head><body><header><strong>bench · gin</strong><button id="c">count: <span id="n">0</span></button></header><main><h1>Products</h1><p class="muted">{{len .Items}} products</p><div class="grid">{{range .Items}}<article class="card"><h2>{{.Title}}</h2><p class="muted">{{.Brand}} · {{.Desc}}</p><div class="row">{{range .Tags}}<span class="chip">{{.}}</span>{{end}}<span class="price">{{money .Price}}</span></div></article>{{end}}</div></main><footer class="muted">rendered by gin + html/template</footer><script>var n=0,el=document.getElementById('n');document.getElementById('c').onclick=function(){el.textContent=++n}</script></body></html>`

func money(c int) string { return fmt.Sprintf("$%d.%02d", c/100, c%100) }

func main() {
	addr := flag.String("addr", ":3002", "")
	flag.Parse()
	b, err := os.ReadFile("../data/products.json")
	if err != nil {
		log.Fatal(err)
	}
	var items []Product
	json.Unmarshal(b, &items)
	css, _ := os.ReadFile("../shared.css")
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.SetHTMLTemplate(template.Must(template.New("p").Funcs(template.FuncMap{"money": money}).Parse(tmpl)))
	data := gin.H{"Items": items, "CSS": template.CSS(css)}
	r.GET("/", func(c *gin.Context) { c.HTML(http.StatusOK, "p", data) })
	r.GET("/healthz", func(c *gin.Context) { c.String(200, "ok") })
	log.Fatal(r.Run(*addr))
}
