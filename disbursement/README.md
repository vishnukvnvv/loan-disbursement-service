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

## Payment Flow

The payment processing flow involves multiple components working together to ensure reliable and efficient disbursement of loans. Here's a detailed explanation of how payments flow through the system:

### End-to-End Payment Flow

```
1. Disbursement Creation
   ↓
2. Background Worker Picks Up Disbursement
   ↓
3. Payment Service Processing
   ├─ Channel Selection
   ├─ Channel Availability Check
   ├─ Transaction Creation
   └─ Payment Gateway Transfer Request
   ↓
4. Payment Gateway Processing
   ├─ Transaction Validation
   ├─ Account Balance Check
   ├─ Payment Execution
   └─ Status Update
   ↓
5. Notification Flow
   ├─ Gateway Notifier Worker
   ├─ HTTP POST to Disbursement Service
   └─ Status Update in Disbursement Service
   ↓
6. Final Status Update
   ├─ Success → Disbursement marked as SUCCESS
   └─ Failure → Disbursement marked as SUSPENDED or FAILED
```

### Detailed Flow Steps

#### 1. Disbursement Creation
When a disbursement is created via the API (`POST /api/v1/disburse`):
- A disbursement record is created with status `INITIATED`
- The disbursement is linked to a loan and beneficiary
- The disbursement ID is returned to the client

#### 2. Background Worker Processing
Three background workers continuously monitor and process disbursements:

**a) Payment Worker** (`StartPaymentDisbursement`):
- Listens on a channel for disbursement IDs
- Processes individual disbursements on-demand
- Skips NEFT transactions (handled separately)
- Used for immediate processing when disbursements are created

**b) Retry Worker** (`StartRetryDisbursement`):
- Runs every 5 seconds (configurable via `retryPollInterval`)
- Fetches disbursements with status `SUSPENDED`
- Only processes UPI and IMPS channels (NEFT handled separately)
- Checks if retry is eligible based on exponential backoff policy
- Processes batches of up to `retryBatchSize` disbursements

**c) NEFT Worker** (`StartNEFTDisbursement`):
- Runs every 30 seconds (configurable via `neftPollInterval`)
- Fetches disbursements with status `INITIATED` or `SUSPENDED`
- Only processes NEFT channel transactions
- Processes batches of up to `neftBatchSize` disbursements
- Handles pagination automatically

#### 3. Payment Service Processing (`Process` method)

When a worker calls `paymentService.Process(disbursement)`:

**Step 3.1: Eligibility Check**
- Checks if disbursement should be processed:
  - `INITIATED`: Always eligible
  - `PROCESSING`: Not eligible (already being processed)
  - `SUCCESS`: Not eligible (already completed)
  - `SUSPENDED`: Eligible only if retry backoff time has elapsed
  - `FAILED`: Not eligible (permanent failure)

**Step 3.2: Data Retrieval**
- Fetches loan details
- Fetches beneficiary information

**Step 3.3: Channel Selection**
- **Initial Selection** (retryCount = 0):
  - Amount ≤ ₹1,00,000 → UPI
  - Amount ≤ ₹5,00,000 → IMPS
  - Amount > ₹5,00,000 → NEFT
- **Retry Selection** (retryCount > 0):
  - If retryCount == 2 and amount ≤ ₹1,00,000 → Switch from UPI to IMPS
  - Otherwise → Switch to NEFT (most reliable)

**Step 3.4: Status Transition**
- Updates disbursement status to `PROCESSING`
- Updates channel if changed

**Step 3.5: Channel Availability Check**
- Checks if selected channel is active via payment gateway
- If inactive, falls back:
  - UPI → IMPS
  - IMPS → Error (no fallback)
  - NEFT → Error (no fallback)

**Step 3.6: Transaction Creation**
- Generates unique transaction ID and reference ID
- Creates transaction record with status `INITIATED`
- Links transaction to disbursement

**Step 3.7: Payment Gateway Transfer**
- Sends transfer request to payment gateway with:
  - Reference ID (unique identifier)
  - Amount and channel
  - Beneficiary details
  - Notification URL (callback endpoint)
- Payment gateway returns immediately with transaction ID

#### 4. Payment Gateway Processing

The payment gateway receives the transfer request and:

