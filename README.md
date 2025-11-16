# casino-transaction-system
## Description:
Implementation of an event-driven system managing casino transactions.
It consumes incoming transaction events from Kafka
and exposes a read-only REST API for external clients.

## Architecture
### Key components
- Message Broker - Messaging system that provides incoming events
- Database - Stores transaction entries
- Tx Manager - Listens to incoming events from Kafka and exposes gRPC API for api gateway
- API Gateway - Exposes read-only REST API for external clients

## Technical stack
- Programming language: Golang
- Message Broker: Apache Kafka
- Database: PostgreSQL
- API: REST for external communication, gRPC for internal communication
- Database migrations: Goose
- Containerization: Docker Compose

## API Endpoints
API documentation can be found in `http://{host}:{port}/swagger/index.html#/`

When running locally, host is localhost and port is the port exposed by the API Gateway service.

## Kafka Integration
Tx Manager consumes events from topic: **casino_transactions**.
The events should be sent in this schema:
```json
{
    "userId": "{uuid}",
    "type": "{{bet|win}",
    "amount": "{int}",
    "timestamp": "{{RFC3339 timestamp}}"
}
```

Any event that fails parsing or validation are later sent to topic: **casino_dlq**

## Database
Database consists of 1 table representing business domain - transactions
The fields are:
- id (uuid)
- user_id (uuid)
- transaction_type varchar(10)
- amount int
- transaction_time timestamp with timezone
- t_hash text ( to guarantee that several exact events aren't written several times on a consumer behalf as no transaction id is initially provided from broker)

Its schema is based on migrations that are located in **$(project)/migrations/tx_manager**

## Installation and Setup

### Prerequisites
- Docker and Docker Compose installed

### Setup Steps

1. **Clone the repository**
```bash
git clone https://github.com/e1esm/casino-transaction-system.git
cd casino-transaction-system
```
2. **Start infrastructure and services**
```bash
    make up
```
### General info
Documentation, protobuf files and mocks are already generated so these commands are optional:
```bash
    make docs
    make protoc
    make mockery
```
Note: For protobuf files generation it's required to have Protobuf SDK installed, as all other commands are run in Docker containers

To run tests (Go SDK is required to be preinstalled):
```bash
    make test
    
```

## Configuration

Everything required for the deployment is located in **deployment** folder.

The Docker Compose setup includes the following **services**:

- `zookeeper`
- `kafka` 
- `kafka-ui`
- `postgres`
- `tx_migrations` 
- `init-kafka`
- `tx-manager`
- `api-gateway`

### Volumes

| Volume                     | Purpose |
|-----------------------------|---------|
| `postgres_transactions_data` | Stores PostgreSQL database data persistently |
| `kafka_data`                 | Stores Kafka logs persistently |

### Networks

| Network     | Driver |
|------------|--------|
| `casino`   | bridge |

> All services are connected through the `casino` network.




