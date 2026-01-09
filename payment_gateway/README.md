# Payment Gateway Service

## Overview

The Payment Gateway Service is a microservice responsible for processing payments through multiple channels (UPI, IMPS, NEFT), managing account balances, and handling asynchronous payment processing with webhook notifications. It provides APIs for payment processing, account management, payment channel configuration, and transaction status queries.

## Features

- **Multi-Channel Payment Processing**: Supports UPI, IMPS, and NEFT payment channels
- **Account Management**: Manage payment accounts with balance and threshold tracking
- **Payment Channel Configuration**: Configure limits, success rates, and fees per channel
- **Asynchronous Processing**: Background workers process payments asynchronously
- **Webhook Notifications**: Sends payment status updates to configured notification URLs
- **Idempotency**: Reference ID-based idempotency prevents duplicate payments
- **Transaction Tracking**: Complete audit trail of all payment attempts
- **Channel Availability**: Check channel availability based on time schedules

## Architecture

The service follows a layered architecture:

- **API Layer** (`api/`): HTTP handlers and routing
- **Service Layer** (`api/service/`): Business logic for payments, accounts, and channels
- **Data Access Layer** (`db/daos/`): Database operations
- **Models** (`models/`): Domain models and DTOs
- **Payment Providers** (`payment/`): Channel-specific payment implementations (UPI, IMPS, NEFT)
- **Worker** (`worker/`): Background processing for payments and notifications
- **HTTP Client** (`http/`): HTTP client for external notifications

## Prerequisites

- Go 1.24.6 or higher
- PostgreSQL database

## Environment Variables

The service requires the following environment variables:

- `DATABASE_URL`: PostgreSQL connection string (required)
  - Example: `postgres://user:password@localhost:5432/payment_gateway_db?sslmode=disable`

## Installation

1. Navigate to the payment_gateway directory:
```bash
cd payment_gateway
```

2. Install dependencies:
```bash
go mod download
```

3. Set up the database:
   - Ensure PostgreSQL is running
   - Create a database for the service
   - The service will auto-migrate tables on startup

4. Configure environment variables:
```bash
export DATABASE_URL="postgres://user:password@localhost:5432/payment_gateway_db?sslmode=disable"
```

5. Run the service:
```bash
go run main.go
```

The service will start on port `8080` by default.

## API Endpoints

All endpoints are prefixed with `/api/v1`

### Account Management

#### Create Account
- **Method**: `POST`
- **Path**: `/api/v1/account`
- **Request Body**:
```json
{
  "name": "Main Account",
  "balance": 1000000.0,
  "threshold": 100000.0
}
```
- **Response** (200):
```json
{
  "id": "ACCxxxxxxxxxxxx",
  "name": "Main Account",
  "balance": 1000000.0,
  "threshold": 100000.0
}
```
- **Note**: Only one account can exist at a time

#### List Accounts
- **Method**: `GET`
- **Path**: `/api/v1/account`
- **Response** (200):
```json
[
  {
    "id": "ACCxxxxxxxxxxxx",
    "name": "Main Account",
    "balance": 1000000.0,
    "threshold": 100000.0
  }
]
```

#### Update Account
- **Method**: `PUT`
- **Path**: `/api/v1/account/{id}`
- **Request Body**:
```json
{
  "balance": 500000.0,
  "threshold": 200000.0
}
```
- **Response** (200): Updated account object
- **Note**: Balance is added to existing balance (can be negative to deduct)

### Payment Channel Management

#### Create Payment Channel
- **Method**: `POST`
- **Path**: `/api/v1/channel`
- **Request Body**:
```json
{
  "channel": "UPI",
  "limit": 100000.0,
  "success_rate": 0.95,
  "fee": 0.0
}
```
- **Response** (200):
```json
{
  "id": "CHNxxxxxxxxxxxx",
  "channel": "UPI",
  "limit": 100000.0,
  "success_rate": 0.95,
  "fee": 0.0
}
```
- **Supported Channels**: `UPI`, `IMPS`, `NEFT`

