# Design Decisions - Loan Disbursement Service

## Overview
This document explains the key architectural and implementation decisions made for the loan disbursement service.

## 1. Channel Selection Strategy

### Decision: Amount-Based Routing with Retry-Driven Fallback
We use a tiered approach based on amount and retry count:

- **Initial selection (retryCount = 0)**:
  - UPI for amounts ≤ ₹1,00,000 (free, instant)
  - IMPS for amounts ≤ ₹5,00,000 (instant, reasonable cost)
  - NEFT for amounts > ₹5,00,000 (no limit, high reliability)
  
- **Fallback on retry (retryCount > 0)**:
  - If retryCount == 2 and original channel was UPI (amount ≤ ₹1,00,000): Switch to IMPS
  - Otherwise: Switch to NEFT (most reliable fallback)
  
- **Rationale**: Start with cost-optimized channels, escalate to more reliable channels on failure

## 2. Exactly-Once Guarantee

### Decision: Idempotency Keys + State Machine + Reference ID Uniqueness

**Three-layer approach:**

1. **Application-level idempotency**: 
   - Generate deterministic idempotency key from loan_id
   - Check disbursements table before creating new record
   - Prevents duplicate disbursement creation

2. **State machine with valid transitions**:
   ```
   INITIATED → PROCESSING → SUCCESS
            ↓              ↓
            → SUSPENDED    → SUSPENDED
            ↓
            → FAILURE
   ```
   - Only valid state transitions are allowed
   - Processing state acts as distributed lock

3. **Payment gateway reference IDs**:
   - Generate unique reference_id (format: "REF" + UUID) per transaction attempt BEFORE payment call
   - Store reference_id in transaction record for audit trail
   - Gateways reject duplicate reference_ids, enabling safe retries with new reference_ids
   - Each retry generates a new reference_id, preventing duplicate payment attempts

## 3. Retry Policy

### Decision: Exponential Backoff with Jitter + Failure Classification

**Retry strategy:**
- **Max retries**: 5 attempts total
- **Retry count logic**: `retryCount >= MaxRetries` means already exhausted retries → mark as FAILED
- **Retry eligibility**: Checked using `IsRetryEligible()` which compares current time with calculated next retry time
- **Retriable failures**: Classified as SUSPENDED status, eligible for retry based on backoff calculation
- **Non-retriable failures**: Classified as FAILED status, no retry

**Backoff formula**: `delay = min(initial_delay * 2^retryCount * (1 + random_jitter), max_delay)`
- Initial delay: 30 seconds
- Max delay: 30 minutes
- Jitter: ±20% to prevent thundering herd
- Calculation: `nextRetryTime = lastAttemptTime + CalculateBackoff(retryCount)`

**Channel switching:**
- Channel switch occurs at retryCount == 2 (after 2 failed attempts)
- UPI → IMPS (for amounts ≤ ₹1,00,000) or UPI/IMPS → NEFT (for all other cases)
- Prevents being stuck on a degraded channel while minimizing unnecessary switches

### Alternatives Considered:
- **Fixed retry intervals**: Rejected due to thundering herd problem
- **Infinite retries**: Rejected to prevent stuck disbursements
- **Immediate retries**: Rejected as it overwhelms degraded services


## 4. Storage Design

### Decision: PostgreSQL with Event-Sourcing Lite

**Schema structure:**
- **beneficiaries**: KYC-verified recipient information
- **loans**: Approved loan records
- **disbursements**: One per disbursement request (idempotency boundary)
- **transactions**: One per payment attempt (created BEFORE gateway call for complete audit trail)

**Key patterns:**
- **Optimistic locking**: Use updated_at for concurrent update detection
- **Dual status tracking**: 
  - Disbursement status: Overall disbursement state (initiated → processing → success/suspended/failed)
  - Transaction status: Individual payment attempt state (initiated → completed/failed)
