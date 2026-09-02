// load: a small, dependency-free HTTP load generator so the benchmark is reproducible without wrk/oha.
// Keep-alive connections, C goroutines, D seconds after a warm-up; prints JSON with throughput and latency percentiles.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

func main() {
	url := flag.String("url", "http://localhost:3000/", "URL to hit")
	c := flag.Int("c", 64, "concurrent connections")
	d := flag.Duration("d", 10*time.Second, "measurement duration")
	warm := flag.Duration("warmup", 2*time.Second, "warm-up duration (not measured)")
	flag.Parse()
	tr := &http.Transport{MaxIdleConns: *c * 2, MaxIdleConnsPerHost: *c * 2, DisableCompression: true}
	client := &http.Client{Transport: tr}
	var mu sync.Mutex
	var lat []time.Duration
	var bytes, errs, count int64
	run := func(dur time.Duration, record bool) {
		var wg sync.WaitGroup
		stop := time.Now().Add(dur)
		for i := 0; i < *c; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				local := make([]time.Duration, 0, 4096)
				for time.Now().Before(stop) {
					t0 := time.Now()
					resp, err := client.Get(*url)
					if err != nil {
						atomic.AddInt64(&errs, 1)
						continue
					}
					n, _ := io.Copy(io.Discard, resp.Body)
					resp.Body.Close()
					if resp.StatusCode != 200 {
						atomic.AddInt64(&errs, 1)
						continue
					}
					if record {
						local = append(local, time.Since(t0))
						atomic.AddInt64(&bytes, n)
						atomic.AddInt64(&count, 1)
					}
				}
				if record {
					mu.Lock()
					lat = append(lat, local...)
					mu.Unlock()
				}
			}()
		}
		wg.Wait()
	}
	run(*warm, false)
	t0 := time.Now()
	run(*d, true)
	elapsed := time.Since(t0)
	sort.Slice(lat, func(i, j int) bool { return lat[i] < lat[j] })
	pct := func(p float64) float64 {
		if len(lat) == 0 {
			return 0
		}
		i := int(float64(len(lat)-1) * p)
		return float64(lat[i].Microseconds()) / 1000
	}
	out := map[string]any{
		"url": *url, "concurrency": *c, "seconds": elapsed.Seconds(), "requests": count, "errors": errs,
		"rps": float64(count) / elapsed.Seconds(), "p50_ms": pct(0.50), "p90_ms": pct(0.90), "p99_ms": pct(0.99), "max_ms": pct(1),
		"bytes_per_response": func() int64 {
			if count == 0 {
				return 0
			}
			return bytes / count
		}(),
	}
	b, _ := json.Marshal(out)
	fmt.Println(string(b))
	if errs > 0 {
		fmt.Fprintf(os.Stderr, "warning: %d errors\n", errs)
	}
}
