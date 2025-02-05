package main

import (
	"flag"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/valyala/fasthttp"
)

func makeRequest(client *fasthttp.Client, wg *sync.WaitGroup, id int, url string) {
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
	// Определяем флаги командной строки
	concurrency := flag.Int("concurrency", 0, "Количество параллельных запросов")
	totalReqs := flag.Int("totalReqs", 0, "Общее количество запросов")
	url := flag.String("url", "", "URL для отправки запросов")

	// Парсим флаги
	flag.Parse()

	// Проверяем, что параметры заданы
	if *concurrency == 0 || *totalReqs == 0 || *url == "" {
		fmt.Println("Ошибка: необходимо задать все параметры: -concurrency, -totalReqs, -url")
		os.Exit(1)
	}

	client := &fasthttp.Client{
		MaxConnsPerHost: 10000, // Увеличиваем лимит соединений
	}

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, *concurrency)

	start := time.Now()

	for i := 0; i < *totalReqs; i++ {
		wg.Add(1)
		semaphore <- struct{}{}

		go func(id int) {
			defer func() { <-semaphore }()
			makeRequest(client, &wg, id, *url)
		}(i)
	}

	wg.Wait()
	elapsed := time.Since(start)

	fmt.Printf("Все запросы завершены за %v\n", elapsed)
	fmt.Printf("Достигнуто RPS: %.2f\n", float64(*totalReqs)/elapsed.Seconds())
}

