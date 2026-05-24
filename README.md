# go-diamond

A distributed configuration management center similar to Apollo/Nacos, built with Go.

## Features

- **Multi-tenant Configuration Management**: Namespace + Group + DataID hierarchical model
- **Configuration Version Control**: Full history tracking and rollback support
- **Real-time Change Detection**: Long-polling and batch watching for configuration changes
- **Language Independent**: HTTP REST API allows integration with any language (Python, Java, Node.js, PHP, etc.)
- **Soft Delete**: Safe configuration removal with recovery capability
- **MD5 Change Detection**: Efficient configuration comparison using content hash

## Tech Stack

- **Backend**: Go + Gin + MySQL
- **Frontend**: React + Vite + TypeScript + Tailwind CSS
- **Metrics**: Prometheus
- **Logging**: Zap

## Quick Start

### Prerequisites

- Go 1.22+
- MySQL 8.0+
- Node.js 18+ (for frontend)

### Backend Setup

#### Option 1: Without Docker

1. **Install MySQL locally** (e.g., via Homebrew on macOS):
   ```bash
   brew install mysql
   brew services start mysql
   ```

2. **Create database and run migrations**:
   ```bash
   mysql -uroot -e "CREATE DATABASE IF NOT EXISTS go_diamond"
   for f in db/migrations/*.sql; do mysql -uroot go_diamond < $f; done
   ```

3. **Modify config for local MySQL** (if needed, update `deploy/config.yaml`):
   - Default DSN for socket: `root:@unix(/var/run/mysqld/mysqld.sock)/go_diamond?charset=utf8mb4&parseTime=True`
   - For TCP connection: `root:@tcp(127.0.0.1:3306)/go_diamond?charset=utf8mb4&parseTime=True`

4. **Build and run server**:
   ```bash
   go build -o bin/go-diamond ./cmd/server
   ./bin/go-diamond -config deploy/config.yaml
   ```

#### Option 2: With Docker

```bash
# Start MySQL with Docker
cd deploy
docker-compose up -d mysql

# Run migrations
mysql -h127.0.0.1 -uroot -ppassword -e "CREATE DATABASE IF NOT EXISTS go_diamond"
for f in db/migrations/*.sql; do mysql -h127.0.0.1 -uroot -ppassword go_diamond < $f; done

# Build and run server
cd ..
go build -o bin/go-diamond ./cmd/server
./bin/go-diamond -config deploy/config.yaml
```

### Frontend Setup

```bash
cd frontend
npm install
npm run dev
```

Visit http://localhost:5173 and login with:
- Email: `admin@example.com`
- Password: `admin123`

## Project Structure

```
.
├── cmd/server/           # Server entry point
├── internal/
│   ├── config/           # Configuration loading
│   ├── handler/          # HTTP handlers
│   ├── middleware/       # Auth & logging middleware
│   ├── model/            # Data models
│   ├── notifier/         # DB polling notifier
│   ├── server/           # Server core
│   ├── service/          # Business logic
│   ├── store/            # Database access
│   └── watcher/          # Watch subscription hub
├── pkg/diamond/          # Go client SDK
├── frontend/             # React frontend
├── examples/             # Multi-language examples
│   ├── python/           # Python client
│   └── java/             # Java client
├── db/migrations/        # SQL migrations
└── deploy/               # Deployment configs
```

## API Reference

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/v1/configs/{ns}/{group}/{dataId}` | GET | Get configuration |
| `/api/v1/watch/{ns}/{group}/{dataId}` | GET | Watch for changes |
| `/api/v1/watch/batch` | POST | Batch watch |
| `/api/v1/configs` | POST | Create config |
| `/api/v1/configs/{ns}/{group}/{dataId}` | PUT | Update config |
| `/api/v1/configs/{ns}/{group}/{dataId}` | DELETE | Delete config |
| `/api/v1/configs/{ns}/{group}/{dataId}/histories` | GET | Get history |
| `/api/v1/configs/{ns}/{group}/{dataId}/rollback` | POST | Rollback |

### Response Format

```json
{
  "code": 0,
  "data": {
    "dataId": "app.json",
    "group": "DEFAULT_GROUP",
    "namespace": "default",
    "content": "{\"key\": \"value\"}",
    "contentMd5": "abc123...",
    "version": 1,
    "format": "json",
    "updatedAt": "2024-01-01T00:00:00Z"
  }
}
```

## Multi-Language Clients

See [examples/README.md](examples/README.md) for Python and Java client examples.

## Testing

```bash
# Backend tests
go test ./... -v -race

# Frontend tests
cd frontend
npm run build
npx playwright test
```

## Load Testing

### Prerequisites

Install [k6](https://grafana.com/docs/k6/latest/set-up/install-k6/):

```bash
# macOS
brew install k6
```

### Run Stress Tests

```bash
# Default targets localhost:8080
k6 run load-test-platform/stress-test/api-stress.js

# Custom API address
BASE_URL=http://your-api:8080 k6 run load-test-platform/stress-test/api-stress.js
```

The API stress test covers health check, config retrieval, and long-polling watch endpoints.

## Configuration

Edit `deploy/config.yaml`:

```yaml
server:
  host: "0.0.0.0"
  port: 8080
  adminToken: "change-me-in-production"  # Change in production

database:
  dsn: "root:password@tcp(127.0.0.1:3306)/go_diamond"
```

## License

MIT