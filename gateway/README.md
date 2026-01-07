# Mock Payment Gateway Service

## Overview

The Mock Payment Gateway is a service that simulates payment processing for UPI, IMPS, and NEFT payment channels. It provides HTTP APIs to initialize and query payment transactions, with configurable failure rates, timeouts, and validation rules.

The service is built on `gin` and loads configuration from YAML. It applies beneficiary and limit validations, and routes payments to mocked channels (UPI / NEFT / IMPS).

## Features

- **Multiple Payment Modes**: Supports UPI, IMPS, and NEFT
- **Configurable Behavior**: Failure rates, timeouts, and limits per payment mode
- **Beneficiary Validation**: IFSC code, account number, and bank status validation
- **Idempotency**: Reference ID tracking to prevent duplicate processing
- **Transaction Status Tracking**: Query payment status by transaction ID

## Prerequisites

- Go 1.24.6 or higher

## Installation

1. Navigate to the gateway directory:
```bash
cd gateway
```

2. Install dependencies:
```bash
go mod download
```

3. Configure the service by editing `config/config.yaml` (see Configuration section)

4. Run the service:
```bash
go run main.go
```

The server will start on the port specified in `config/config.yaml` (default: 8080).

---

## Configuration

Configuration is loaded from `config/config.yaml`. The configuration file supports:

- **`server.port`**: Port number the HTTP server listens on (default: 8080)
- **`invalid_ifsc`**: List of IFSC codes considered invalid (will trigger `INVALID_IFSC` error)
- **`invalid_account_number`**: List of beneficiary account numbers considered invalid (will trigger `INACTIVE_ACCOUNT` error)
- **`beneficiary_bank_down`**: List of bank names treated as "down" (will trigger `BENEFICIARY_BANK_DOWN` error)
- **`payment_modes`**: Configuration for each payment mode:
  - **`upi`**:
    - `limit`: Maximum transaction amount (default: 100000)
    - `timeout`: Processing timeout in seconds (default: 3)
    - `failure_rate`: Probability of failure (0.0 to 1.0, default: 0.09)
  - **`neft`**:
    - `timeout`: Processing timeout in seconds (default: 6)
    - `failure_rate`: Probability of failure (0.0 to 1.0, default: 0.005)
  - **`imps`**:
    - `limit`: Maximum transaction amount (default: 500000)
    - `timeout`: Processing timeout in seconds (default: 4.5)
    - `failure_rate`: Probability of failure (0.0 to 1.0, default: 0.06)

### Example Configuration

```yaml
server:
  port: 8080

invalid_ifsc:
  - IFSC1
  - IFSC2
  - IFSC3

invalid_account_number:
  - ACCOUNT_NUMBER1
  - ACCOUNT_NUMBER2
  - ACCOUNT_NUMBER3

beneficiary_bank_down:
  - BANK1
  - BANK2
  - BANK3

payment_modes:
  upi:
    limit: 100000
    timeout: 3
    failure_rate: 0.09
  neft:
    timeout: 6
    failure_rate: 0.005
  imps:
    limit: 500000
    timeout: 4.5
    failure_rate: 0.06
```

---

## API Endpoints

All endpoints are prefixed with: **`/api/v1/payment`**

### 1. Initialize Payment

- **Method**: `POST`
- **Path**: `/api/v1/payment/init`
- **Description**: Creates a new payment transaction and processes it through the specified payment channel

**Request Body:**

```json
{
  "amount": 1500.0,
  "mode": "UPI",
  "beneficiary": {
    "name": "John Doe",
    "account": "1234567890",
    "ifsc": "IFSC0001234",
    "bank": "Example Bank"
  },
  "metadata": {
    "order_id": "ORD-123"
  },
  "reference_id": "REFxxxxxxxxxxxx"
}
```

**Request Fields:**
- `amount` (number, required): Payment amount in currency units
- `mode` (string, required): Payment mode – must be one of `"UPI"`, `"NEFT"`, or `"IMPS"`
- `beneficiary` (object, required): Beneficiary details
  - `name` (string, required): Beneficiary name
  - `account` (string, required): Beneficiary account number
  - `ifsc` (string, required): IFSC code
  - `bank` (string, required): Bank name
- `metadata` (object, optional): Arbitrary key-value pairs for additional information
- `reference_id` (string, required): Unique reference ID for idempotency

**Success Response (200):**

```json
{
  "transactionId": "UPIxxxxxxxxxxxx",
  "amount": 1500.0,
  "status": "SUCCESS",
  "error": null,
  "mode": "UPI",
  "beneficiary": {
    "name": "John Doe",
    "account": "1234567890",
    "ifsc": "IFSC0001234",
    "bank": "Example Bank"
  },
  "acceptedAt": "2025-01-01T12:00:00Z",
  "processedAt": "2025-01-01T12:00:01Z",
  "metadata": {
    "order_id": "ORD-123"
  }
}
```

**Error Responses:**

- **400 Bad Request** (Validation / Business Failure):
```json
{
  "error": "Invalid IFSC code"
}
```

- **400 Bad Request** (Duplicate Reference ID):
```json
{
  "error": "Reference ID already processed"
}
```

- **400 Bad Request** (Limit Exceeded):
```json
{
  "error": "Limit Exceeded"
}
```

