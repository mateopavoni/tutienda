// Command loadtest hammers the checkout path with many concurrent buyers competing for a scarce
// SKU, then verifies the engine never oversold. It targets the orders and inventory services
// directly so the result reflects the concurrency guarantee, not the gateway's rate limiter.
//
// Usage (inside the compose network):
//
//	docker compose run --rm loadtest
//
// Flags: -sku, -n (total buyers), -c (max in flight).
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/mateopavoni/archive-commerce/internal/platform/tenant"
)

func main() {
	ordersURL := env("ORDERS_URL", "http://localhost:8083")
	inventoryURL := env("INVENTORY_URL", "http://localhost:8082")
	sku := flag.String("sku", "AX1-85", "SKU every buyer competes for")
	total := flag.Int("n", 500, "number of concurrent buyers")
	concurrency := flag.Int("c", 150, "max requests in flight")
	flag.Parse()

	client := &http.Client{Timeout: 10 * time.Second}

	// Reset stock to seed defaults so the demo is repeatable run after run.
	reseed(client, inventoryURL)

	before, err := readStock(client, inventoryURL, *sku)
	if err != nil {
		fmt.Printf("could not read initial stock for %s: %v\n", *sku, err)
		os.Exit(1)
	}

	fmt.Printf("Target SKU %s — starting available: %d\n", *sku, before.Available)
	fmt.Printf("Firing %d buyers (max %d in flight)...\n\n", *total, *concurrency)

	var confirmed, outOfStock, failed int64
	sem := make(chan struct{}, *concurrency)
	var wg sync.WaitGroup
	start := time.Now()

	for i := 0; i < *total; i++ {
		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer wg.Done()
			defer func() { <-sem }()
			switch buy(client, ordersURL, *sku) {
			case http.StatusCreated:
				atomic.AddInt64(&confirmed, 1)
			case http.StatusConflict:
				atomic.AddInt64(&outOfStock, 1)
			default:
				atomic.AddInt64(&failed, 1)
			}
		}()
	}
	wg.Wait()
	elapsed := time.Since(start)

	after, _ := readStock(client, inventoryURL, *sku)

	fmt.Println("Results")
	fmt.Println("-------")
	fmt.Printf("  confirmed orders : %d\n", confirmed)
	fmt.Printf("  out of stock     : %d\n", outOfStock)
	fmt.Printf("  other failures   : %d\n", failed)
	fmt.Printf("  elapsed          : %s\n", elapsed.Round(time.Millisecond))
	fmt.Printf("  stock available  : %d -> %d\n\n", before.Available, after.Available)

	oversold := confirmed > int64(before.Available) || after.Available < 0
	if oversold {
		fmt.Printf("OVERSOLD: %d confirmed against %d available\n", confirmed, before.Available)
		os.Exit(1)
	}
	fmt.Printf("OK: sold exactly %d of %d units, zero oversell.\n", confirmed, before.Available)
}

// reseed asks inventory to reset stock to its seed defaults (best-effort) so the SKU under test
// starts full on every run, even if a previous run or the persisted volume left it depleted.
func reseed(client *http.Client, inventoryURL string) {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, inventoryURL+"/seed", nil)
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err == nil {
		_ = resp.Body.Close()
	}
}

type stock struct {
	SKU       string `json:"sku"`
	Available int    `json:"available"`
	Reserved  int    `json:"reserved"`
}

func readStock(client *http.Client, inventoryURL, sku string) (stock, error) {
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, inventoryURL+"/stock?skus="+sku, nil)
	req.Header.Set(tenant.Header, tenant.DemoID) // demo store: services are tenant-scoped
	resp, err := client.Do(req)
	if err != nil {
		return stock{}, err
	}
	defer func() { _ = resp.Body.Close() }()

	var out struct {
		Items []stock `json:"items"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return stock{}, err
	}
	if len(out.Items) == 0 {
		return stock{}, fmt.Errorf("no stock document for %s", sku)
	}
	return out.Items[0], nil
}

// buy posts a one-unit checkout and returns the HTTP status.
func buy(client *http.Client, ordersURL, sku string) int {
	body, _ := json.Marshal(map[string]any{
		"items": []map[string]any{{"sku": sku, "qty": 1}},
	})
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, ordersURL+"/orders", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(tenant.Header, tenant.DemoID) // demo store: services are tenant-scoped
	resp, err := client.Do(req)
	if err != nil {
		return 0
	}
	defer func() { _ = resp.Body.Close() }()
	return resp.StatusCode
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
