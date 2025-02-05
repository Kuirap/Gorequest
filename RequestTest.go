package main

import (
	"fmt"
	"sync"
	"time"
	"github.com/valyala/fasthttp"
)

const (
	concurrency = 28
	totalReqs   = 1000000
	url         = "https://s60822.cdn.ngenix.net/"
)

func makeRequest(client *fasthttp.Client, wg *sync.WaitGroup, id int) {
	defer wg.Done()

	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)

	req.SetRequestURI(url)
	req.Header.Set("User-Agent", "CustomUserAgent/1.0") // Устанавливаем кастомный User-Agent

	err := client.Do(req, resp)
	if err != nil {
		fmt.Printf("Request %d failed: %v\n", id, err)
		return
	}

	fmt.Printf("Request %d completed with status: %d\n", id, resp.StatusCode())
}

func main() {
	client := &fasthttp.Client{
		MaxConnsPerHost: 10000, // Увеличиваем лимит соединений
	}

	wg := &sync.WaitGroup{}
	semaphore := make(chan struct{}, concurrency)

	start := time.Now()

	for i := 0; i < totalReqs; i++ {
		wg.Add(1)
		semaphore <- struct{}{}

		go func(id int) {
			defer func() { <-semaphore }()
			makeRequest(client, wg, id)
		}(i)
	}

	wg.Wait()
	elapsed := time.Since(start)

	fmt.Printf("All requests completed in %v\n", elapsed)
	fmt.Printf("Achieved RPS: %.2f\n", float64(totalReqs)/elapsed.Seconds())
}