#### List Payment Channels
- **Method**: `GET`
- **Path**: `/api/v1/channel`
- **Response** (200):
```json
[
  {
    "id": "CHNxxxxxxxxxxxx",
    "channel": "UPI",
    "limit": 100000.0,
    "success_rate": 0.95,
    "fee": 0.0
  },
  {
    "id": "CHNyyyyyyyyyyyy",
    "channel": "IMPS",
    "limit": 500000.0,
    "success_rate": 0.90,
    "fee": 5.0
  }
]
```

#### Update Payment Channel
- **Method**: `PUT`
- **Path**: `/api/v1/channel/{channel}`
- **Request Body**:
```json
{
  "limit": 200000.0,
  "success_rate": 0.98,
  "fee": 2.0
}
```
- **Response** (200): Updated payment channel object

#### Check Channel Availability
- **Method**: `GET`
- **Path**: `/api/v1/channel/{channel}/status`
- **Response** (200):
```json
{
  "available": true
}
```
- **Note**: Checks if channel is available based on time schedules

### Payment Processing

#### Process Payment
- **Method**: `POST`
- **Path**: `/api/v1/payment`
- **Request Body**:
```json
{
  "reference_id": "REF-123456789012",
  "amount": 50000.0,
  "channel": "UPI",
  "beneficiary": {
    "name": "John Doe",
    "account": "1234567890",
    "ifsc": "IFSC0001234",
    "bank": "Example Bank"
  },
  "metadata": {
    "loan_id": "LOAN-123",
    "disbursement_id": "DISB-456",
    "notification_url": "http://disbursement-service:7070/api/v1/payment/notification"
  }
}
```
- **Response** (200):
```json
{
  "id": "TXN-123456789012",
  "reference_id": "REF-123456789012",
  "amount": 50000.0,
  "channel": "UPI",
  "fee": 0.0,
  "beneficiary": {
    "name": "John Doe",
    "account": "1234567890",
    "ifsc": "IFSC0001234",
    "bank": "Example Bank"
  },
  "metadata": {
    "loan_id": "LOAN-123",
    "disbursement_id": "DISB-456",
    "notification_url": "http://disbursement-service:7070/api/v1/payment/notification"
  },
  "status": "initiated",
  "message": null,
  "created_at": "2025-01-01T12:00:00Z",
  "updated_at": "2025-01-01T12:00:00Z",
  "processed_at": null,
  "notified_at": null
}
```
- **Note**: Payment is processed asynchronously. Status will be updated via notification.

#### Get Transaction
- **Method**: `GET`
- **Path**: `/api/v1/payment/{channel}/txn/{id}`
- **Response** (200):
```json
{
  "id": "TXN-123456789012",
  "reference_id": "REF-123456789012",
  "amount": 50000.0,
  "channel": "UPI",
  "fee": 0.0,
  "beneficiary": {
    "name": "John Doe",
    "account": "1234567890",
    "ifsc": "IFSC0001234",
    "bank": "Example Bank"
  },
  "status": "success",
  "message": "Transaction successful",
  "created_at": "2025-01-01T12:00:00Z",
  "updated_at": "2025-01-01T12:00:01Z",
  "processed_at": "2025-01-01T12:00:01Z",
  "notified_at": "2025-01-01T12:00:02Z"
}
```

## Payment Flow

### End-to-End Payment Processing

```
1. Payment Request Received
   ↓
2. Validation
   ├─ Duplicate Reference ID Check
   ├─ Beneficiary Validation
   ├─ Channel Validation
   └─ Amount Limit Check
   ↓
3. Transaction Creation
   ├─ Generate Transaction ID
   ├─ Create Transaction Record (status: INITIATED)
   └─ Return Transaction to Client
   ↓
4. Async Processing (Background Worker)
   ├─ Update Status to PROCESSING
   ├─ Simulate Processing Delay
   ├─ Check Account Balance
   ├─ Deduct Amount + Fee
   └─ Update Status (SUCCESS or FAILED)
   ↓
5. Notification (Background Worker)
   ├─ Fetch Transaction Details
   ├─ Extract Notification URL from Metadata
   ├─ Send HTTP POST to Notification URL
   └─ Update notified_at Timestamp
```

