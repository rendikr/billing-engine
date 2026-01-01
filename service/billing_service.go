package service

import (
	"fmt"
	"sync"

	"github.com/rendikr/billing-engine/domain"
	"github.com/shopspring/decimal"
)

type BillingService struct {
	loans map[string]*domain.Loan
	mu    sync.RWMutex
}

func NewBillingService() *BillingService {
	return &BillingService{
		loans: make(map[string]*domain.Loan),
	}
}

// CreateLoan creates a new loan with specific terms
// Terms: 50 weeks, 10% annual interest
func (s *BillingService) CreateLoan(loanID, borrowerID string, principal domain.Money) (*domain.Loan, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if loan already exists
	if _, exists := s.loans[loanID]; exists {
		return nil, fmt.Errorf("loan with ID %s already exists", loanID)
	}

	// Create loan with the terms
	annualInterestRate := decimal.NewFromFloat(0.10) // 10% per annum
	loan := domain.NewLoan(loanID, borrowerID, principal, annualInterestRate)

	s.loans[loanID] = loan

	return loan, nil
}

// GetLoan retrieves a loan by ID
func (s *BillingService) GetLoan(loanID string) (*domain.Loan, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	loan, exists := s.loans[loanID]
	if !exists {
		return nil, fmt.Errorf("loan with ID %s not found", loanID)
	}

	return loan, nil
}

// GetOutstanding returns the outstanding amount for a loan
func (s *BillingService) GetOutstanding(loanID string) (domain.Money, error) {
	loan, err := s.GetLoan(loanID)
	if err != nil {
		return domain.Money{}, err
	}

	return loan.GetOutstanding(), nil
}

// IsDelinquent checks if a borrower is delinquent on a loan
func (s *BillingService) IsDelinquent(loanID string) (bool, error) {
	loan, err := s.GetLoan(loanID)
	if err != nil {
		return false, err
	}

	return loan.IsDelinquent(), nil
}

// MakePayment processes a payment on a loan
func (s *BillingService) MakePayment(loanID string, amount domain.Money, weekNumber int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	loan, exists := s.loans[loanID]
	if !exists {
		return fmt.Errorf("loan with ID %s not found", loanID)
	}

	return loan.MakePayment(amount, weekNumber)
}

// MakeNextPayment process a payment for the next due week
func (s *BillingService) MakeNextPayment(loanID string, amount domain.Money) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	loan, exists := s.loans[loanID]
	if !exists {
		return fmt.Errorf("loan with ID %s not found", loanID)
	}

	nextWeek := loan.GetNextDueWeek()
	if nextWeek == 0 {
		return domain.ErrLoanFullyPaid
	}

	return loan.MakePayment(amount, nextWeek)
}

// GetSchedule returns the payment schedule for a loan
func (s *BillingService) GetSchedule(loanID string) ([]domain.ScheduleEntry, error) {
	loan, err := s.GetLoan(loanID)
	if err != nil {
		return nil, err
	}

	return loan.GetSchedule(), nil
}

// GetPaymentHistory returns the payment history for a loan
func (s *BillingService) GetPaymentHistory(loanID string) ([]domain.Payment, error) {
	loan, err := s.GetLoan(loanID)
	if err != nil {
		return nil, err
	}

	return loan.GetPaymentHistory(), nil
}
