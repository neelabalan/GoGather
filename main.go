package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
)

const kb_gb = 1024 * 1024 * 1024

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: go run main.go <interval_in_seconds>")
	}

	interval, err := strconv.Atoi(os.Args[1])
	if err != nil {
		log.Fatalf("Invalid interval: %v", err)
	}

	db, err := sql.Open("sqlite3", "./stats.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	createTableSQL := `
    CREATE TABLE IF NOT EXISTS stats (
        timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
        cpu_usage REAL,
        memory_used INTEGER,
        memory_total INTEGER
    );`
	_, err = db.Exec(createTableSQL)
	if err != nil {
		log.Fatalf("Failed to create table: %v", err)
	}

	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			recordStats(db)
		}
	}
}

func recordStats(db *sql.DB) {
	cpuPercentages, err := cpu.Percent(0, true)
	if err != nil {
		log.Printf("Failed to get CPU usage: %v", err)
		return
	}

	memoryStats, err := mem.VirtualMemory()
	memoryUsedGB := float64(memoryStats.Used) / kb_gb
	memoryTotalGB := float64(memoryStats.Total) / kb_gb

	if err != nil {
		log.Printf("Failed to get memory stats: %v", err)
		return
	}

	avgCPU := 0.0
	for _, cpuPercentage := range cpuPercentages {
		avgCPU += cpuPercentage
	}
	avgCPU /= float64(len(cpuPercentages))

	insertSQL := `INSERT INTO stats (cpu_usage, memory_used, memory_total) VALUES (?, ?, ?)`
	_, err = db.Exec(insertSQL, avgCPU, memoryUsedGB, memoryTotalGB)
	if err != nil {
		log.Printf("Failed to insert stats: %v", err)
		return
	}

	fmt.Printf("Recorded stats: CPU usage: %.2f%%, Memory used: %.2f GB , Memory total: %.2f GB\n", avgCPU, memoryUsedGB, memoryTotalGB)
}
