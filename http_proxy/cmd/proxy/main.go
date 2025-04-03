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

	fmt.Printf("Запускаем MITM-прокси на порту %s...\n", port)

	// 1) Загружаем CA
	err := proxy.LoadCA("ca.crt", "ca.key")
	if err != nil {
		fmt.Printf("Ошибка загрузки CA: %v\n", err)
		os.Exit(1)
	}

	// 2) Запуск прокси-сервера
	err = proxy.StartProxy(port)
	if err != nil {
		fmt.Printf("Ошибка запуска прокси: %v\n", err)
		os.Exit(1)
	}
}
