package domain

import (
	"testing"

	"github.com/shopspring/decimal"
)

func TestNewLoan(t *testing.T) {
	principal := NewMoney(5000000)
	interestRate := decimal.NewFromFloat(0.10)

	loan := NewLoan("loan-1", "borrower-1", principal, interestRate)

	// Test basic loan properties
	if loan.ID != "loan-1" {
		t.Errorf("Expected loan ID 'loan-1', got '%s'", loan.ID)
	}

	if loan.BorrowerID != "borrower-1" {
		t.Errorf("Expected borrower ID 'borrower-1', got '%s'", loan.BorrowerID)
	}

	// Test interest calculation: 5,000,000 * 0.10 = 500,000
	expectedInterest := NewMoney(500000)
	expectedTotal := principal.Add(expectedInterest) // 5,500,000

	if !loan.TotalAmount.Equals(expectedTotal) {
		t.Errorf("Expected total amount %s, got %s", expectedTotal, loan.TotalAmount)
	}

	// Test weekly payment: 5,500,000 / 50 = 110,000
	expectedWeeklyPayment := NewMoney(110000)
	if !loan.WeeklyPayment.Equals(expectedWeeklyPayment) {
		t.Errorf("Expected weekly payment %s, got %s", expectedWeeklyPayment, loan.WeeklyPayment)
	}

	// Test schedule generation
	if len(loan.Schedule) != LoanDurationWeeks {
		t.Errorf("Expected %d schedule entries, got %d", LoanDurationWeeks, len(loan.Schedule))
	}

	// Verify all schedule entries
	for i, entry := range loan.Schedule {
		if entry.WeekNumber != i+1 {
			t.Errorf("Expected week number %d, got %d", i+1, entry.WeekNumber)
		}
		if !entry.Amount.Equals(expectedWeeklyPayment) {
			t.Errorf("Expected amount %s for week %d, got %s", expectedWeeklyPayment, i+1, entry.Amount)
		}
		if entry.IsPaid {
			t.Errorf("Expected week %d to be unpaid initially", i+1)
		}
	}
}

func TestGetOutstanding(t *testing.T) {
	loan := createTestLoan()

	// Initially, outstanding should equal total amount
	expected := NewMoney(5500000)
	if !loan.GetOutstanding().Equals(expected) {
		t.Errorf("Expected initial outstanding %s, got %s", expected, loan.GetOutstanding())
	}

	// After one payment
	loan.MakePayment(NewMoney(110000), 1)
	expected = NewMoney(5390000) // 5,500,000 - 110,000
	if !loan.GetOutstanding().Equals(expected) {
		t.Errorf("Expected outstanding %s after 1 payment, got %s", expected, loan.GetOutstanding())
	}

	// After two payments
	loan.MakePayment(NewMoney(110000), 2)
	expected = NewMoney(5280000) // 5,500,000 - 220,000
	if !loan.GetOutstanding().Equals(expected) {
		t.Errorf("Expected outstanding %s after 2 payments, got %s", expected, loan.GetOutstanding())
	}
}

func TestIsDelinquent(t *testing.T) {
	tests := []struct {
		name               string
		paidWeeks          []int
		currentWeek        int
		expectedDelinquent bool
		description        string
	}{
		{
			name:               "Week 1 - no payments yet",
			paidWeeks:          []int{},
			currentWeek:        1,
			expectedDelinquent: false, // Week 1, last paid = 0, behind = 1
			description:        "Should NOT be delinquent in week 1 with no payments",
		},
		{
			name:               "Week 3 - no payments",
			paidWeeks:          []int{},
			currentWeek:        3,
			expectedDelinquent: true, // Week 3, last paid = 0, behind = 3 >= 2
			description:        "Should be delinquent in week 3 with no payments",
		},
		{
			name:               "Week 3 - paid week 1 only",
			paidWeeks:          []int{1},
			currentWeek:        3,
			expectedDelinquent: true, // Week 3, last paid = 1, behind = 2
			description:        "Should be delinquent when 2 weeks behind",
		},
		{
			name:               "Week 3 - paid weeks 1 and 2",
			paidWeeks:          []int{1, 2},
			currentWeek:        3,
			expectedDelinquent: false, // Week 3, last paid = 2, behind = 1
			description:        "Should NOT be delinquent when only 1 week behind",
		},
		{
			name:               "Week 5 - paid weeks 1,2,3",
			paidWeeks:          []int{1, 2, 3},
			currentWeek:        5,
			expectedDelinquent: true, // Week 5, last paid = 3, behind = 2
			description:        "Should be delinquent when 2 weeks behind",
		},
		{
			name:               "All paid up to current week",
			paidWeeks:          makeRange(1, 10),
			currentWeek:        10,
			expectedDelinquent: false, // Week 10, last paid = 10, behind = 0
			description:        "Should not be delinquent when current",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loan := createTestLoan()
			loan.SetCurrentWeek(tt.currentWeek)

			// Make payments for specified weeks
			for _, week := range tt.paidWeeks {
				if err := loan.MakePayment(NewMoney(110000), week); err != nil {
					t.Fatalf("Failed to make payment for week %d: %v", week, err)
				}
			}

			result := loan.IsDelinquent()
			if result != tt.expectedDelinquent {
				t.Errorf("%s: Expected delinquent=%v, got %v (current week=%d, last paid=%d)",
					tt.description, tt.expectedDelinquent, result, loan.CurrentWeek, findLastPaidWeek(loan))
			}
		})
	}
}

