# go-realtime-chat

Real-time chat on Go with WebSocket, RabbitMQ, Redis PubSub, and PostgreSQL.

## Architecture

```
┌──────────┐   WS    ┌──────────────────────────────────────────────────────────┐
│  Client  │────────>│                      API (Gin)                           │
│ (browser │         │                                                          │
│  ,curl,  │         │  ┌──────────────┐   ┌──────────────┐   ┌──────────────┐ │
│  wsclient│         │  │ ChatHandler  │──>│ ChatService  │──>│    Hub       │ │
└──────────┘         │  │ (upgrade WS) │   │              │   │ (broadcast)  │ │
       ▲             │  └──────────────┘   │  PublishTo   │   └──────┬───────┘ │
       │             │                     │  Broker ─────┐         │         │
       │             │                     └──────────────┘ │         │         │
       │             │                                      │         │         │
       │             └──────────────────────────────────────┼─────────┼─────────┘
       │                                                     │         │
       │                                                     ▼         │
       │                                              ┌──────────┐    │
       │                                              │ RabbitMQ │    │
       │                                              │ messages:│    │
       │                                              │ new      │    │
       │                                              └────┬─────┘    │
       │                                                   │          │
       │                                                   ▼          │
       │                                          ┌────────────────┐  │
       │                                          │   Worker       │  │
       │                                          │                │  │
       │                                          │ 1. Unmarshal   │  │
       │                                          │ 2. CreateNew   │  │
       │                                          │    Message()   │  │
       │                                          │ 3. Publish()   │  │
       │                                          │ 4. Ack()       │  │
       │                                          └───┬────┬───────┘  │
       │                                              │    │          │
       │                ┌─────────────────────────────┘    │          │
       │                ▼                                  ▼          │
       │     ┌──────────────────┐              ┌──────────────────┐   │
       │     │   PostgreSQL     │              │  Redis PubSub    │   │
       │     │  messages table  │              │ messages:new     │   │
       │     └──────────────────┘              └────────┬─────────┘   │
       │                                               │             │
       │                  Subscriber (in API) ──────────┘             │
       │                         │                                    │
       └─────────────────────────┴────────────────────────────────────┘
                              Broadcast to all clients in chat room
```

### Flow

1. Client connects via WebSocket to `/api/ws/chat`
2. Client sends `{"user_id": 1, "chat_id": "room-1", "message": "hello"}`
3. `ChatHandler` upgrades WS, `ChatService.HandleIncomingMessage` publishes to RabbitMQ queue
4. `Worker` consumes the message:
   - Unmarshals to `domain.Payload`
   - Inserts into Postgres `messages` table via `MessageRepository.CreateNewMessage`
   - Publishes to Redis PubSub channel
   - Acknowledges the RabbitMQ message
5. API subscribes to Redis PubSub — on receive, broadcasts via Hub to all WS clients in that chat room

## Configuration

All via environment variables. Copy `.env.example` to `.env`:

```bash
cp .env.example .env
```

| Variable | Default | Description |
|---|---|---|
| `API_PORT` | `8000` | HTTP server listen port |
| `AMQP_USER` | `guest` | RabbitMQ username |
| `AMQP_PASSWORD` | `guest` | RabbitMQ password |
| `AMQP_HOST` | `rabbitmq-realtime-chat` | RabbitMQ hostname |
| `AMQP_PORT` | `5672` | RabbitMQ port |
| `REDIS_HOST` | `redis-realtime-chat` | Redis hostname |
| `REDIS_PORT` | `6379` | Redis port |
| `REDIS_PASSWORD` | — | Redis password (empty = no auth) |
| `REDIS_DB` | `0` | Redis database number |
| `REDIS_MAX_RETRIES` | `3` | Redis connection retries |
| `PUBSUB_CHANNEL` | `messages:new` | Redis PubSub channel name |
| `POSTGRES_USER` | `postgres` | PostgreSQL user |
| `POSTGRES_PASSWORD` | `postgres` | PostgreSQL password |
| `POSTGRES_DB` | `chat` | PostgreSQL database name |
| `POSTGRES_HOST` | `postgres-realtime-chat` | PostgreSQL hostname |
| `POSTGRES_PORT` | `5432` | PostgreSQL port |

## Run

### Docker (recommended)

```bash
docker compose up -d
```

### Locally

Start Postgres, Redis, and RabbitMQ manually, then:

```bash
# apply migrations
atlas migrate apply --env local

# start binaries
go run ./cmd/api &
go run ./cmd/worker &
```

### WebSocket test client

```bash
go run ./cmd/wsclient/
```

Type JSON messages in stdin, responses are logged.

## API

| Method | Path | Description |
|---|---|---|
| GET | `/api/health` | Health check |
| GET | `/api/chats/:chat_id/messages` | Paginated message history |

Query params for messages: `limit` (default 50), `offset` (default 0).

### WebSocket

```
Endpoint: ws://localhost:8000/api/ws/chat
```

Send:
```json
{"user_id": 1, "chat_id": "room-1", "message": "hello"}
```

Receive (broadcast):
```json
{"user_id": 1, "chat_id": "room-1", "message": "hello"}
```

### Example

```bash
curl "http://localhost:8000/api/chats/room-1/messages?limit=10"
```

Response:
```json
{
  "chat_id": "room-1",
  "messages": [
    {
      "id": 1,
      "user_id": 1,
      "chat_id": "room-1",
      "text": "hello",
      "created_at": "2026-07-19T20:17:06.586549Z"
    }
  ],
  "limit": 10,
  "offset": 0,
  "total": 1
}
```

## Project structure

```
cmd/
├── api/          — HTTP/WS server entrypoint
├── worker/       — queue consumer entrypoint
└── wsclient/     — WebSocket CLI test client

internal/
├── api/
│   ├── handlers/ — Gin handlers (WS upgrade, messages endpoint)
│   ├── routes.go — route registration
│   └── ws/       — WebSocket client, Hub, upgrader
├── config/       — env-based configuration
├── domain/       — interfaces, entities (Payload, MessageResponse)
├── infra/
│   ├── postgres/ — sqlc gen, migrations (Atlas), pool, MessageRepository
│   ├── rabbitmq/ — Publisher, Consumer
│   └── redis/    — PubSubPublisher, PubSubSubscriber
└── service/
    ├── chat.go   — ChatService (handle, broadcast, history)
    └── worker.go — WorkerService (consume, save, publish)
```

## Tech

| Layer | Technology |
|---|---|
| Language | Go 1.26 |
| HTTP framework | Gin |
| WebSocket | gorilla/websocket |
| Database | PostgreSQL 16 |
| Query generation | sqlc |
| Migrations | Atlas |
| Message queue | RabbitMQ |
| PubSub | Redis 7 |
| Runtime | Docker Compose |
