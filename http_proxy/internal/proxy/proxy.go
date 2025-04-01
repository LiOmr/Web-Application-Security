package proxy

import (
	"fmt"
	"net"
)

// StartHTTPProxy - запускает TCP-сервер, слушает на 0.0.0.0:port,
// каждое новое соединение обрабатывается в отдельной горутине handleClient.
func StartHTTPProxy(port string) error {
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}
	defer listener.Close()

	fmt.Printf("HTTP-прокси слушает на порту %s\n", port)

	for {
		clientConn, err := listener.Accept()
		if err != nil {
			fmt.Println("Ошибка Accept:", err)
			continue
		}

		go handleClient(clientConn)
	}
}
