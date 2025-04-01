package main

import (
	"fmt"
	"os"

	"http_proxy/internal/proxy"
)

func main() {
	port := "8080"
	if len(os.Args) > 1 {
		port = os.Args[1]
	}

	fmt.Printf("Запускаем HTTP-прокси на порту %s...\n", port)
	err := proxy.StartHTTPProxy(port)
	if err != nil {
		fmt.Printf("Ошибка запуска прокси: %v\n", err)
		os.Exit(1)
	}
}
