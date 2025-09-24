# Configurable Payment Deadline Usage

## Overview

The payment deadline is now configurable through environment variables, making it easy to adjust the timeout period for different environments or requirements.

## Configuration

### Environment Variable

Set the `PAYMENT_DEADLINE` environment variable to specify the payment deadline in minutes:

```bash
# Set payment deadline to 30 minutes
export PAYMENT_DEADLINE=30

# Set payment deadline to 5 minutes (for testing)
export PAYMENT_DEADLINE=5

# Default is 15 minutes if not set
```

### Docker Compose

Update the `docker-compose.yml` file:

```yaml
services:
  app:
    environment:
      - PAYMENT_DEADLINE=30  # 30 minutes
```

### Docker Run

```bash
docker run -e PAYMENT_DEADLINE=30 your-app-image
```

## Examples

### Development Environment
```bash
# Quick testing with 2-minute deadline
export PAYMENT_DEADLINE=2
go run main.go
```

### Production Environment
```bash
# Longer deadline for production
export PAYMENT_DEADLINE=30
go run main.go
```

### Testing Environment
```bash
# Very short deadline for testing timeout behavior
export PAYMENT_DEADLINE=1
go run main.go
```

## Code Implementation

The configuration is handled in the `config` package:

```go
type Config struct {
    DatabaseURL      string
    RedisURL         string
    PaymentDeadline  int // in minutes
}

func Load() *Config {
    return &Config{
        DatabaseURL:     getEnv("DATABASE_URL", "postgres://..."),
        RedisURL:        getEnv("REDIS_URL", "localhost:6379"),
        PaymentDeadline: getEnvAsInt("PAYMENT_DEADLINE", 15), // Default 15 minutes
    }
}
```

The booking service uses this configuration:

```go
func (s *BookingService) CreateBooking(req *models.CreateBookingRequest) (*models.Booking, error) {
    // ... validation logic ...
    
    // Set payment deadline from config
    paymentDeadline := time.Now().Add(time.Duration(s.paymentDeadline) * time.Minute)
    
    // ... rest of the logic ...
}
```

## Benefits

1. **Environment-specific configuration**: Different deadlines for dev, staging, and production
2. **Easy testing**: Short deadlines for testing timeout scenarios
3. **Runtime flexibility**: Change deadline without code changes
4. **Default fallback**: Always has a sensible default (15 minutes)
5. **Type safety**: Proper integer parsing with error handling

## Validation

The system includes proper validation:
- Invalid values fall back to the default (15 minutes)
- Empty values use the default
- Non-numeric values are handled gracefully
- Type conversion errors are caught and handled
