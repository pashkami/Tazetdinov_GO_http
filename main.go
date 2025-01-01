package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	serverURL       = "http://srv.msk01.gigacorp.local/_stats"
	errorThreshold  = 3
	loadAverageLimit = 30.0
	memoryUsageLimit = 80.0
	diskSpaceLimit   = 0.1 // 10% (inverted for free space calculation)
	networkUsageLimit = 0.1 // 10% (inverted for free bandwidth calculation)
)

func fetchServerStats() ([]float64, error) {
	resp, err := http.Get(serverURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("invalid response status")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	parts := strings.Split(strings.TrimSpace(string(body)), ",")
	if len(parts) != 7 {
		return nil, errors.New("unexpected data format")
	}

	stats := make([]float64, 7)
	for i, part := range parts {
		stats[i], err = strconv.ParseFloat(part, 64)
		if err != nil {
			return nil, errors.New("invalid numeric value")
		}
	}

	return stats, nil
}

func checkAndReport(stats []float64) {
	loadAverage := stats[0]
	if loadAverage > loadAverageLimit {
		fmt.Printf("Load Average is too high: %.0f\n", loadAverage)
	}

	totalMemory := stats[1]
	usedMemory := stats[2]
	memoryUsage := (usedMemory / totalMemory) * 100
	if memoryUsage > memoryUsageLimit {
		fmt.Printf("Memory usage too high: %.0f%%\n", memoryUsage)
	}

	totalDisk := stats[3]
	usedDisk := stats[4]
	freeDisk := totalDisk - usedDisk
	if freeDisk/totalDisk < diskSpaceLimit {
		fmt.Printf("Free disk space is too low: %.0f Mb left\n", freeDisk/1024/1024)
	}

	totalNetwork := stats[5]
	usedNetwork := stats[6]
	freeNetwork := totalNetwork - usedNetwork
	if freeNetwork/totalNetwork < networkUsageLimit {
		fmt.Printf("Network bandwidth usage high: %.0f Mbit/s available\n", freeNetwork*8/1024/1024)
	}
}

func main() {
	errorCount := 0

	for {
		stats, err := fetchServerStats()
		if err != nil {
			errorCount++
			if errorCount >= errorThreshold {
				fmt.Println("Unable to fetch server statistic")
				return
			}
			time.Sleep(5 * time.Second)
			continue
		}

		errorCount = 0
		checkAndReport(stats)
		time.Sleep(10 * time.Second)
	}
}
