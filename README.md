# Quantix - Event Ticket Booking System

A RESTful API for event ticket booking with transaction safety, concurrency handling, and payment processing simulation.

## Features

- **Event Management**: Create, read, update, and delete events
- **User Management**: User registration and management
- **Ticket Booking**: Safe concurrent ticket booking with row locking
- **Payment Processing**: Simulated payment processing with Redis queue
- **Statistics**: Event statistics including revenue and ticket sales
- **Automatic Cancellation**: Expired bookings are automatically cancelled

## Architecture

- **Language**: Go 1.21
- **Database**: PostgreSQL 15
- **Cache/Queue**: Redis 7
- **Web Framework**: Gin
- **Containerization**: Docker & Docker Compose
- **Migration**: golang-migrate

## Project Structure

```
.
├── docker-compose.yml          # Docker services configuration
├── Dockerfile                  # Go application container
├── go.mod                      # Go dependencies
├── init.sql                    # Database initialization
├── main.go                     # Application entry point
├── migrations/                 # Database migrations
│   ├── 000001_initial_schema.up.sql
│   └── 000001_initial_schema.down.sql
└── internal/
    ├── config/                 # Configuration management
    ├── database/               # Database connection and migrations
    ├── handlers/               # HTTP handlers
    ├── models/                 # Data models and DTOs
    ├── repository/             # Data access layer
    └── services/               # Business logic layer
```

## Quick Start

### Prerequisites

- Docker and Docker Compose
- Go 1.21+ (for local development)

### Running with Docker

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd quantix-hometest
   ```

2. **Start the services**
   ```bash
   docker-compose up --build
   ```

3. **The API will be available at**
   ```
   http://localhost:8080
   ```

### Running Locally

1. **Start PostgreSQL and Redis**
   ```bash
   docker-compose up postgres redis
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Run the application**
   ```bash
   go run main.go
   ```

## API Endpoints

### Events

- `GET /api/v1/events` - Get all events
- `GET /api/v1/events/:id` - Get event by ID
- `POST /api/v1/events` - Create new event
- `PUT /api/v1/events/:id` - Update event
- `DELETE /api/v1/events/:id` - Delete event
- `GET /api/v1/events/:id/statistics` - Get event statistics

### Users

- `GET /api/v1/users` - Get all users
- `GET /api/v1/users/:id` - Get user by ID
- `POST /api/v1/users` - Create new user
- `PUT /api/v1/users/:id` - Update user
- `DELETE /api/v1/users/:id` - Delete user

### Bookings

- `POST /api/v1/bookings` - Create new booking
- `GET /api/v1/bookings/:id` - Get booking by ID
- `PUT /api/v1/bookings/:id/cancel` - Cancel booking
- `GET /api/v1/bookings/user/:user_id` - Get user's bookings

## API Examples

### Create an Event

```bash
curl -X POST http://localhost:8080/api/v1/events \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Concert 2024",
    "description": "Amazing concert event",
    "date_time": "2024-12-31T20:00:00Z",
    "total_tickets": 1000,
    "ticket_price": 75.50
  }'
```

### Create a User

```bash
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe",
    "email": "john@example.com"
  }'
```

### Book Tickets

```bash
curl -X POST http://localhost:8080/api/v1/bookings \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user-uuid-here",
    "event_id": "event-uuid-here",
    "quantity": 2
  }'
```

### Get Event Statistics

```bash
curl http://localhost:8080/api/v1/events/event-uuid-here/statistics
```

## Concurrency and Transaction Safety

The system implements several mechanisms to ensure data consistency and prevent race conditions:

### 1. Row Locking
- Uses `SELECT ... FOR UPDATE` to lock event rows during ticket reservation
- Prevents concurrent bookings from overselling tickets

### 2. Database Transactions
- All booking operations are wrapped in database transactions
- Ensures atomicity of ticket reservation and booking creation

### 3. Atomic Updates
- Uses atomic database operations to update booking statuses
- Prevents partial updates in case of failures

### 4. Optimistic Concurrency Control
- Checks available tickets before creating bookings
- Validates booking status before allowing cancellations

## Payment Processing

The system includes a simulated payment processing system:

### Features
- **Async Processing**: Uses Redis queue for background payment processing
- **Automatic Confirmation**: Successful payments automatically confirm bookings
- **Timeout Handling**: Bookings expire after 15 minutes if payment isn't completed
- **Dead Letter Queue**: Failed payments are logged for manual review

### Payment Flow
1. Booking is created with `PENDING` status
2. Payment job is queued in Redis
3. Background processor handles payment simulation
4. Successful payment confirms the booking
5. Expired bookings are automatically cancelled

## Testing

Run the unit tests:

```bash
go test ./internal/services/...
```

The test suite includes:
- Booking creation with various scenarios
- Concurrency handling tests
- Error condition testing
- Payment processing simulation

## Database Schema

### Events Table
- `id` (UUID, Primary Key)
- `name` (VARCHAR)
- `description` (TEXT)
- `date_time` (TIMESTAMP)
- `total_tickets` (INTEGER)
- `ticket_price` (DECIMAL)
- `created_at`, `updated_at` (TIMESTAMP)

### Users Table
- `id` (UUID, Primary Key)
- `name` (VARCHAR)
- `email` (VARCHAR, Unique)
- `created_at`, `updated_at` (TIMESTAMP)

### Bookings Table
- `id` (UUID, Primary Key)
- `user_id` (UUID, Foreign Key)
- `event_id` (UUID, Foreign Key)
- `quantity` (INTEGER)
- `status` (VARCHAR: PENDING, CONFIRMED, CANCELLED)
- `total_amount` (DECIMAL)
- `payment_deadline` (TIMESTAMP)
- `created_at`, `updated_at` (TIMESTAMP)

## Performance Optimizations

### Database Indexes
- Event date/time for efficient querying
- Booking status for payment processing
- User and event foreign keys for joins
- Payment deadline for expired booking cleanup

### Connection Pooling
- Configured PostgreSQL connection pool
- Redis connection reuse
- Efficient resource management

## Monitoring and Logging

The application includes comprehensive logging for:
- API requests and responses
- Database operations
- Payment processing
- Error conditions
- Performance metrics

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `DB_HOST` | localhost | Database host |
| `DB_PORT` | 5432 | Database port |
| `DB_USER` | postgres | Database user |
| `DB_PASSWORD` | password | Database password |
| `DB_NAME` | ticket_booking | Database name |
| `REDIS_URL` | localhost:6379 | Redis connection URL |
| `PORT` | 8080 | Application port |
| `PAYMENT_DEADLINE` | 15 | Payment deadline in minutes |

## Development

### Adding New Features

1. Create models in `internal/models/`
2. Add repository methods in `internal/repository/`
3. Implement business logic in `internal/services/`
4. Create HTTP handlers in `internal/handlers/`
5. Add routes in `main.go`
6. Write tests for new functionality

### Database Migrations

Create new migrations:

```bash
# Create new migration files
touch migrations/000002_add_new_feature.up.sql
touch migrations/000002_add_new_feature.down.sql
```

## Production Considerations

- Use environment-specific configuration
- Implement proper authentication and authorization
- Add rate limiting and request validation
- Set up monitoring and alerting
- Use production-grade database and Redis instances
- Implement proper logging and error handling
- Add health check endpoints
- Consider horizontal scaling for high traffic

## License

This project is for demonstration purposes as part of a technical assessment.
