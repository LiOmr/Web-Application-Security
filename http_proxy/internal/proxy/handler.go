package proxy

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// handleClient обрабатывает одно соединение от клиента (например, от curl).
func handleClient(clientConn net.Conn) {
	defer clientConn.Close()

	reader := bufio.NewReader(clientConn)

	// 1) Считываем первую строку запроса
	firstLine, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Ошибка чтения первой строки запроса:", err)
		return
	}
	firstLine = strings.TrimSpace(firstLine)

	// Разбиваем по пробелам - метод, URL, версия
	parts := strings.SplitN(firstLine, " ", 3)
	if len(parts) != 3 {
		fmt.Println("Неверная стартовая строка:", firstLine)
		return
	}
	method, rawURL, version := parts[0], parts[1], parts[2]

	// 2) Считываем заголовки (до пустой строки "\r\n")
	var headers []string
	var contentLength = -1
	var isChunked bool

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Ошибка чтения заголовков:", err)
			return
		}
		line = strings.TrimRight(line, "\r\n")

		if line == "" {
			// Пустая строка — конец секции заголовков
			break
		}

		headers = append(headers, line)
	}

	// 3) Парсим URL для извлечения хоста, порта, пути
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		fmt.Println("Ошибка парсинга URL:", err)
		return
	}
	host := parsedURL.Hostname()
	port := parsedURL.Port()
	if port == "" {
		port = "80"
	}
	path := parsedURL.RequestURI()

	newFirstLine := fmt.Sprintf("%s %s %s", method, path, version)

	// 4) Удаляем Proxy-Connection, правим Host, ищем Content-Length, Transfer-Encoding
	newHeaders := make([]string, 0, len(headers))
	var hasHost bool

	for _, h := range headers {

		hv := strings.SplitN(h, ":", 2)
		if len(hv) != 2 {
			continue
		}
		name := strings.TrimSpace(hv[0])
		val := strings.TrimSpace(hv[1])
		lowerName := strings.ToLower(name)

		if lowerName == "proxy-connection" {

			continue
		}

		if lowerName == "host" {
			val = host
			hasHost = true
		}

		if lowerName == "content-length" {
			cl, err := strconv.Atoi(val)
			if err == nil {
				contentLength = cl
			}
		}

		// Если это Transfer-Encoding, проверим chunked
		if lowerName == "transfer-encoding" && strings.Contains(strings.ToLower(val), "chunked") {
			isChunked = true
		}

		newHeaders = append(newHeaders, fmt.Sprintf("%s: %s", name, val))
	}

	if !hasHost {
		newHeaders = append(newHeaders, "Host: "+host)
	}

	// 5) Считываем тело запроса (POST/PUT и т.д.)

	var bodyData []byte

	if method == "GET" || method == "HEAD" || method == "OPTIONS" {
		bodyData = nil
	} else {
		if contentLength > 0 {
			// Считаем ровно contentLength байт
			bodyData = make([]byte, contentLength)
			if _, err := io.ReadFull(reader, bodyData); err != nil {
				fmt.Println("Ошибка чтения тела запроса по Content-Length:", err)
			}
		} else if isChunked {
			fmt.Println("Запрос с chunked encoding")
		} else {
			// Нет Content-Length, нет chunked => читаем до EOF с небольшим таймаутом
			_ = clientConn.SetReadDeadline(time.Now().Add(1 * time.Second))
			buf := make([]byte, 4096)
			for {
				n, err := reader.Read(buf)
				if n > 0 {
					bodyData = append(bodyData, buf[:n]...)
				}
				if err != nil {
					break
				}
			}
			_ = clientConn.SetReadDeadline(time.Time{})
		}
	}

	// 6) Устанавливаем соединение с целевым сервером
	remoteConn, err := net.Dial("tcp", net.JoinHostPort(host, port))
	if err != nil {
		fmt.Println("Ошибка подключения к удалённому серверу:", err)
		return
	}
	defer remoteConn.Close()

	// 7) Формируем новый запрос: первая строка + заголовки + пустая строка + тело
	finalRequest := newFirstLine + "\r\n" + strings.Join(newHeaders, "\r\n") + "\r\n\r\n"
	_, err = remoteConn.Write([]byte(finalRequest))
	if err != nil {
		fmt.Println("Ошибка записи заголовков на сервер:", err)
		return
	}

	if len(bodyData) > 0 {
		_, err = remoteConn.Write(bodyData)
		if err != nil {
			fmt.Println("Ошибка записи тела на сервер:", err)
			return
		}
	}

	// 8) Читаем ответ от сервера и пересылаем обратно клиенту
	respBuf := make([]byte, 4096)
	for {
		n, err := remoteConn.Read(respBuf)
		if n > 0 {
			_, err2 := clientConn.Write(respBuf[:n])
			if err2 != nil {
				fmt.Println("Ошибка при отправке ответа клиенту:", err2)
				return
			}
		}
		if err != nil {
			if err != io.EOF {
				fmt.Println("Ошибка при чтении ответа от сервера:", err)
			}
			break
		}
	}
}
