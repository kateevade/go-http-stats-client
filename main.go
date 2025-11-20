package main

import (
	"bufio"
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func main() {
	// Счётчик последовательных ошибок
	errorCount := 0
	// URL сервера (в автотестах он будет доступен)
	url := "http://srv.msk01.gigacorp.local/_stats"

	for {
		resp, err := http.Get(url)
		if err != nil || resp.StatusCode != http.StatusOK {
			errorCount++
			if errorCount >= 3 {
				fmt.Println("Unable to fetch server statistic")
				// Сбрасываем счётчик, чтобы сообщение не спамило каждые 2 секунды
				errorCount = 0
			}
			time.Sleep(2 * time.Second)
			continue
		}

		// Сбрасываем счётчик при успешном ответе
		errorCount = 0

		defer resp.Body.Close()

		// Читаем тело ответа
		scanner := bufio.NewScanner(resp.Body)
		var line string
		for scanner.Scan() {
			line = scanner.Text()
		}
		if err := scanner.Err(); err != nil {
			log.Println("Error reading body:", err)
			time.Sleep(2 * time.Second)
			continue
		}

		// Убираем возможные пробелы и разбиваем по запятым
		fields := strings.Split(strings.TrimSpace(line), ",")
		if len(fields) != 7 {
			errorCount++
			if errorCount >= 3 {
				fmt.Println("Unable to fetch server statistic")
				errorCount = 0
			}
			time.Sleep(2 * time.Second)
			continue
		}

		// Парсим числа
		loadAvg, _ := strconv.ParseFloat(fields[0], 64)

		totalRAM, _ := strconv.ParseInt(fields[1], 10, 64)
		usedRAM, _ := strconv.ParseInt(fields[2], 10, 64)

		totalDisk, _ := strconv.ParseInt(fields[3], 10, 64)
		usedDisk, _ := strconv.ParseInt(fields[4], 10, 64)

		totalNet, _ := strconv.ParseInt(fields[5], 10, 64)  // байт/сек
		usedNet, _ := strconv.ParseInt(fields[6], 10, 64)   // байт/сек

		// 1. Load Average > 30
		if loadAvg > 30 {
			fmt.Printf("Load Average is too high: %.0f\n", loadAvg)
		}

		// 2. RAM usage > 80%
		if totalRAM > 0 {
			ramPercent := float64(usedRAM) / float64(totalRAM) * 100
			if ramPercent > 80 {
				fmt.Printf("Memory usage too high: %.0f%%\n", math.Round(ramPercent))
			}
		}

		// 3. Free disk < 10% → выводим сколько МБ осталось свободно
		if totalDisk > 0 {
			freeDiskBytes := totalDisk - usedDisk
			freeDiskMB := float64(freeDiskBytes) / 1024 / 1024
			usedPercent := float64(usedDisk) / float64(totalDisk) * 100
			if usedPercent > 90 {
				fmt.Printf("Free disk space is too low: %.0f Mb left\n", math.Round(freeDiskMB))
			}
		}

		// 4. Network bandwidth usage > 90% → сколько Мбит/с осталось свободно
		if totalNet > 0 {
			freeNetBytesPerSec := totalNet - usedNet
			// 1 байт/сек = 8 бит/сек → переводим в Мбит/сек
			freeNetMbitPerSec := float64(freeNetBytesPerSec*8) / 1000 / 1000
			usedNetPercent := float64(usedNet) / float64(totalNet) * 100
			if usedNetPercent > 90 {
				fmt.Printf("Network bandwidth usage high: %.0f Mbit/s available\n", math.Round(freeNetMbitPerSec))
			}
		}

		// Ждём 2 секунды до следующего запроса
		time.Sleep(2 * time.Second)
	}
}