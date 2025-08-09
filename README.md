# gin-gorm-article-like-example

A sample Go API built with **Gin** and **GORM**, implementing a simple **Article & Like** feature.  
This repository is designed as a learning resource for **gradually refactoring from a Fat Controller** toward a cleaner architecture.

## Features

- **Article CRUD**
  - Create, list (with pagination), and delete articles
- **Like / Unlike**
  - Idempotent API design
  - Lock-free like counter updates using **unique constraints** and affected rows
- **Pagination & Sorting**
  - `limit` / `offset` pagination
  - Sort by `new` or `popular`
- **Transaction Boundaries**
  - All business logic transactions are managed in the UseCase layer

## Tech Stack

- **Go** (latest)
- [Gin](https://github.com/gin-gonic/gin) - HTTP router & framework
- [GORM](https://gorm.io/) - ORM for MySQL
- **MySQL** 8.x (via Docker Compose)
- [golang-migrate](https://github.com/golang-migrate/migrate) - DB migrations

## Getting Started

### Prerequisites

- Go >= 1.22
- Docker & Docker Compose

### Setup

```bash
# Clone the repository
git clone https://github.com/shinagawa-web/gin-gorm-article-like-example
cd gin-gorm-article-like-example

# Start MySQL
docker compose up -d

# Apply migrations
make migrate-up

# Run the API
go run cmd/api/main.go
```

## Environment Variables
Create a .env file:

```.env
MYSQL_DSN=user:password@tcp(localhost:3306)/article_like?parseTime=true
```

## API Endpoints (Initial Version)

| Method | Endpoint                            | Description         |
| ------ | ----------------------------------- | ------------------- |
| POST   | `/articles`                         | Create article      |
| GET    | `/articles?limit=&offset=&sort=new` | List articles (new) |
| DELETE | `/articles/:id`                     | Delete article      |
| POST   | `/articles/:id/like`                | Like an article     |
| DELETE | `/articles/:id/like`                | Unlike an article   |

> Pagination and sorting will be expanded in later steps.


## Roadmap

This repository is structured for a multi-part blog series:

- Part 1: Fat Controller → Minimal Clean Architecture + Article CRUD
- Part 2: Like/Unlike with idempotent counter updates
- Part 3: Pagination optimizations, sorting, metrics, and CI integration

## curl example

### Health Check

```bash
curl localhost:8080/healthz
```

### Get

```bash
curl "localhost:8080/articles?limit=5&offset=0&sort=new"
```

### Post

```bash
curl -X POST localhost:8080/articles \
  -H "Content-Type: application/json" \
  -d '{"authorId":1,"title":"Hello","body":"First post"}'
```

### Update

```bash
curl -X PUT localhost:8080/articles/1 \
  -H "Content-Type: application/json" \
  -d '{"title":"Updated","body":"Updated body"}'
```

### Delete

```bash
curl -X DELETE localhost:8080/articles/1 -i
```

### Like

```bash
curl -X POST "localhost:8080/articles/1/like?userId=1" -i
```

### Unlike

```bash
curl -X DELETE "localhost:8080/articles/1/like?userId=1" -i
```
