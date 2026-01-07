# Loan Disbursement Service

## Overview

The Loan Disbursement Service is a microservice responsible for managing loan disbursements, processing payments through multiple payment channels (UPI, IMPS, NEFT), and handling retries with intelligent channel switching. It provides APIs for loan management, disbursement creation, status tracking, and reconciliation.

## Features

- **Loan Management**: Create, update, list, and retrieve loan records
- **Disbursement Processing**: Create and track loan disbursements with idempotency guarantees
- **Multi-Channel Payments**: Automatic channel selection (UPI, IMPS, NEFT) based on amount and retry count
- **Intelligent Retry Logic**: Exponential backoff with jitter and automatic channel switching
- **Background Worker**: Polls and processes pending disbursements automatically
- **Reconciliation**: On-demand reconciliation API for matching transactions with bank statements
- **Exactly-Once Guarantee**: Idempotency keys, state machine, and unique reference IDs prevent duplicate payments

## Architecture

The service follows a layered architecture:

- **API Layer** (`api/`): HTTP handlers and routing
- **Service Layer** (`api/services/`): Business logic for loans, disbursements, payments, and reconciliation
- **Data Access Layer** (`db/daos/`): Database operations
- **Models** (`models/`): Domain models and DTOs
- **Worker** (`worker/`): Background processing for pending disbursements
- **Providers** (`providers/`): External service integrations (payment gateway)

## Prerequisites

- Go 1.24.6 or higher
- PostgreSQL database
- Payment Gateway service running (see `gateway/README.md`)

## Environment Variables

The service requires the following environment variables:

- `DATABASE_URL`: PostgreSQL connection string (required)
  - Example: `postgres://user:password@localhost:5432/disbursement_db?sslmode=disable`
- `PAYMENT_PROVIDER_URL`: Base URL of the payment gateway service (required)
  - Example: `http://localhost:8080`

## Installation

1. Navigate to the disbursement directory:
```bash
cd disbursement
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
export DATABASE_URL="postgres://user:password@localhost:5432/disbursement_db?sslmode=disable"
export PAYMENT_PROVIDER_URL="http://localhost:8080"
```

5. Run the service:
```bash
go run main.go
```

The service will start on port `7070` by default.

## API Endpoints

All endpoints are prefixed with `/api/v1`

### Loan Management

#### Create Loan
- **Method**: `POST`
- **Path**: `/api/v1/loan`
- **Request Body**:
```json
{
  "amount": 50000.0
}
```
- **Response** (200):
```json
{
  "id": "LOANxxxxxxxxxxxx",
  "amount": 50000.0,
  "disbursed": false,
  "created_at": "2025-01-01T12:00:00Z",
  "updated_at": "2025-01-01T12:00:00Z"
}
```

#### Get Loan
- **Method**: `GET`
- **Path**: `/api/v1/loan/{id}`
- **Response** (200): Same as Create Loan response
- **Error** (404): Loan not found

#### List Loans
- **Method**: `GET`
- **Path**: `/api/v1/loan`
- **Response** (200):
```json
[
  {
    "id": "LOANxxxxxxxxxxxx",
    "amount": 50000.0,
    "disbursed": false,
    "created_at": "2025-01-01T12:00:00Z",
    "updated_at": "2025-01-01T12:00:00Z"
  }
]
```

#### Update Loan
- **Method**: `PUT`
- **Path**: `/api/v1/loan/{id}`
- **Request Body**:
```json
{
  "amount": 60000.0
}
```
- **Response** (200): Updated loan object
- **Error** (404): Loan not found

### Disbursement Management

#### Create Disbursement
- **Method**: `POST`
- **Path**: `/api/v1/disburse`
- **Request Body**:
```json
{
  "loan_id": "LOANxxxxxxxxxxxx",
  "amount": 50000.0,
  "beneficiary_name": "John Doe",
  "account_number": "1234567890",
  "ifsc_code": "IFSC0001234",
  "beneficiary_bank": "Example Bank"
}
```
- **Response** (200):
```json
{
  "disbursement_id": "DISxxxxxxxxxxxx",
  "status": "initiated",
  "message": "Disbursement created"
}
```
- **Note**: If a disbursement already exists for the loan, returns existing disbursement (idempotent)

#### Get Disbursement
- **Method**: `GET`
- **Path**: `/api/v1/disburse/{id}`
- **Response** (200):
```json
{
  "disbursement_id": "DISxxxxxxxxxxxx",
  "amount": 50000.0,
  "loan_id": "LOANxxxxxxxxxxxx",
  "status": "processing",
  "transaction": [
    {
      "transaction_id": "REFxxxxxxxxxxxx",
      "status": "completed",
      "mode": "UPI",
      "message": null,
      "created_at": "2025-01-01T12:00:00Z",
      "updated_at": "2025-01-01T12:00:01Z"
    }
  ],
  "created_at": "2025-01-01T12:00:00Z",
  "updated_at": "2025-01-01T12:00:01Z"
}
```