- **Transaction-first approach**: Create transaction record with "initiated" status BEFORE gateway call
  - Provides complete audit trail even if gateway call fails immediately
  - Transaction status updated based on gateway response
- **Immutable transactions**: Never delete, only append new transaction records
- **Foreign key cascades**: Maintain referential integrity

**Indexes:**
- `loan_id`, `beneficiary_id`, `reference_id` for lookups
- `status` for background worker queries
- `created_at` for reconciliation time-range queries

## 5. Failure Classification

### Decision: String-Based Error Matching with Default-to-Fail Strategy

**Classification approach:**
- Errors are classified by matching error message strings (gateway returns error messages)
- Default behavior: All errors are treated as non-retriable (FAILED) unless explicitly matched as retriable
- Retry count check: If `retryCount >= MaxRetries` (5), mark as FAILED regardless of error type

**Retriable errors** (SUSPENDED status, eligible for retry):
- "gateway error" - Network/gateway issues
- "Limit Exceeded" - Rate limiting, can retry later
- "Inactive Beneficiary Account" - Account may become active
- "Beneficiary Bank is Down" - Temporary bank outage
- "Reference ID already processed" - Duplicate reference, retry with new reference_id

**Non-retriable errors** (FAILED status, no retry):
- All other errors default to FAILED
- Includes: Invalid IFSC, account closed, regulatory restrictions, etc.
- After MaxRetries (5) reached, always FAILED

**Implementation details:**
- Error classification happens in `evaluateFailure()` method
- Transaction record is updated with failure status and error message
- Disbursement status and retry count updated based on classification
- Retry eligibility checked using exponential backoff calculation before next attempt

## 6. Reconciliation Strategy

### Decision: On-Demand API-Based Reconciliation with Reference ID Matching

**Architecture:**
- Reconciliation is performed via API endpoint (not automated daily batch)
- External system (bank/finance team) provides statement data via API request
- Matching performed in-memory using reference_id as primary key

**Process:**
1. **Input**: Reconciliation request with:
   - `statement_date`: Date in YYYY-MM-DD format
   - `transactions`: Array of bank statement transactions with reference_id, amount, date, status

2. **Matching logic**:
   - Fetch our successful transactions for the statement date
   - Build maps keyed by reference_id for both our records and bank statement
   - Compare transactions by reference_id

3. **Discrepancy detection**:
   - **Missing transactions**: In our records (status=SUCCESS) but not in bank statement
   - **Amount mismatches**: Reference ID matches but amount differs (tolerance: ₹0.01)
   - **Status mismatches**: Reference ID matches but bank status is not SUCCESS/COMPLETED
   - **Ghost transactions**: In bank statement but not in our records (potential fraud/data loss)

**Amount matching:**
- Uses tolerance of ₹0.01 to handle floating-point precision issues
- Formula: `|expected - actual| < 0.01`