// Helper to find last paid week
func findLastPaidWeek(loan *Loan) int {
	lastPaid := 0
	for _, entry := range loan.Schedule {
		if entry.IsPaid && entry.WeekNumber > lastPaid {
			lastPaid = entry.WeekNumber
		}
	}
	return lastPaid
}

func TestMakePayment_Success(t *testing.T) {
	loan := createTestLoan()

	// Make first payment
	err := loan.MakePayment(NewMoney(110000), 1)
	if err != nil {
		t.Errorf("Expected successful payment, got error: %v", err)
	}

	// Verify payment was recorded
	if len(loan.Payments) != 1 {
		t.Errorf("Expected 1 payment recorded, got %d", len(loan.Payments))
	}

	// Verify schedule was updated
	if !loan.Schedule[0].IsPaid {
		t.Error("Expected week 1 to be marked as paid")
	}

	// Make second payment
	err = loan.MakePayment(NewMoney(110000), 2)
	if err != nil {
		t.Errorf("Expected successful payment, got error: %v", err)
	}

	if len(loan.Payments) != 2 {
		t.Errorf("Expected 2 payments recorded, got %d", len(loan.Payments))
	}
}

func TestMakePayment_InvalidAmount(t *testing.T) {
	loan := createTestLoan()

	// Try to pay wrong amount
	err := loan.MakePayment(NewMoney(100000), 1)
	if err != ErrInvalidPaymentAmount {
		t.Errorf("Expected ErrInvalidPaymentAmount, got %v", err)
	}

	// Try to pay more than required
	err = loan.MakePayment(NewMoney(120000), 1)
	if err != ErrInvalidPaymentAmount {
		t.Errorf("Expected ErrInvalidPaymentAmount, got %v", err)
	}

	// Try negative amount
	err = loan.MakePayment(NewMoney(-110000), 1)
	if err != ErrNegativeAmount {
		t.Errorf("Expected ErrNegativeAmount, got %v", err)
	}
}

func TestMakePayment_InvalidWeekNumber(t *testing.T) {
	loan := createTestLoan()

	// Try week 0
	err := loan.MakePayment(NewMoney(110000), 0)
	if err != ErrInvalidWeekNumber {
		t.Errorf("Expected ErrInvalidWeekNumber for week 0, got %v", err)
	}

	// Try week 51
	err = loan.MakePayment(NewMoney(110000), 51)
	if err != ErrInvalidWeekNumber {
		t.Errorf("Expected ErrInvalidWeekNumber for week 51, got %v", err)
	}
}

func TestMakePayment_AlreadyPaid(t *testing.T) {
	loan := createTestLoan()

	// Pay week 1
	loan.MakePayment(NewMoney(110000), 1)

	// Try to pay week 1 again
	err := loan.MakePayment(NewMoney(110000), 1)
	if err != ErrWeekAlreadyPaid {
		t.Errorf("Expected ErrWeekAlreadyPaid, got %v", err)
	}
}

func TestMakePayment_OutOfSequence(t *testing.T) {
	loan := createTestLoan()

	// Try to pay week 2 before week 1
	err := loan.MakePayment(NewMoney(110000), 2)
	if err != ErrPaymentOutOfSequence {
		t.Errorf("Expected ErrPaymentOutOfSequence, got %v", err)
	}

	// Pay week 1
	loan.MakePayment(NewMoney(110000), 1)

	// Try to skip week 2 and pay week 3
	err = loan.MakePayment(NewMoney(110000), 3)
	if err != ErrPaymentOutOfSequence {
		t.Errorf("Expected ErrPaymentOutOfSequence, got %v", err)
	}
}

