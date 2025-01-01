package main

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func main() {
	errorCount := 0
	ticker := time.NewTicker(requestInterval)
	defer ticker.Stop()

	for range ticker.C {
		err := fetchAndAnalyzeStats()
		if err != nil {
			fmt.Println("Error:", err)
			errorCount++
			if errorCount >= maxErrorCount {
				fmt.Println("Unable to fetch server statistic.")
				break
			}
		} else {
			errorCount = 0 // Сброс ошибки при успешном запросе
		}
	}
}

const (
	serverURL         = "http://srv.msk01.gigacorp.local/_stats"
	requestInterval   = 10 * time.Second // Интервал запросов
	maxErrorCount     = 3
	loadAverageLimit  = 30
	memoryUsageLimit  = 0.8 // 80%
	diskSpaceLimit    = 0.9 // 90%
	networkUsageLimit = 0.9 // 90%
)

func fetchAndAnalyzeStats() error {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(serverURL)
	if err != nil {
		return fmt.Errorf("failed to fetch server stats: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %v", err)
	}

	stats, err := parseStats(string(body))
	if err != nil {
		return fmt.Errorf("failed to parse stats: %v", err)
	}

	analyzeStats(stats)
	return nil
}

func parseStats(data string) ([]float64, error) {
	parts := strings.Split(strings.TrimSpace(data), ",")
	if len(parts) != 7 {
		return nil, fmt.Errorf("unexpected data format")
	}

	stats := make([]float64, len(parts))
	for i, part := range parts {
		value, err := strconv.ParseFloat(part, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid number: %s", part)
		}
		stats[i] = value
	}

	return stats, nil
}

func analyzeStats(stats []float64) {
	loadAverage := stats[0]
	totalMemory := stats[1]
	usedMemory := stats[2]
	totalDisk := stats[3]
	usedDisk := stats[4]
	totalNetwork := stats[5]
	usedNetwork := stats[6]

	// Проверка Load Average
	if loadAverage > loadAverageLimit {
		fmt.Printf("Load Average is too high: %d\n", int(loadAverage))
	}

	// Проверка использования памяти
	if totalMemory > 0 {
		memoryUsage := usedMemory / totalMemory
		if memoryUsage > memoryUsageLimit {
			fmt.Printf("Memory usage too high: %d%%\n", int(memoryUsage*100))
		}
	}

	// Проверка свободного места на диске
	if totalDisk > 0 {
		freeDiskSpace := totalDisk - usedDisk
		if freeDiskSpace/totalDisk < (1 - diskSpaceLimit) {
			fmt.Printf("Free disk space is too low: %d Mb left\n", int(freeDiskSpace/1024/1024))
		}
	}

	// Проверка использования сети
	if totalNetwork > 0 {
		freeNetwork := totalNetwork - usedNetwork
		if usedNetwork/totalNetwork > networkUsageLimit {
			fmt.Printf("Network bandwidth usage high: %f Mbit/s available\n", (freeNetwork))
		}
	}
}
