package main

import (
	"fmt"
	"os"
	"time"

	"github.com/beevik/ntp"
)

func main() {
	servers := []string{"0.ru.pool.ntp.org", "1.ru.pool.ntp.org"}

	for {
		var timeNow time.Time
		var err error

		for _, server := range servers {
			timeNow, err = ntp.Time(server)
			if err == nil {
				break
			}
			fmt.Fprintf(os.Stderr, "Ошибка при получении времени с %s: %v\n", server, err)
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "Не удалось получить время ни с одного сервера\n")
			os.Exit(1)
		}

		fmt.Println("Текущее время (NTP):", timeNow)
		fmt.Println("Текущее время в формате RFC3339:", timeNow.Format(time.RFC3339))

		response, err := ntp.Query(servers[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Ошибка при запросе смещения: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Смещение от системного времени:", response.ClockOffset)

		time.Sleep(5 * time.Second)
	}
}
