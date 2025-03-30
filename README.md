# Web Application Security

Написать http прокси сервер и простой сканер уязвимости на его основе (часть функционала burp suite без gui). Язык программирования любой, можно использовать любые библиотеки, кроме тех, которые реализуют функционал http прокси (https://mitmproxy.org и тп). Пример реализации 1 и 2 пунктов (слегка устаревшие) https://github.com/john-pentest/fproxy 
https://github.com/kr/mitm/blob/master/mitm.go 
Обязательно сделать задание в виде докер контейнера (или docker-compose из нескольких контейнеров), в котором прокси слушает на порту 8080, на порту 8000 веб-апи (например, 
/requests – список запросов
/requests/id – вывод 1 запроса
/repeat/id – повторная отправка запроса
/scan/id – сканирование запроса)
1. Проксирование HTTP запросов – 20 баллов
Должны успешно проксироваться HTTP запросы. Команда curl -x http://127.0.0.1:8080 http://mail.ru (8080 – порт, на котором запущена программа) должна возвращать

<html>
<head><title>301 Moved Permanently</title></head>
<body bgcolor="white">
<center><h1>301 Moved Permanently</h1></center>
<hr><center>nginx/1.14.1</center>
</body>
</html>

На вход прокси приходит запрос вида

GET http://mail.ru/ HTTP/1.1
Host: mail.ru
User-Agent: curl/7.64.1
Accept: */*
Proxy-Connection: Keep-Alive

Необходимо:
- считать хост и порт из первой строчки
- заменить путь на относительный
- удалить заголовок Proxy-Connection
Отправить на считанный хост (mail.ru:80) получившийся запрос

GET / HTTP/1.1
Host: mail.ru
User-Agent: curl/7.64.1
Accept: */*

Перенаправить все, что будет получено в ответ

HTTP/1.1 301 Moved Permanently
Server: nginx/1.14.1
Date: Sat, 12 Sep 2020 08:04:13 GMT
Content-Type: text/html
Content-Length: 185
Connection: close
Location: https://mail.ru/

<html>
<head><title>301 Moved Permanently</title></head>
<body bgcolor="white">
<center><h1>301 Moved Permanently</h1></center>
<hr><center>nginx/1.14.1</center>
</body>
</html>

Убедиться, что
- проксируются все типы запросов (GET, POST, HEAD, OPTIONS)
- проксируются все заголовки
- корректно возвращаются все коды ответов (200, 302, 404)
2. Проксирование HTTPS запросов – 20 баллов
Должны успешно проксироваться https запросы. В настройках браузера указать http/https прокси, добавить в ОС корневой сертификат, все сайты должны работать корректно.
Запрос curl -x http://127.0.0.1:8080 https://mail.ru (8080 – порт, на котором запущена программа) должен обрабатываться следующим образом:
- На 8080 порт придет в открытом виде запрос CONNECT https://developer.mozilla.org/en-US/docs/Web/HTTP/Methods/CONNECT 

CONNECT mail.ru:443 HTTP/1.1
Host: mail.ru:443
User-Agent: curl/7.64.1
Proxy-Connection: Keep-Alive

Необходимо считать хост и порт (mail.ru 443) из первой строчки.
Необходимо сразу вернуть ответ (сокет не закрывать, использовать его для последующего зашифрованного соединения)

HTTP/1.0 200 Connection established

После этого curl начнет установку защищенного соединения. Для установки такого соединения необходимо сгенерировать и подписать сертификат для хоста (mail.ru). Команды для генерации корневого сертификата и сертификата хоста https://github.com/john-pentest/fproxy/blob/master/gen_ca.sh https://github.com/john-pentest/fproxy/blob/master/gen_cert.sh  
Необходимо установить защищенное соединение с хостом (mail.ru:443), отправить в него все, что было получено и расшифровано от curl и вернуть ответ.

Убедиться, что получается зайти на сайт mail.ru, авторизоваться и получить список писем
