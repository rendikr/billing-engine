package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

const (
	// LoanDurationWeeks is the fixed loan duration
	LoanDurationWeeks = 50

	// DelinquencyThreshold is the number of consecutive missed payments to be delinquent
	DelinquencyThreshold = 2
)

type ScheduleEntry struct {
	WeekNumber int
	Amount     Money
	IsPaid     bool
}

type Payment struct {
	WeekNumber int
	Amount     Money
	PaidAt     time.Time
}

type Loan struct {
	ID            string
	BorrowerID    string
	Principal     Money
	InterestRate  decimal.Decimal // Annual interest rate (e.g., 0.10 for 10%)
	TotalAmount   Money           // Principal + Interest
	WeeklyPayment Money
	Schedule      []ScheduleEntry
	Payments      []Payment
	CurrentWeek   int
}

// NewLoan creates a new loan with the given parameters
func NewLoan(id, borrowerID string, principal Money, annualInterestRate decimal.Decimal) *Loan {
	// Calculate total interest: principal * rate (flat interest, not compound)
	interest := principal.Multiply(annualInterestRate)
	totalAmount := principal.Add(interest)

	// Calculate weekly payment: total amount / number of weeks
	weeklyPayment := totalAmount.Multiply(decimal.NewFromInt(1).Div(decimal.NewFromInt(LoanDurationWeeks)))

	// Generate payment schedule
	schedule := make([]ScheduleEntry, LoanDurationWeeks)
	for i := range LoanDurationWeeks {
		schedule[i] = ScheduleEntry{
			WeekNumber: i + 1,
			Amount:     weeklyPayment,
			IsPaid:     false,
		}
	}

	return &Loan{
		ID:            id,
		BorrowerID:    borrowerID,
		Principal:     principal,
		InterestRate:  annualInterestRate,
		TotalAmount:   totalAmount,
		WeeklyPayment: weeklyPayment,
		Schedule:      schedule,
		Payments:      make([]Payment, 0),
		CurrentWeek:   1,
	}
}

// GetOutstanding returns the current outstanding amount on the loan
// Outstanding = Total Amount - Sum of all successful payments
func (l *Loan) GetOutstanding() Money {
	totalPaid := NewMoney(0)
	for _, payment := range l.Payments {
		totalPaid = totalPaid.Add(payment.Amount)
	}
	return l.TotalAmount.Subtract(totalPaid)
}

// IsDelinquent checks if the borrower is delinquent
// A borrower is delinquent if they are behind by 2 or more weeks
// (current week - last paid week >= 2)
func (l *Loan) IsDelinquent() bool {
	// Find the last paid week
	lastPaidWeek := 0
	for _, entry := range l.Schedule {
		if entry.IsPaid && entry.WeekNumber > lastPaidWeek {
			lastPaidWeek = entry.WeekNumber
		}
	}

	// Calculate how many weeks behind
	weeksBehind := l.CurrentWeek - lastPaidWeek

	// Delinquent if 2 or more weeks behind
	return weeksBehind >= DelinquencyThreshold
}

// SetCurrentWeek sets the current week (for testing/simulation)
// In production, this would be calculated from dates
func (l *Loan) SetCurrentWeek(week int) {
	if week >= 1 && week <= LoanDurationWeeks {
		l.CurrentWeek = week
	}
}

// MakePayment records a payment for a specific week
// Validation:
// - Amount is correct (must match weekly payment)
// - Week is valid
// - Week hasn't been paid already
// - Payment is in sequence
func (l *Loan) MakePayment(amount Money, weekNumber int) error {
	// Validate amount is not negative
	if amount.IsNegative() {
		return ErrNegativeAmount
	}

	// Validate amount matches weekly payment
	if !amount.Equals(l.WeeklyPayment) {
		return ErrInvalidPaymentAmount
	}

	// Check if loan is already fully paid
	if l.GetOutstanding().IsZero() {
		return ErrLoanFullyPaid
	}

	// Validate week number
	if weekNumber < 1 || weekNumber > LoanDurationWeeks {
		return ErrInvalidWeekNumber
	}

	// Check if this specific week is already paid
	scheduleIndex := weekNumber - 1
	if l.Schedule[scheduleIndex].IsPaid {
		return ErrWeekAlreadyPaid
	}

	// Ensure payments are made in sequence
	// Find the first unpaid week
	firstUnpaidWeek := l.findFirstUnpaidWeek()
	if weekNumber != firstUnpaidWeek {
		return ErrPaymentOutOfSequence
	}

	// Record the payment
	payment := Payment{
		WeekNumber: weekNumber,
		Amount:     amount,
		PaidAt:     time.Now(),
	}
	l.Payments = append(l.Payments, payment)

	// Update schedule
	l.Schedule[scheduleIndex].IsPaid = true

	return nil
}

// findFirstUnpaidWeek returns the week number of the first unpaid week
// Returns 0 if all weeks are paid
func (l *Loan) findFirstUnpaidWeek() int {
	for _, entry := range l.Schedule {
		if !entry.IsPaid {
			return entry.WeekNumber
		}
	}
	return 0
}

// GetSchedule returns a copy of the payment schedule
func (l *Loan) GetSchedule() []ScheduleEntry {
	scheduleCopy := make([]ScheduleEntry, len(l.Schedule))
	copy(scheduleCopy, l.Schedule)
	return scheduleCopy
}

// GetPaymentHistory returns a copy of the payment history
func (l *Loan) GetPaymentHistory() []Payment {
	paymentsCopy := make([]Payment, len(l.Payments))
	copy(paymentsCopy, l.Payments)
	return paymentsCopy
}

// GetNextDueWeek returns the next week number that needs to be paid
// Returns 0 if all weeks are paid
func (l *Loan) GetNextDueWeek() int {
	return l.findFirstUnpaidWeek()
}

func (l *Loan) IsClosed() bool {
	return l.GetOutstanding().IsZero()
}
