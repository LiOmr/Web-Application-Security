# Web Application Security

## Описание
Реализация HTTP/HTTPS-прокси сервера с функционалом простого сканера уязвимостей (аналог части возможностей Burp Suite без GUI). Приложение выполнено в виде Docker-контейнера с API для управления запросами и сканирования.

## Технические требования
- Язык программирования на выбор
- Запрещено использовать библиотеки, реализующие функционал HTTP-прокси (например, mitmproxy)
- Реализация Docker-контейнера (или docker-compose)

## Структура сервиса

| Порт | Назначение                           |
|------|--------------------------------------|
| 8080 | HTTP/HTTPS прокси сервер             |
| 8000 | Web API                             |

### Web API

- `/requests` – список проксированных запросов
- `/requests/{id}` – получение данных конкретного запроса
- `/repeat/{id}` – повторная отправка сохраненного запроса
- `/scan/{id}` – запуск сканирования запроса на уязвимости

## Реализованные функции

### 1. Проксирование HTTP-запросов
**Оценка: 5 баллов**
- Проксирование любых HTTP-запросов (GET, POST, HEAD, OPTIONS)
- Корректное проксирование заголовков и кодов ответа
- Пример команды проверки:
```bash
curl -x http://127.0.0.1:8080 http://mail.ru
```

### 2. Проксирование HTTPS-запросов
**Оценка: 5 баллов**
- Поддержка CONNECT-метода
- Генерация и подписание сертификатов (используйте скрипты: [gen_ca.sh](https://github.com/john-pentest/fproxy/blob/master/gen_ca.sh), [gen_cert.sh](https://github.com/john-pentest/fproxy/blob/master/gen_cert.sh))
- Пример команды проверки:
```bash
curl -x http://127.0.0.1:8080 https://mail.ru
```

### 3. Повторная отправка и сохранение запросов
**Оценка: 5 баллов**
- Сохранение запросов и ответов в БД (SQL или NoSQL)
- Парсинг запросов (метод, путь, заголовки, cookie, параметры GET и POST)
- Парсинг ответов (код, сообщение, заголовки, тело)

**Пример структуры сохранённого запроса:**
```json
{
  "method": "POST",
  "path": "/path1/path2",
  "get_params": {"x": 123, "y": "qwe"},
  "headers": {"Host": "example.org", "Header": "value"},
  "cookies": {"cookie1": 1, "cookie2": "qwe"},
  "post_params": {"z": "zxc"}
}
```

**Пример структуры сохранённого ответа:**
```json
{
  "code": 200,
  "message": "OK",
  "headers": {"Server": "nginx/1.14.1", "Header": "value"},
  "body": "<html>..."
}
```

### 4. Сканер уязвимостей
**Оценка: 5 баллов**
Реализовать один из вариантов проверки (указан преподавателем):

1. **Command Injection**
   - Проверка параметров через внедрение команд
2. **SQL Injection**
   - Проверка изменения кода и длины ответа при добавлении кавычек
3. **XXE (XML External Entity)**
   - Вставка внешней сущности в XML-запрос
4. **XSS (Cross-Site Scripting)**
   - Проверка внедрения скрипта
5. **Dirbuster**
   - Перебор директорий из словаря [dicc.txt](https://github.com/maurosoria/dirsearch/blob/master/db/dicc.txt)
6. **Param-miner**
   - Поиск скрытых GET-параметров из [словаря](https://github.com/PortSwigger/param-miner/blob/master/resources/params)

## Инструкция по запуску
```bash
docker-compose up --build
```

## Проверка работы
Использовать curl или браузер с настроенным прокси для отправки запросов:
- HTTP: `curl -x http://127.0.0.1:8080 http://example.com`
- HTTPS: `curl -x http://127.0.0.1:8080 https://example.com`



