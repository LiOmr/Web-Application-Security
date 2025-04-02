package proxy

import (
	"fmt"
	"net"
)

// StartProxy - запускает прокси на 0.0.0.0:port
func StartProxy(port string) error {
	ln, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}
	defer ln.Close()

	fmt.Printf("MITM-прокси слушает на порту %s\n", port)

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Ошибка Accept:", err)
			continue
		}
		go handleClient(conn) // см. handler.go
	}
}