**Step 4.1: Transaction Validation**
- Checks for duplicate reference IDs (idempotency)
- Validates beneficiary details (IFSC, account number)
- Checks payment channel availability and limits

**Step 4.2: Transaction Creation**
- Creates transaction record with status `INITIATED`
- Stores notification URL in metadata

**Step 4.3: Async Processing**
- Sends transaction to processor channel
- Processor worker picks up transaction

**Step 4.4: Payment Execution** (in Processor Worker)
- Updates transaction status to `PROCESSING`
- Simulates processing delay
- Checks account balance
- Deducts amount + fee from account
- Updates transaction status:
  - `SUCCESS` if balance sufficient
  - `FAILED` if insufficient balance or other errors

**Step 4.5: Notification Trigger**
- On successful processing, sends transaction ID to notifier channel
- On failure, transaction is marked as failed (no notification)

#### 5. Notification Flow

**Step 5.1: Notifier Worker** (in Payment Gateway)
- Listens on notifier channel for transaction IDs
- Fetches transaction details
- Extracts notification URL from metadata
- Sends HTTP POST request to notification URL with:
  ```json
  {
    "transaction_id": "TXN-123",
    "reference_id": "REF-456",
    "status": "success",
    "message": "Transaction successful",
    "amount": 50000.0,
    "fee": 10.0,
    "channel": "UPI",
    "created_at": "2025-01-01T12:00:00Z",
    "updated_at": "2025-01-01T12:00:01Z",
    "processed_at": "2025-01-01T12:00:01Z"
  }
  ```
- Updates transaction `notified_at` timestamp on success

**Step 5.2: Disbursement Service Receives Notification**
- Payment handler receives POST request at `/api/v1/payment/notification`
- Calls `HandleNotification` service method
- Looks up transaction by reference ID
- Fetches associated disbursement

**Step 5.3: Status Update**
- **If Success**:
  - Updates transaction status to `SUCCESS`
  - Updates disbursement status to `SUCCESS`
  - Both updates happen in a database transaction
- **If Failure**:
  - Calls `HandleFailure` method
  - Updates transaction status to `FAILED`
  - Evaluates failure:
    - If retry count < 5 and error is retriable → `SUSPENDED`
    - Otherwise → `FAILED`
  - Updates disbursement with new status and retry count

#### 6. Failure Handling

When a payment fails:

**Immediate Failure** (during transfer request):
- Payment gateway returns error immediately
- `HandleFailure` is called with the error
- Status evaluated and updated

**Delayed Failure** (during processing):
- Payment gateway processes transaction
- Transaction marked as `FAILED`
- Notification sent with failure status
- Disbursement service receives notification and updates status

**Reference ID Already Processed**:
- If duplicate reference ID detected:
  - Fetches actual payment status from gateway
  - If successful → Updates transaction and disbursement to success
  - If failed → Uses actual error for failure evaluation

### Background Workers

The service runs three background workers concurrently:

#### 1. Payment Worker (`StartPaymentDisbursement`)

**Purpose**: Process individual disbursements on-demand

**Trigger**: Receives disbursement IDs via channel (`paymentChan`)

**Processing**:
- Fetches disbursement by ID
- Skips NEFT transactions (handled by NEFT worker)
- Calls `paymentService.Process()` for non-NEFT disbursements

**Use Cases**:
- Immediate processing when disbursement is created
- Manual retry via API

#### 2. Retry Worker (`StartRetryDisbursement`)

**Purpose**: Automatically retry suspended UPI/IMPS disbursements

**Schedule**: Runs every 5 seconds (configurable)

**Query**:
- Status: `SUSPENDED`
- Channels: `UPI`, `IMPS`
- Batch size: Configurable (default in code)

**Processing**:
- Fetches batch of suspended disbursements
- For each disbursement:
  - Checks if retry is eligible (backoff time elapsed)
  - Calls `paymentService.Process()` if eligible
- Continues pagination until no more disbursements

**Retry Eligibility**:
- Checks exponential backoff policy
- Only processes if enough time has passed since last update
- Prevents thundering herd with jitter

#### 3. NEFT Worker (`StartNEFTDisbursement`)

**Purpose**: Process NEFT transactions separately (slower, batch-oriented)

**Schedule**: Runs every 30 seconds (configurable)

**Query**:
- Status: `INITIATED` or `SUSPENDED`
- Channels: `NEFT`
- Batch size: Configurable (default in code)