### Payment Processing Steps

#### Step 1: Payment Request Validation

When a payment request is received:

1. **Idempotency Check**: Verifies if a transaction with the same `reference_id` already exists
   - If exists → Returns `REFERENCE_ID_ALREADY_PROCESSED` error
   - If not → Proceeds

2. **Beneficiary Validation**: Validates beneficiary details
   - Checks IFSC code against invalid IFSC list
   - Checks account number against invalid account list
   - Returns appropriate errors if validation fails

3. **Channel Validation**: Verifies payment channel exists and is configured
   - Returns `INVALID_PAYMENT_CHANNEL` if channel not found

4. **Amount Limit Check**: Validates amount against channel limit
   - UPI/IMPS: Checks against configured limit
   - NEFT: No limit check (unlimited)
   - Returns `LIMIT_EXCEEDED` if amount exceeds limit

#### Step 2: Transaction Creation

1. **Generate Transaction ID**: Channel-specific ID generation
   - UPI: `UPI-{timestamp}-{random}`
   - IMPS: `IMPS-{timestamp}-{random}`
   - NEFT: `NEFT-{timestamp}-{random}`

2. **Create Transaction Record**:
   - Status: `INITIATED`
   - Store all payment details
   - Store metadata (including notification URL)

3. **Return Transaction**: Immediately return transaction to client

#### Step 3: Asynchronous Processing

The payment is sent to a processor channel and picked up by the Processor Worker:

1. **Status Update**: Updates transaction status to `PROCESSING`

2. **Processing Delay**: Simulates real-world processing delay
   - UPI: 0-2000ms random delay
   - IMPS: 0-1500ms random delay
   - NEFT: 0-1500ms random delay

3. **Success Rate Check**: Randomly determines success/failure based on channel's `success_rate`
   - If random > success_rate → Transaction fails
   - Failure reason randomly selected from failure list

4. **Account Balance Check**:
   - Fetches account from database
   - Calculates total amount (amount + fee)
   - Checks if balance >= total amount

5. **Balance Deduction** (if successful):
   - Uses database transaction for atomicity
   - Updates account balance (balance - total_amount)
   - Uses optimistic locking to prevent race conditions
   - If balance changed concurrently → Transaction fails

6. **Status Update**:
   - **Success**: Updates status to `SUCCESS`, sets `processed_at`
   - **Failure**: Updates status to `FAILED`, sets error message

#### Step 4: Notification

After successful processing, transaction ID is sent to notifier channel:

1. **Fetch Transaction**: Retrieves transaction details from database

2. **Extract Notification URL**: Gets `notification_url` from transaction metadata

3. **Send Notification**: HTTP POST request to notification URL with payload:
```json
{
  "transaction_id": "TXN-123456789012",
  "reference_id": "REF-123456789012",
  "status": "success",
  "message": "Transaction successful",
  "amount": 50000.0,
  "fee": 0.0,
  "channel": "UPI",
  "created_at": "2025-01-01T12:00:00Z",
  "updated_at": "2025-01-01T12:00:01Z",
  "processed_at": "2025-01-01T12:00:01Z"
}
```

4. **Update Timestamp**: Updates `notified_at` timestamp on success

5. **Error Handling**: Logs errors but doesn't retry (non-blocking)

## Background Workers

The service runs two background workers concurrently:

### 1. Processor Worker (`StartProcessor`)

**Purpose**: Process payment transactions asynchronously

**Trigger**: Receives `ProcessorMessage` via `processor` channel

**Processing**:
- Updates transaction status to `PROCESSING`
- Simulates processing delay (channel-specific)
- Checks success rate (random failure simulation)
- Validates account balance
- Deducts amount + fee from account
- Updates transaction status (`SUCCESS` or `FAILED`)
- Sends transaction ID to notifier channel on success

**Failure Scenarios**:
- Insufficient balance
- Account not found
- Concurrent balance update conflicts
- Random failures based on success rate

### 2. Notifier Worker (`StartNotifier`)

