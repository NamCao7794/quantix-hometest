#!/bin/bash

# Test script for Event Ticket Booking System API
# Make sure the application is running with: docker compose up

BASE_URL="http://localhost:8080/api/v1"

echo "=== Event Ticket Booking System API Test ==="
echo

# Test 1: Create a user
echo "1. Creating a user..."
USER_RESPONSE=$(curl -s -X POST "$BASE_URL/users" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe",
    "email": "john@example.com"
  }')

echo "User created: $USER_RESPONSE"
USER_ID=$(echo $USER_RESPONSE | jq -r '.id')
echo "User ID: $USER_ID"
echo

# Test 2: Create an event
echo "2. Creating an event..."
EVENT_RESPONSE=$(curl -s -X POST "$BASE_URL/events" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Concert 2024",
    "description": "Amazing concert event",
    "date_time": "2024-12-31T20:00:00Z",
    "total_tickets": 1000,
    "ticket_price": 75.50
  }')

echo "Event created: $EVENT_RESPONSE"
EVENT_ID=$(echo $EVENT_RESPONSE | jq -r '.id')
echo "Event ID: $EVENT_ID"
echo

# Test 3: Get all events
echo "3. Getting all events..."
curl -s -X GET "$BASE_URL/events" | jq '.'
echo

# Test 4: Book tickets
echo "4. Booking tickets..."
BOOKING_RESPONSE=$(curl -s -X POST "$BASE_URL/bookings" \
  -H "Content-Type: application/json" \
  -d "{
    \"user_id\": \"$USER_ID\",
    \"event_id\": \"$EVENT_ID\",
    \"quantity\": 2
  }")

echo "Booking created: $BOOKING_RESPONSE"
BOOKING_ID=$(echo $BOOKING_RESPONSE | jq -r '.id')
echo "Booking ID: $BOOKING_ID"
echo

# Test 5: Get booking details
echo "5. Getting booking details..."
curl -s -X GET "$BASE_URL/bookings/$BOOKING_ID" | jq '.'
echo

# Test 6: Get user bookings
echo "6. Getting user bookings..."
curl -s -X GET "$BASE_URL/bookings/user/$USER_ID" | jq '.'
echo

# Test 7: Get event statistics
echo "7. Getting event statistics..."
curl -s -X GET "$BASE_URL/events/$EVENT_ID/statistics" | jq '.'
echo

# Test 8: Test concurrent booking (simulate multiple requests)
echo "8. Testing concurrent booking..."
for i in {1..3}; do
  echo "Creating booking $i..."
  curl -s -X POST "$BASE_URL/bookings" \
    -H "Content-Type: application/json" \
    -d "{
      \"user_id\": \"$USER_ID\",
      \"event_id\": \"$EVENT_ID\",
      \"quantity\": 1
    }" | jq '.id, .status'
done
echo

# Test 9: Get updated event statistics
echo "9. Getting updated event statistics..."
curl -s -X GET "$BASE_URL/events/$EVENT_ID/statistics" | jq '.'
echo

echo "=== Test completed ==="
