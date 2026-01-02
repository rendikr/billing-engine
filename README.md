# Billing Engine - Loan Management System

A billing engine for managing 50-week loan schedules, payments, and delinquency tracking.

## Overview

This system provides:
- Loan schedule generation with 10% annual flat interest
- Outstanding balance tracking
- Delinquency detection (2+ weeks behind)
- Payment processing with validation
- Thread-safe operations

## Business Rules

1. **Loan Terms**: 50 weeks, 10% annual flat interest, Rp 5,000,000 principal → Rp 110,000 weekly payment
2. **Sequential Payments**: Must pay weeks in order (no skipping)
3. **Exact Amount**: Only exact weekly payment amount accepted
4. **Delinquency**: Borrower is 2+ weeks behind → delinquent
5. **Outstanding**: Total Amount - Sum of Payments

## Project Structure

```
billing-engine/
├── domain/
│   ├── loan.go          # Core business logic
│   ├── money.go         # Money value object
│   ├── errors.go        # Domain errors
│   └── loan_test.go     # Tests
├── service/
│   └── billing_service.go
├── main.go              # Demo
├── Makefile
└── README.md
```

## Quick Start

```bash
# Clone
git clone https://github.com/rendikr/billing-engine.git
cd billing-engine

# Install dependencies
make deps

# Run tests
make test

# Run demo
make run
```

## Usage Examples

### Create Loan
```go
billingService := service.NewBillingService()
principal := domain.NewMoney(5000000)
loan, _ := billingService.CreateLoan("loan-100", "borrower-123", principal)
loan.SetCurrentWeek(1)
```

### Make Payment
```go
// Specific week
billingService.MakePayment("loan-100", domain.NewMoney(110000), 1)

// Next due week
billingService.MakeNextPayment("loan-100", domain.NewMoney(110000))
```

### Check Status
```go
outstanding, _ := billingService.GetOutstanding("loan-100")
isDelinquent, _ := billingService.IsDelinquent("loan-100")
```

## API Reference

### BillingService
- `CreateLoan(loanID, borrowerID, principal) (*Loan, error)`
- `GetOutstanding(loanID) (Money, error)`
- `IsDelinquent(loanID) (bool, error)`
- `MakePayment(loanID, amount, weekNumber) error`
- `MakeNextPayment(loanID, amount) error`
- `GetSchedule(loanID) ([]ScheduleEntry, error)`
- `GetPaymentHistory(loanID) ([]Payment, error)`

### Loan
- `GetOutstanding() Money`
- `IsDelinquent() bool`
- `MakePayment(amount, weekNumber) error`
- `GetNextDueWeek() int`
- `IsClosed() bool`
- `SetCurrentWeek(week)`

## Error Handling

| Error | When |
|-------|------|
| `ErrInvalidPaymentAmount` | Wrong payment amount |
| `ErrNegativeAmount` | Negative amount |
| `ErrLoanFullyPaid` | Loan already closed |
| `ErrWeekAlreadyPaid` | Week already paid |
| `ErrInvalidWeekNumber` | Week out of range |
| `ErrPaymentOutOfSequence` | Skipping weeks |

## Testing

```bash
make test          # Run all tests
make coverage      # Generate coverage report
go test -v ./...   # Verbose output
```

**Coverage**: test cases covering:
- Loan creation, payments, delinquency
- Validation, edge cases, full lifecycle

## Extensibility

**Easy to Add**:
- Different loan products (change terms)
- Partial payments (modify validation)
- Grace periods (adjust threshold)
- Date-based tracking (replace week numbers)
- Database persistence (implement repository)
- Event notifications (emit domain events)

## Delinquency Logic

```
Delinquency = (CurrentWeek - LastPaidWeek) >= 2
```

**Examples**:
- Week 1, no payments: 1 - 0 = 1 → **NOT** delinquent
- Week 3, no payments: 3 - 0 = 3 → **DELINQUENT**
- Week 3, paid week 1: 3 - 1 = 2 → **DELINQUENT**
- Week 3, paid weeks 1-2: 3 - 2 = 1 → **NOT** delinquent

**Note**: `CurrentWeek` is manually set for demo. Production would calculate from dates.

## Assumptions

- Currency: IDR (Indonesian Rupiah)
- Interest: Flat 10% annually
- Payment timing: Week-based (manual tracking)
- No fees, partial payments, or overpayments
- Sequential payments only
- Week tracking: Manual (date-based in production)

## Edge Cases

- Duplicate payments (returns error)
- Out-of-sequence (rejected)
- Wrong amounts (validated)
- Negative amounts (rejected)
- Invalid week numbers (validated)
- Payments after closure (rejected)
- Concurrent access (thread-safe)

## Performance Considerations

- **Outstanding**: O(n) payments
- **Delinquency**: O(50) constant
- **Payment Lookup**: O(1)
- **Memory**: O(n) payments + O(50) schedule

**Production**: Cache outstanding, use DB indexes, pagination for history.

## Idempotency Considerations

**Current**:
- `CreateLoan`: Returns existing if duplicate ID
- `MakePayment`: Returns `ErrWeekAlreadyPaid` for duplicates

**Production Ready**:
- Idempotency keys for retries
- Database transactions for atomicity
- Event sourcing for audit trail

## Author

**Rendi** - Backend Engineer
Amartha Code Assignment (Example 1: Billing Engine)

## License

Code assignment for Amartha recruitment process.

---

**Tech Stack**: Go 1.24 | Clean Architecture | DDD | Thread-Safe | Comprehensive Tests
