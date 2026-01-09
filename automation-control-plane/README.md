# Automation Control Plane

Multi-tenant automation control plane with MySQL, Valkey, Centrifugo, and Quickwit integration.

## Features

- Multi-tenant architecture with project isolation
- RBAC with fine-grained permissions
- Project-aware job scheduling
- Agent presence tracking
- Audit logging
- OpenAPI 3.1 API

## Quick Start

### Using Docker Compose

```bash
cd deploy
docker-compose up -d
```

### Using Kubernetes Helm

```bash
cd deploy/helm
helm install automation-control-plane .
```

## Configuration

Set the following environment variables:

- `MYSQL_DSN` - MySQL connection string
- `VALKEY_ADDR` - Valkey/Redis address
- `CENTRIFUGO_URL` - Centrifugo URL
- `QUICKWIT_URL` - Quickwit URL
- `JWT_SECRET` - JWT signing secret

## API

The API is available at `/api` with OpenAPI 3.1 specification.

## License

Proprietary
