# URL Shortener

Простой REST API для сокращения URL на Go.

## Запуск

### 1. Клонировать репозиторий

```bash
git clone https://github.com/yellooww/url-shortener.git
cd url-shortener
```

### 2. Создать `.env`

Скопировать пример конфигурации:

```bash
cp .env.example .env
```

При необходимости изменить значения в `.env`.

### 3. Запустить приложение

```bash
docker compose up 
```

После запуска API доступен по адресу:

```text
http://localhost:8082
```

PostgreSQL запускается автоматически вместе с приложением.

## API

### Создать короткий URL

```http
POST /url
```

Требуется Basic Auth.

Пример тела:

```json
{
  "url": "https://google.com",
  "alias": "google"
}
```

Если `alias` не указан, он генерируется автоматически.

### Перейти по короткому URL

```http
GET /{alias}
```

Например:

```text
http://localhost:8082/google
```

### Обновить URL

```http
PUT /url
```

Пример:

```json
{
  "alias": "google",
  "url": "https://github.com"
}
```

Требуется Basic Auth.

### Удалить URL

```http
DELETE /url/{alias}
```

Требуется Basic Auth.