**Purpose**: Send payment status notifications to external services

**Trigger**: Receives transaction IDs via `notifier` channel

**Processing**:
- Fetches transaction details
- Extracts notification URL from metadata
- Sends HTTP POST request to notification URL
- Updates `notified_at` timestamp on success
- Logs errors (non-blocking)

**Error Handling**:
- Missing notification URL: Logged as error
- Network errors: Logged but don't block
- HTTP errors (non-2xx): Logged with status code
- Database errors: Logged but don't retry

## Payment Channels

### UPI (Unified Payments Interface)

- **Limit**: Configurable (default: ₹1,00,000)
- **Success Rate**: Configurable (default: 0.95)
- **Fee**: Configurable (default: ₹0)
- **Processing Delay**: 0-2000ms random
- **Use Case**: Small amounts, instant transfers

### IMPS (Immediate Payment Service)

- **Limit**: Configurable (default: ₹5,00,000)
- **Success Rate**: Configurable (default: 0.90)
- **Fee**: Configurable (default: ₹5)
- **Processing Delay**: 0-1500ms random
- **Use Case**: Medium amounts, instant transfers

### NEFT (National Electronic Funds Transfer)

- **Limit**: Unlimited
- **Success Rate**: Configurable (default: 0.85)
- **Fee**: Configurable (default: ₹10)
- **Processing Delay**: 0-1500ms random
- **Use Case**: Large amounts, batch processing

## Transaction States

```
INITIATED
   │
   ├─→ PROCESSING (in processor worker)
   │      │
   │      ├─→ SUCCESS (balance sufficient, processing successful)
   │      │      │
   │      │      └─→ Notification sent
   │      │
   │      └─→ FAILED (insufficient balance, processing failed, or random failure)
   │
   └─→ FAILED (immediate validation failure)
```

## Database Schema

The service uses PostgreSQL with the following main tables:

- **accounts**: Payment account with balance and threshold
- **payment_channels**: Channel configuration (limit, success_rate, fee)
- **transactions**: Complete audit trail of all payment attempts

## Error Handling

The service handles errors gracefully:

- **Validation Errors**: Return 400 Bad Request with error message
- **Database Errors**: Logged and returned as 500 Internal Server Error
- **Not Found Errors**: Return appropriate error messages
- **Processing Errors**: Transaction marked as failed, error logged
- **Notification Errors**: Logged but don't block transaction completion

## Logging

The service uses `zerolog` for structured logging. Logs include:

- Request/response details
- Error traces
- Worker processing status
- Transaction state changes
- Notification attempts

## Graceful Shutdown

The service supports graceful shutdown:

- Listens for SIGINT and SIGTERM signals
- Stops background workers gracefully
- Closes HTTP server connections
- Ensures in-flight requests complete

## Testing

Run tests:
```bash
go test ./...
```

## Dependencies

Key dependencies:

- `github.com/gorilla/mux`: HTTP router
- `gorm.io/gorm`: ORM for database operations
- `gorm.io/driver/postgres`: PostgreSQL driver
- `github.com/rs/zerolog`: Structured logging

## Integration with Disbursement Service

The payment gateway integrates with the disbursement service:

1. **Payment Request**: Disbursement service sends payment requests to gateway
2. **Notification**: Gateway sends payment status updates to disbursement service's notification endpoint
3. **Transaction Lookup**: Disbursement service can query transaction status

### Notification URL Format

The notification URL should be provided in the payment request metadata:
```json
{
  "metadata": {
    "notification_url": "http://disbursement-service:7070/api/v1/payment/notification"
  }
}
```

## Key Design Decisions

1. **Asynchronous Processing**: Payments processed asynchronously to handle high volume
2. **Idempotency**: Reference IDs prevent duplicate payments
3. **Channel-Specific Providers**: Separate implementations for each channel
4. **Configurable Success Rates**: Simulates real-world failure scenarios
5. **Webhook Notifications**: Event-driven architecture for status updates
6. **Atomic Balance Updates**: Database transactions ensure consistency
7. **Non-Blocking Notifications**: Notification failures don't affect payment processing