#### Retry Disbursement
- **Method**: `POST`
- **Path**: `/api/v1/disburse/{id}`
- **Response** (200):
```json
{
  "disbursement_id": "DISxxxxxxxxxxxx",
  "status": "initiated",
  "message": "Disbursement retried"
}
```
- **Error** (400): Disbursement is in-progress or already completed

## Disbursement Status Flow

The disbursement follows this state machine:

```
INITIATED → PROCESSING → SUCCESS
         ↓              ↓
         → SUSPENDED    → SUSPENDED
         ↓
         → FAILURE
```

- **INITIATED**: Disbursement created, waiting for processing
- **PROCESSING**: Currently being processed by the worker
- **SUSPENDED**: Temporary failure, eligible for retry after backoff period
- **SUCCESS**: Payment completed successfully
- **FAILURE**: Permanent failure, no further retries

## Channel Selection Strategy

The service automatically selects payment channels based on amount and retry count:

### Initial Selection (retryCount = 0)
- **UPI**: Amount ≤ ₹1,00,000 (free, instant)
- **IMPS**: Amount ≤ ₹5,00,000 (instant, reasonable cost)
- **NEFT**: Amount > ₹5,00,000 (no limit, high reliability)

### Fallback on Retry (retryCount > 0)
- If retryCount == 2 and original channel was UPI (amount ≤ ₹1,00,000): Switch to IMPS
- Otherwise: Switch to NEFT (most reliable fallback)

## Retry Policy

- **Max Retries**: 5 attempts total
- **Backoff Strategy**: Exponential backoff with jitter
  - Initial delay: 30 seconds
  - Max delay: 30 minutes
  - Jitter: ±20% to prevent thundering herd
- **Retriable Failures**: Gateway errors, limit exceeded, bank down, inactive account (temporary)
- **Non-Retriable Failures**: Invalid IFSC, account closed, regulatory restrictions

## Background Worker

The service includes a background worker that:
- Polls the database every 5 seconds for pending disbursements
- Processes disbursements with status "initiated" or "suspended" (if retry time elapsed)
- Processes up to 10 disbursements per batch
- Automatically retries suspended disbursements based on exponential backoff

## Database Schema

The service uses PostgreSQL with the following main tables:

- **beneficiaries**: KYC-verified recipient information
- **loans**: Approved loan records
- **disbursements**: One per disbursement request (idempotency boundary)
- **transactions**: One per payment attempt (complete audit trail)

## Reconciliation

The service provides reconciliation capabilities to match internal transactions with bank statements. The reconciliation service is available in `api/services/reconcil.go` and can be integrated via API endpoint.

### Reconciliation Request Format

```json
{
  "statement_date": "2025-01-06",
  "transactions": [
    {
      "reference_id": "REFxxxxxxxxxxxx",
      "amount": 5000.00,
      "date": "2025-01-06T12:00:00Z",
      "status": "SUCCESS"
    }
  ]
}
```

### Reconciliation Response

```json
{
  "reconciliation_id": "uuid",
  "statement_date": "2025-01-06",
  "total_expected": 1000000.00,
  "total_actual": 995000.00,
  "matched_count": 198,
  "discrepancies": [
    {
      "type": "missing",
      "reference_id": "REF123",
      "expected_amount": 5000.00,
      "actual_amount": 0,
      "message": "Transaction marked as SUCCESS in our records but not found in bank statement"
    }
  ]
}
```

## Error Handling

The service handles errors gracefully:
- Database errors are logged and returned as 500 Internal Server Error
- Validation errors return 400 Bad Request
- Not found errors return 404 Not Found
- Payment gateway errors are classified and handled according to retry policy

## Logging

The service uses `zerolog` for structured logging. Logs include:
- Request/response details
- Error traces
- Worker processing status
- Payment gateway interactions

## Graceful Shutdown

The service supports graceful shutdown:
- Listens for SIGINT and SIGTERM signals
- Stops the background worker gracefully
- Closes HTTP server connections
- Ensures in-flight requests complete

## Design Decisions

For detailed architectural decisions, see `../DESIGN_DECISIONS.md` in the parent directory.

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
- `github.com/google/uuid`: UUID generation

## Troubleshooting

### Database Connection Issues
- Verify `DATABASE_URL` is correctly set
- Ensure PostgreSQL is running and accessible
- Check database credentials and permissions

### Payment Gateway Issues
- Verify `PAYMENT_PROVIDER_URL` is correct
- Ensure the gateway service is running
- Check network connectivity between services

### Worker Not Processing
- Check logs for errors
- Verify disbursements exist with status "initiated" or "suspended"
- Ensure retry backoff time has elapsed for suspended disbursements

## License

[Add your license information here]