func TestMakePayment_FullyPaid(t *testing.T) {
	loan := createTestLoan()

	// Pay all 50 weeks
	for i := 1; i <= 50; i++ {
		err := loan.MakePayment(NewMoney(110000), i)
		if err != nil {
			t.Fatalf("Failed to make payment for week %d: %v", i, err)
		}
	}

	// Verify loan is closed
	if !loan.IsClosed() {
		t.Error("Expected loan to be closed after all payments")
	}

	// Verify outstanding is zero
	if !loan.GetOutstanding().IsZero() {
		t.Errorf("Expected outstanding to be zero, got %s", loan.GetOutstanding())
	}

	// Try to make another payment
	err := loan.MakePayment(NewMoney(110000), 1)
	if err != ErrLoanFullyPaid {
		t.Errorf("Expected ErrLoanFullyPaid, got %v", err)
	}
}

func TestGetNextDueWeek(t *testing.T) {
	loan := createTestLoan()

	// Initially, week 1 should be due
	if loan.GetNextDueWeek() != 1 {
		t.Errorf("Expected next due week to be 1, got %d", loan.GetNextDueWeek())
	}

	// After paying week 1, week 2 should be due
	loan.MakePayment(NewMoney(110000), 1)
	if loan.GetNextDueWeek() != 2 {
		t.Errorf("Expected next due week to be 2, got %d", loan.GetNextDueWeek())
	}

	// After paying all weeks, should return 0
	for i := 2; i <= 50; i++ {
		loan.MakePayment(NewMoney(110000), i)
	}
	if loan.GetNextDueWeek() != 0 {
		t.Errorf("Expected next due week to be 0 (all paid), got %d", loan.GetNextDueWeek())
	}
}

func TestDelinquencyScenarios(t *testing.T) {
	t.Run("New loan in week 1 is not delinquent", func(t *testing.T) {
		loan := createTestLoan()
		loan.SetCurrentWeek(1)

		// New loan in week 1 should NOT be delinquent
		if loan.IsDelinquent() {
			t.Error("Expected new loan in week 1 to NOT be delinquent")
		}
	})

	t.Run("Week 3 with no payments is delinquent", func(t *testing.T) {
		loan := createTestLoan()
		loan.SetCurrentWeek(3)

		// Week 3, no payments (behind by 3) - should be delinquent
		if !loan.IsDelinquent() {
			t.Error("Expected to be delinquent in week 3 with no payments")
		}
	})

	t.Run("Paid week 1, now in week 3 - delinquent", func(t *testing.T) {
		loan := createTestLoan()
		loan.MakePayment(NewMoney(110000), 1)
		loan.SetCurrentWeek(3)

		// Week 3, last paid week 1 (behind by 2) - should be delinquent
		if !loan.IsDelinquent() {
			t.Error("Expected to be delinquent when 2 weeks behind")
		}
	})

	t.Run("Paid weeks 1-2, now in week 3 - not delinquent", func(t *testing.T) {
		loan := createTestLoan()
		loan.MakePayment(NewMoney(110000), 1)
		loan.MakePayment(NewMoney(110000), 2)
		loan.SetCurrentWeek(3)

		// Week 3, last paid week 2 (behind by 1) - should NOT be delinquent
		if loan.IsDelinquent() {
			t.Error("Expected to NOT be delinquent when only 1 week behind")
		}
	})

	t.Run("Customer catches up from delinquency", func(t *testing.T) {
		loan := createTestLoan()

		// Start at week 3, paid only week 1 (delinquent)
		loan.MakePayment(NewMoney(110000), 1)
		loan.SetCurrentWeek(3)
		if !loan.IsDelinquent() {
			t.Error("Expected to be delinquent (2 weeks behind)")
		}

		// Pay week 2 to catch up
		loan.MakePayment(NewMoney(110000), 2)
		// Still in week 3, now only 1 week behind
		if loan.IsDelinquent() {
			t.Error("Expected to NOT be delinquent after catching up (1 week behind)")
		}
	})
}

// Helper functions
func createTestLoan() *Loan {
	principal := NewMoney(5000000)
	interestRate := decimal.NewFromFloat(0.10)
	return NewLoan("test-loan", "test-borrower", principal, interestRate)
}

func makeRange(min, max int) []int {
	result := make([]int, max-min+1)
	for i := range result {
		result[i] = min + i
	}
	return result
}