**Response structure:**
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
    },
    {
      "type": "amount_mismatch",
      "reference_id": "REF456",
      "expected_amount": 10000.00,
      "actual_amount": 9999.99,
      "message": "Amount mismatch: expected 10000.00, got 9999.99"
    },
    {
      "type": "ghost",
      "reference_id": "REF789",
      "expected_amount": 0,
      "actual_amount": 2000.00,
      "message": "Transaction in bank statement but not in our records - potential fraud or data loss"
    }
  ]
}
```

**Implementation details:**
- Reconciliation service queries transactions by date range (start of day to end of day)
- Filters to only successful transactions for comparison
- All discrepancies include descriptive messages for investigation
- Totals calculated separately for expected (our records) and actual (bank statement SUCCESS/COMPLETED only)


## 7. Background Worker Design

### Decision: Multi-Worker Architecture with Channel-Specific Processing

**Architecture:**
- Three separate goroutine workers, each optimized for different use cases
- No job queue - direct database queries for pending disbursements
- Synchronous batch processing within each polling cycle
- Channel-based separation allows different polling intervals and batch sizes

**Worker Types:**

#### 1. Payment Worker (`StartPaymentDisbursement`)
- **Trigger**: Channel-based, event-driven via `paymentChan`
- **Purpose**: Process individual disbursements on-demand
- **Processing**: 
  - Listens on `paymentChan` for disbursement IDs
  - Fetches disbursement by ID
  - Skips NEFT transactions (handled by NEFT worker)
  - Calls `PaymentService.Process()` synchronously
- **Use Cases**: 
  - Immediate processing when disbursement is created via API
  - Manual retry via API
- **Advantages**: Low latency for immediate processing needs

#### 2. Retry Worker (`StartRetryDisbursement`)
- **Trigger**: Time-based polling via ticker
- **Poll Interval**: 5 seconds (configurable via `retryPollInterval`)
- **Batch Size**: Configurable via `retryBatchSize`
- **Status Filter**: `SUSPENDED` only
- **Channel Filter**: `UPI` and `IMPS` only (NEFT handled separately)
- **Processing Flow**:
  1. **Polling loop**: Worker wakes up every 5 seconds via ticker
  2. **Batch processing loop** (within each polling cycle):
     - Start with offset = 0
     - Fetch batch of suspended UPI/IMPS disbursements
     - Process each disbursement:
       - Eligibility checked by `shouldProcess()` (retry backoff time must have elapsed)
       - Calls `PaymentService.Process()` synchronously
       - Errors logged but don't stop batch processing
     - If batch is full (len == retryBatchSize): increment offset, fetch next batch
     - If batch is partial (len < retryBatchSize): all eligible disbursements processed, exit loop
  3. **Cycle completion**: Worker sleeps until next ticker interval
- **Purpose**: Automatically retry suspended UPI/IMPS transactions with exponential backoff

#### 3. NEFT Worker (`StartNEFTDisbursement`)
- **Trigger**: Time-based polling via ticker
- **Poll Interval**: 30 seconds (configurable via `neftPollInterval`)
- **Batch Size**: Configurable via `neftBatchSize`
- **Status Filter**: `INITIATED` or `SUSPENDED`
- **Channel Filter**: `NEFT` only
- **Processing Flow**: Same as Retry Worker but with different filters and interval
- **Purpose**: Handle NEFT transactions separately due to:
  - Slower processing requirements
  - Batch-oriented nature
  - Different SLA expectations
  - Reduced polling frequency to minimize load

**State Management:**
- Offset counter resets to 0 at start of each polling cycle (for Retry and NEFT workers)
- Within a cycle, processes all eligible disbursements in batches until none remain
- Each disbursement's eligibility checked by `shouldProcess()`:
  - `initiated`: Always eligible
  - `suspended`: Eligible if retry backoff time has elapsed
  - `processing`: Not eligible (prevents concurrent processing)
  - `success`: Not eligible
  - `failed`: Not eligible

**Graceful Shutdown:**
- Supports two shutdown mechanisms:
  1. **Context cancellation**: Workers stop when parent context is cancelled
  2. **Stop channel**: External `Stop()` call closes `stopChan`, all workers stop
- Both mechanisms allow in-flight batches to complete before shutdown
- All three workers share the same `stopChan` for coordinated shutdown

**Error Handling:**
- Database query errors: Logged and return (skip this polling cycle)
- Individual disbursement processing errors: Logged but continue with next disbursement
- Workers continue running despite individual failures
- Payment worker errors don't affect retry/NEFT workers

**Advantages:**
- Simple implementation, no queue infrastructure needed
- Automatic retry via polling (suspended disbursements retried when eligible)
- Stateless design (no in-memory job state)
- Channel-specific optimization (different intervals for different channels)
- Separation of concerns (immediate vs. retry vs. NEFT processing)

**Limitations:**
- Fixed polling intervals (not event-driven for Retry/NEFT workers)
- Synchronous processing (one disbursement at a time within batch)
- No priority queue (processed by creation date, newest first)
- Payment worker requires external trigger (channel send)