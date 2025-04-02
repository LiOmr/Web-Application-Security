package proxy

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/url"
	"strings"
)

func handleClient(clientConn net.Conn) {
	defer clientConn.Close()

	reader := bufio.NewReader(clientConn)
	firstLine, err := reader.ReadString('\n')
	if err != nil {
		return
	}
	firstLine = strings.TrimSpace(firstLine)
	if firstLine == "" {
		return
	}

	parts := strings.SplitN(firstLine, " ", 3)
	if len(parts) != 3 {
		fmt.Println("Неверная стартовая строка:", firstLine)
		return
	}
	method, rawURL, version := parts[0], parts[1], parts[2]

	if strings.ToUpper(method) == "CONNECT" {
		handleConnect(clientConn, reader, rawURL, version)
	} else {
		handleHTTP(clientConn, reader, method, rawURL, version)
	}
}

func handleHTTP(clientConn net.Conn, reader *bufio.Reader, method, rawURL, version string) {

	headers := []string{}
	for {
		line, err := reader.ReadString('\n')
		if err != nil {

			break
		}
		line = strings.TrimRight(line, "\r\n")
		if line == "" {

			break
		}
		headers = append(headers, line)
	}

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		fmt.Println("Ошибка парсинга URL:", rawURL, err)
		return
	}
	host := parsedURL.Hostname()
	port := parsedURL.Port()
	if port == "" {
		port = "80"
	}
	path := parsedURL.RequestURI()
	if path == "" {
		path = "/"
	}

	newFirstLine := fmt.Sprintf("%s %s %s", method, path, version)

	newHeaders := make([]string, 0, len(headers))
	var hasHost bool
	for _, h := range headers {
		hv := strings.SplitN(h, ":", 2)
		if len(hv) == 2 {
			name := strings.TrimSpace(hv[0])
			val := strings.TrimSpace(hv[1])
			lower := strings.ToLower(name)

			if lower == "proxy-connection" {

				continue
			}
			if lower == "host" {

				val = host
				hasHost = true
			}
			newHeaders = append(newHeaders, fmt.Sprintf("%s: %s", name, val))
		}
	}
	if !hasHost {
		newHeaders = append(newHeaders, "Host: "+host)
	}

	body := []byte{}
	buf := make([]byte, 4096)
	for {
		n, err := reader.Read(buf)
		if n > 0 {
			body = append(body, buf[:n]...)
		}
		if err != nil {

			break
		}
	}

	remoteConn, err := net.Dial("tcp", net.JoinHostPort(host, port))
	if err != nil {
		fmt.Println("Ошибка подключения к серверу:", err)
		return
	}
	defer remoteConn.Close()

	reqStr := newFirstLine + "\r\n" + strings.Join(newHeaders, "\r\n") + "\r\n\r\n"
	_, _ = remoteConn.Write([]byte(reqStr))
	if len(body) > 0 {
		_, _ = remoteConn.Write(body)
	}

	go io.Copy(remoteConn, reader)
	io.Copy(clientConn, remoteConn)
}

func handleConnect(clientConn net.Conn, reader *bufio.Reader, rawURL, version string) {

	hostPort := rawURL
	if !strings.Contains(hostPort, ":") {
		hostPort += ":443"
	}

	resp := "HTTP/1.0 200 Connection established\r\n\r\n"
	_, _ = clientConn.Write([]byte(resp))

	host, port, err := net.SplitHostPort(hostPort)
	if err != nil {
		fmt.Println("Ошибка SplitHostPort:", hostPort, err)
		return
	}

	cert, err := getOrGenerateCert(host)
	if err != nil {
		fmt.Println("Ошибка getOrGenerateCert:", err)
		return
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{*cert},
		ServerName:   host,
	}
	tlsClientConn := tls.Server(clientConn, tlsConfig)

	err = tlsClientConn.Handshake()
	if err != nil {
		fmt.Println("Ошибка TLS Handshake (клиент->прокси):", err)
		tlsClientConn.Close()
		return
	}

	realTLS, err := tls.Dial("tcp", net.JoinHostPort(host, port), &tls.Config{
		InsecureSkipVerify: true,
	})
	if err != nil {
		fmt.Println("Ошибка Dial к реальному серверу:", err)
		tlsClientConn.Close()
		return
	}

	// Пересылаем данные: клиент <-> (наш TLS-сервер) <-> реальный TLS-сервер
	go io.Copy(realTLS, tlsClientConn)
	io.Copy(tlsClientConn, realTLS)

	realTLS.Close()
	tlsClientConn.Close()
}