### 2. Fetch Payment Status

- **Method**: `GET`
- **Path**: `/api/v1/payment/status/:id`
- **Description**: Retrieves the status of a payment transaction by transaction ID

**Path Parameters:**
- `:id` (string, required): Transaction ID returned by the init API (e.g., `"UPIxxxxxxxxxxxx"`, `"NEFTxxxxxxxxxxxx"`, `"IMPSxxxxxxxxxxxx"`)

**Success Response (200):**

```json
{
  "transactionId": "UPIxxxxxxxxxxxx",
  "amount": 1500.0,
  "status": "SUCCESS",
  "error": null,
  "mode": "UPI",
  "beneficiary": {
    "name": "John Doe",
    "account": "1234567890",
    "ifsc": "IFSC0001234",
    "bank": "Example Bank"
  },
  "acceptedAt": "2025-01-01T12:00:00Z",
  "processedAt": "2025-01-01T12:00:01Z",
  "metadata": {
    "order_id": "ORD-123"
  }
}
```

**Error Responses:**

- **400 Bad Request** (Missing ID):
```json
{
  "error": "transaction ID is required"
}
```

- **404 Not Found** (Invalid Transaction ID):
```json
{
  "error": "Invalid transaction ID"
}
```

---

## Payment Status Values

- **`SUCCESS`**: Payment processed successfully
- **`FAILED`**: Payment failed (non-retriable error)
- **`PENDING`**: Payment is being processed (rare, typically instant)

## Failure Modes & Error Messages

Business failures are defined in `failures/code.go` and returned via the API as:

```json
{ "error": "<error message>" }
```

### Key Failure Types

- **`INVALID_IFSC`** – `"Invalid IFSC code"`
  - Triggered when `beneficiary.ifsc` is in `config.invalid_ifsc` list
  - Non-retriable error

- **`INACTIVE_ACCOUNT`** – `"Inactive Beneficiary Account"`
  - Triggered when `beneficiary.account` is in `config.invalid_account_number` list
  - May be retriable (account could be activated)

- **`BENEFICIARY_BANK_DOWN`** – `"Beneficiary Bank is Down"`
  - Triggered when `beneficiary.bank` is in `config.beneficiary_bank_down` list
  - Retriable error (temporary bank outage)

- **`LIMIT_EXCEEDED`** – `"Limit Exceeded"`
  - Triggered when the `amount` exceeds the configured limit for the selected payment mode
  - UPI limit: default ₹1,00,000
  - IMPS limit: default ₹5,00,000
  - NEFT: no limit
  - May be retriable (rate limiting)

- **`Reference ID already processed`**
  - Triggered when the same `reference_id` is used twice
  - Ensures idempotency
  - Non-retriable error (use a different reference_id)

### Other Possible Errors

- **Invalid payment mode**: When `mode` is not one of `"UPI"`, `"NEFT"`, or `"IMPS"`
- **Invalid transaction ID**: When status is requested for an ID that doesn't match any stored transaction or has an unrecognized prefix

## Payment Processing Behavior

### UPI
- **Limit**: Configurable (default: ₹1,00,000)
- **Timeout**: Configurable (default: 3 seconds)
- **Failure Rate**: Configurable (default: 9%)
- **Transaction ID Format**: `UPI` + random string

### IMPS
- **Limit**: Configurable (default: ₹5,00,000)
- **Timeout**: Configurable (default: 4.5 seconds)
- **Failure Rate**: Configurable (default: 6%)
- **Transaction ID Format**: `IMPS` + random string

### NEFT
- **Limit**: None (unlimited)
- **Timeout**: Configurable (default: 6 seconds)
- **Failure Rate**: Configurable (default: 0.5%)
- **Transaction ID Format**: `NEFT` + random string

## Idempotency

The gateway tracks `reference_id` values to prevent duplicate processing. If the same `reference_id` is submitted twice, the second request will return an error: `"Reference ID already processed"`.

## Testing

Run tests:
```bash
go test ./...
```

## Dependencies

Key dependencies:
- `github.com/gin-gonic/gin`: HTTP web framework
- `go.yaml.in/yaml/v2`: YAML configuration parsing
- `github.com/google/uuid`: UUID generation
- `github.com/rs/zerolog`: Structured logging

## Troubleshooting

### Server Won't Start
- Verify `config/config.yaml` exists and is valid YAML
- Check that the configured port is not already in use
- Ensure all required configuration fields are present

### Payments Always Failing
- Check `failure_rate` configuration for the payment mode
- Verify beneficiary details are not in invalid lists
- Check that amount doesn't exceed payment mode limits

### Transaction Not Found
- Verify transaction ID format matches expected prefix (UPI/IMPS/NEFT)
- Ensure transaction was successfully created (check init API response)
- Transaction IDs are case-sensitive

## Architecture

The service follows a simple layered architecture:

- **API Layer** (`api/`): HTTP handlers and routing
- **Service Layer** (`api/services/`): Business logic for payment processing
- **Payment Layer** (`payment/`): Payment channel implementations (UPI, IMPS, NEFT)
- **Config** (`config/`): Configuration loading and management
- **Models** (`models/`): Domain models and DTOs
- **Failures** (`failures/`): Error code definitions

## License

[Add your license information here]
