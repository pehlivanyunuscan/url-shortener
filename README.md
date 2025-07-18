# URL Shortener Service

A simple, production-ready URL shortening service built with Go, Fiber, PostgreSQL, and Redis.  
This service allows users to shorten long URLs, redirect to original URLs, view statistics, and manage short links with expiration and rate limiting.

---

## Features

- Shorten long URLs to unique, alphanumeric codes (minimum 6 characters)
- Prevent duplicate short codes and original URLs
- Expiration support for short URLs 
- Redis-based caching and rate limiting
- View usage statistics and expiration info for each short URL
- Soft delete support (deleted URLs' stats are still accessible)
- Dockerized for easy deployment

---

## Endpoints

| Method | Route                      | Description                                 |
|--------|----------------------------|---------------------------------------------|
| POST   | `/shorten`                 | Shorten a new URL                           |
| GET    | `/`                        | List all shortened URLs                     |
| GET    | `/:shortCode`              | Redirect to the original URL                |
| GET    | `/:shortCode/stats`        | Show statistics for a short URL             |
| DELETE | `/:shortCode`              | Delete a short URL                          |

---

## Example Create Response

```json
{
  "created_at": "2021-01-01T00:00:00Z",
  "deleted_at": null,
  "original_url": "https://www.google.com",
  "short_url": "http://localhost:3000/abc123",
  "expires_at": "2021-01-02T00:00:00Z",
  "usage_count": 0
}
```

---


## How to Run on Your Localhost

### **With Docker (Recommended)**

1. **Clone the repository:**
    ```sh
    git clone https://github.com/pehlivanyunuscan/url-shortener.git
    cd url-shortener
    ```

2. **Start all services (API, PostgreSQL, Redis) with Docker Compose:**
    ```sh
    docker-compose up --build
    ```

3. **Services**

- API: [http://localhost:3000](http://localhost:3000)
- PostgreSQL: localhost:5432
- Redis: localhost:6379

---

## API Requests

### GET all URLs
```sh
curl -X GET http://localhost:3000/
```

### POST (Shorten a URL)
```sh
curl --header "Content-Type: application/json" --request POST --data '{"original_url": "https://www.google.com"}' http://localhost:3000/shorten
```

### GET (Redirect)
```sh
curl -v http://localhost:3000/<shortCode>
```

### GET (Stats)
```sh
curl -X GET http://localhost:3000/<shortCode>/stats
```

### DELETE (Delete a short URL)
```sh
curl -X DELETE http://localhost:3000/<shortCode>
```

---

## Environment Variables

These environment variables are set in `docker-compose.yml` and used by the application:

| Variable        | Description                       | Example Value         |
|-----------------|-----------------------------------|----------------------|
| DB_HOST         | PostgreSQL host                   | postgres             |
| DB_PORT         | PostgreSQL port                   | 5432                 |
| DB_USER         | PostgreSQL user                   | urluser              |
| DB_PASSWORD     | PostgreSQL password               | 12345                |
| DB_NAME         | PostgreSQL database name          | urlshortener         |
| REDIS_ADDR      | Redis address                     | redis:6379           |

---

## Project Structure

```
.
├── main.go
├── Dockerfile
├── docker-compose.yml
├── go.mod
├── db/
│   ├── postgres.go
│   └── redis.go
├── handlers/
│   └── url_handler.go
├── middleware/
│   └── rate_limit.go
├── models/
│   └── url.go
├── utils/
│   └── generator.go
└── start-docker.sh
```

---

## Notes

- Both short URLs and original URLs are unique in the database.
- Statistics for deleted (soft deleted) URLs can still be accessed via the `/stats` endpoint.
- Rate limiting is applied per IP using Redis.

---