**Processing**:
- Fetches batch of NEFT disbursements
- Processes each disbursement
- Handles pagination automatically
- Stops when batch size < configured batch size

**Why Separate**:
- NEFT transactions are slower and batch-oriented
- Different processing requirements
- Separate polling interval reduces load

### Notifier System

The notifier system ensures that the disbursement service is informed about payment status changes asynchronously.

#### Architecture

```
Payment Gateway                    Disbursement Service
─────────────────                  ─────────────────────
                                   
Processor Worker                   Payment Handler
     │                                    ▲
     │ (on success)                       │
     ▼                                    │
Notifier Channel                         │
     │                                    │
     │                                    │
Notifier Worker                          │
     │                                    │
     │ HTTP POST                          │
     └────────────────────────────────────┘
```

#### Notifier Worker (Payment Gateway)

**Purpose**: Send payment status notifications to disbursement service

**Trigger**: Receives transaction IDs via `notifier` channel after successful processing

**Process**:
1. Fetches transaction details from database
2. Extracts `notification_url` from transaction metadata
3. Constructs notification payload with transaction details
4. Sends HTTP POST request to notification URL
5. Updates transaction `notified_at` timestamp on success
6. Logs errors if notification fails (non-blocking)

**Notification Payload**:
```json
{
  "transaction_id": "TXN-123456789012",
  "reference_id": "REF-123456789012",
  "status": "success",
  "message": "Transaction successful",
  "amount": 50000.0,
  "fee": 10.0,
  "channel": "UPI",
  "created_at": "2025-01-01T12:00:00Z",
  "updated_at": "2025-01-01T12:00:01Z",
  "processed_at": "2025-01-01T12:00:01Z"
}
```

**Error Handling**:
- Network errors: Logged but don't block transaction
- HTTP errors (non-2xx): Logged with status code
- Missing notification URL: Logged as error
- Database errors: Logged but don't retry

#### Notification Handler (Disbursement Service)

**Endpoint**: `POST /api/v1/payment/notification`

**Process**:
1. Receives `PaymentNotificationRequest` payload
2. Looks up transaction by `reference_id`
3. Fetches associated disbursement
4. Updates status based on notification:
   - **Success**: Updates transaction and disbursement to `SUCCESS`
   - **Failure**: Evaluates failure and updates to `SUSPENDED` or `FAILED`

**Idempotency**:
- Uses reference ID to look up existing transaction
- Prevents duplicate processing
- Handles duplicate notifications gracefully

#### Notification URL Configuration

The notification URL is configured via environment variable:
- `NOTIFICATION_URL`: Base URL for payment notifications
  - Example: `http://disbursement-service:7070/api/v1/payment/notification`
- Passed to payment gateway in transaction metadata
- Payment gateway stores it and uses it for callbacks

### State Transitions

#### Disbursement States

```
INITIATED
   │
   ├─→ PROCESSING (when worker picks up)
   │      │
   │      ├─→ SUCCESS (on successful payment)
   │      │
   │      └─→ SUSPENDED (on retriable failure)
   │             │
   │             └─→ PROCESSING (on retry)
   │                    │
   │                    ├─→ SUCCESS
   │                    │
   │                    └─→ SUSPENDED (if still failing)
   │                           │
   │                           └─→ FAILED (after max retries)
   │
   └─→ FAILED (on non-retriable failure)
```

#### Transaction States

```
INITIATED
   │
   ├─→ PROCESSING (in payment gateway)
   │      │
   │      ├─→ SUCCESS (payment successful)
   │      │      │
   │      │      └─→ Notification sent
   │      │
   │      └─→ FAILED (payment failed)
   │             │
   │             └─→ Notification sent (if applicable)
   │
   └─→ FAILED (immediate failure)
```

### Key Design Decisions

1. **Asynchronous Processing**: Payment gateway processes transactions asynchronously to handle high volume
2. **Separate Workers**: Different workers for different channels optimize processing based on channel characteristics
3. **Notification-Based Updates**: Webhook-style notifications ensure eventual consistency
4. **Idempotency**: Reference IDs prevent duplicate payments
5. **Retry Logic**: Exponential backoff with jitter prevents system overload
6. **Channel Fallback**: Automatic channel switching improves success rates
7. **Batch Processing**: Workers process batches for efficiency

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
