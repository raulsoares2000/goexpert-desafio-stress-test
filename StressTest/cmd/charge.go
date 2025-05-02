package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

// chargeCmd represents the charge command
var chargeCmd = &cobra.Command{
	Use:   "charge",
	Short: "Realiza um stress test no url especificado",
	Long:  "Realiza um stress test no url especificado, usando um número x de requests e x de workers",
	Run: func(cmd *cobra.Command, args []string) {
		url, err := cmd.Flags().GetString("url")
		if err != nil {
			panic(err)
		}
		requests, err := cmd.Flags().GetInt("requests")
		if err != nil {
			panic(err)
		}
		concurrency, err := cmd.Flags().GetInt("concurrency")
		if err != nil {
			panic(err)
		}
		loadTest(url, requests, concurrency)
	},
}

func init() {
	rootCmd.AddCommand(chargeCmd)
	chargeCmd.Flags().StringP("url", "u", "", "Charge Name")
	chargeCmd.Flags().IntP("requests", "r", 0, "Charge Name")
	chargeCmd.Flags().IntP("concurrency", "c", 0, "Charge Name")
	chargeCmd.MarkFlagRequired("url")
	chargeCmd.MarkFlagRequired("requests")
	chargeCmd.MarkFlagRequired("concurrency")
}

func loadTest(url string, requests, concurrency int) {
	var wg sync.WaitGroup
	var successCount atomic.Int64
	var totalDuration time.Duration
	var statusCodesMutex sync.Mutex
	var connectionErrorCount atomic.Int64
	statusCodes := make(map[int]int)
	requestTimes := make(chan time.Duration, requests)

	startTime := time.Now()

	// Controla o número total de requisições
	requestCounter := make(chan struct{}, requests)

	// Inicia as goroutines (workers)
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			client := http.Client{}
			for range requestCounter {
				start := time.Now()
				resp, err := client.Get(url)
				duration := time.Since(start)
				statusCodesMutex.Lock()
				requestTimes <- duration
				statusCodesMutex.Unlock()
				if err == nil {
					if resp.StatusCode == http.StatusOK {
						successCount.Add(1)
					}
					statusCodesMutex.Lock()
					statusCodes[resp.StatusCode]++
					statusCodesMutex.Unlock()
					resp.Body.Close()
				} else {
					connectionErrorCount.Add(1)
				}
			}
		}()
	}

	// Envia os sinais para as goroutines realizarem as requisições
	for i := 0; i < requests; i++ {
		requestCounter <- struct{}{}
	}
	close(requestCounter) // Sinaliza para as goroutines que não há mais requisições a serem feitas

	wg.Wait()
	close(requestTimes)

	endTime := time.Now()
	totalDuration = endTime.Sub(startTime)

	var totalResponseTime time.Duration
	var requestCount int
	for rt := range requestTimes {
		totalResponseTime += rt
		requestCount++
	}
	var averageResponseTime time.Duration
	if requestCount > 0 {
		averageResponseTime = totalResponseTime / time.Duration(requestCount)
	}

	counter := int(successCount.Load())
	errors := int(connectionErrorCount.Load())

	fmt.Printf("Relatório de Teste de Carga:\n")
	fmt.Printf("Tempo total de execução: %s\n", totalDuration)
	fmt.Printf("Número total de requests: %d\n", requests)
	fmt.Printf("Requests com status HTTP 200: %d\n", counter)
	fmt.Printf("Distribuição de outros códigos de status HTTP:\n")
	for code, count := range statusCodes {
		if code != http.StatusOK {
			fmt.Printf("  HTTP %d: %d\n", code, count)
		}
	}

	fmt.Printf("Dados adicionais:\n")
	fmt.Printf("URL: %s\n", url)
	fmt.Printf("Chamadas simultâneas: %d\n", concurrency)
	fmt.Printf("Tempo médio de resposta: %s\n", averageResponseTime)
	fmt.Printf("Erros: %d\n", errors)
}
