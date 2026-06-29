# Shop API

Backend API untuk e-commerce shop, dibangun dengan **Go (Golang)**, **Gin Framework**, **MongoDB**, dan **Elasticsearch**.

## 🚀 Tech Stack

- **Go 1.26.1**
- **Gin** - HTTP web framework
- **MongoDB Driver v2** - Database driver
- **Elasticsearch v8** - Search engine
- **JWT** - Authentication
- **AWS S3** - File storage
- **Docker** - Containerization

## 📁 Project Structure

```
.
├── cmd/
│   ├── api/main.go              # Entry point aplikasi API
│   └── sync/es-sync/main.go     # Entry point sync Elasticsearch
├── internal/
│   ├── article/                  # Article module
│   ├── category/                 # Category module
│   ├── config/                   # Database & external service config
│   ├── domain/                   # Domain models
│   ├── master_product/           # Master product module
│   ├── middleware/               # HTTP middleware
│   ├── mocks/                    # Test mocks
│   ├── product/                  # Product module
│   ├── promotion/                # Promotion module
│   ├── response/                 # Response helpers
│   ├── service/                  # Background services
│   ├── testutil/                 # Test utilities
│   ├── user/                     # User module
│   └── util/                     # Utility functions
├── Dockerfile                    # Docker image definition
├── go.mod                        # Go module dependencies
└── .env.example                  # Environment variables template
```

## 🛠️ Local Development Setup

### Prerequisites

- Go 1.26.1 atau lebih baru
- MongoDB (Atlas atau local)
- Elasticsearch (opsional, untuk fitur search)
- AWS S3 credentials (opsional, untuk upload file)

### 1. Clone Repository

```bash
git clone <repository-url>
cd shop-api
```

### 2. Setup Environment Variables

```bash
cp .env.example .env
```

Edit file `.env` dan isi semua variabel yang diperlukan:

```env
GIN_MODE=release
PORT=8080
APP_ENV=development

ALLOWED_ORIGINS=http://localhost:3000
COOKIE_DOMAIN=localhost
JWT_SECRET=your_jwt_secret_here
RATE_LIMIT_PER_MINUTE=100

MONGODB_URI=your_mongodb_uri
MONGODB_NAME=scraper

AWS_BUCKET_NAME=your_bucket_name
AWS_BUCKET_REGION=your_bucket_region
AWS_ACCESS_KEY=your_access_key
AWS_SECRET_KEY=your_secret_key

ELASTICSEARCH_ENABLED=true
ELASTICSEARCH_URL=https://your-elasticsearch-url
ELASTICSEARCH_API_KEY=your_api_key
ELASTICSEARCH_INDEX=products
```

### 3. Install Dependencies

```bash
go mod download
```

### 4. Run Application

```bash
go run ./cmd/api/main.go
```

Server akan berjalan di `http://localhost:8080`.

## 🐳 Docker Setup

Project ini sudah menyediakan [`Dockerfile`](Dockerfile) untuk containerization.

### Build Docker Image

```bash
docker build -t shop-api .
```

### Run Container

#### Opsi 1: Passing Environment Variables Manual

```bash
docker run -d \
  --name shop-api \
  -p 8080:8080 \
  -e GIN_MODE=release \
  -e PORT=8080 \
  -e APP_ENV=development \
  -e MONGODB_URI="your_mongodb_uri" \
  -e MONGODB_NAME=scraper \
  -e JWT_SECRET="your_jwt_secret" \
  -e ALLOWED_ORIGINS=http://localhost:3000 \
  -e ELASTICSEARCH_ENABLED=true \
  -e ELASTICSEARCH_URL="your_elasticsearch_url" \
  -e ELASTICSEARCH_API_KEY="your_elasticsearch_api_key" \
  -e ELASTICSEARCH_INDEX=products \
  -e AWS_BUCKET_NAME="your_bucket_name" \
  -e AWS_BUCKET_REGION="your_bucket_region" \
  -e AWS_ACCESS_KEY="your_access_key" \
  -e AWS_SECRET_KEY="your_secret_key" \
  shop-api
```

#### Opsi 2: Menggunakan File `.env`

```bash
docker run -d \
  --name shop-api \
  -p 8080:8080 \
  --env-file .env \
  shop-api
```

### Verifikasi Container

```bash
# Cek container yang berjalan
docker ps

# Lihat logs
docker logs -f shop-api

# Test endpoint
curl http://localhost:8080
```

### Stop dan Hapus Container

```bash
# Stop container
docker stop shop-api

# Hapus container
docker rm shop-api

# Hapus image (opsional)
docker rmi shop-api
```

## 🔧 Environment Variables

| Variable                | Required | Default                 | Description                                  |
| ----------------------- | -------- | ----------------------- | -------------------------------------------- |
| `GIN_MODE`              | No       | -                       | Mode Gin: `release` atau `debug`             |
| `PORT`                  | No       | `8080`                  | Port server                                  |
| `APP_ENV`               | No       | -                       | Environment: `development` atau `production` |
| `ALLOWED_ORIGINS`       | No       | `http://localhost:3000` | CORS allowed origins                         |
| `COOKIE_DOMAIN`         | No       | `localhost`             | Cookie domain                                |
| `JWT_SECRET`            | **Yes**  | -                       | Secret key untuk JWT                         |
| `RATE_LIMIT_PER_MINUTE` | No       | `100`                   | Rate limit per menit                         |
| `MONGODB_URI`           | **Yes**  | -                       | MongoDB connection string                    |
| `MONGODB_NAME`          | No       | `shop_db`               | Nama database MongoDB                        |
| `MONGODB_TEST_URI`      | No       | -                       | MongoDB URI untuk testing                    |
| `MONGODB_TEST_NAME`     | No       | -                       | Database name untuk testing                  |
| `AWS_BUCKET_NAME`       | No       | -                       | AWS S3 bucket name                           |
| `AWS_BUCKET_REGION`     | No       | -                       | AWS S3 region                                |
| `AWS_ACCESS_KEY`        | No       | -                       | AWS access key                               |
| `AWS_SECRET_KEY`        | No       | -                       | AWS secret key                               |
| `ELASTICSEARCH_ENABLED` | No       | -                       | Aktifkan Elasticsearch: `true`               |
| `ELASTICSEARCH_URL`     | No       | `http://localhost:9200` | Elasticsearch URL                            |
| `ELASTICSEARCH_API_KEY` | No       | -                       | Elasticsearch API key                        |
| `ELASTICSEARCH_INDEX`   | No       | `products`              | Elasticsearch index name                     |

## 🧪 Testing

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

## 📦 Build Binary

```bash
# Build untuk Linux (CGO disabled)
CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o build/shop-api ./cmd/api/main.go
```

## 📝 Notes

- Rate limiter **aktif otomatis** ketika `APP_ENV=production`.
- Elasticsearch **opsional**; jika tidak tersedia, aplikasi akan fallback ke MongoDB untuk product search.
- Untuk production, gunakan secrets management (Docker secrets, Kubernetes secrets, atau cloud provider secrets) alih-alih hardcode credentials di environment variables.
