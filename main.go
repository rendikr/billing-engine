package main

import (
	"fmt"

	"github.com/rendikr/billing-engine/domain"
	"github.com/rendikr/billing-engine/service"
)

func main() {
	fmt.Println("=== Billing Engine Demo ===")
	fmt.Println()

	// Create billing service
	billingService := service.NewBillingService()

	// Create a loan for borrower
	principal := domain.NewMoney(5000000)
	loan, err := billingService.CreateLoan("loan-100", "borrower-123", principal)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Loan Created: %s\n", loan.ID)
	fmt.Printf("Borrower: %s\n", loan.BorrowerID)
	fmt.Printf("Principal: %s\n", principal)
	fmt.Printf("Total Amount (with 10%% interest): %s\n", loan.TotalAmount)
	fmt.Printf("Weekly Payment: %s\n", loan.WeeklyPayment)
	fmt.Printf("Duration: %d weeks\n\n", domain.LoanDurationWeeks)

	// Display payment schedule (first 5 weeks as sample)
	fmt.Println("Payment Schedule (first 5 weeks):")
	schedule := loan.GetSchedule()
	for i := 0; i < 5 && i < len(schedule); i++ {
		entry := schedule[i]
		fmt.Printf("  W%d: %s\n", entry.WeekNumber, entry.Amount)
	}
	fmt.Println("  ...")
	fmt.Printf("  (Total %d weeks)\n\n", len(schedule))

	// Check initial status (Week 1)
	fmt.Println("=== Initial Status (Week 1) ===")
	loan.SetCurrentWeek(1)
	outstanding, _ := billingService.GetOutstanding(loan.ID)
	isDelinquent, _ := billingService.IsDelinquent(loan.ID)
	fmt.Printf("Current Week: %d\n", loan.CurrentWeek)
	fmt.Printf("Outstanding: %s\n", outstanding)
	fmt.Printf("Is Delinquent: %v (current week: %d)\n\n", isDelinquent, loan.CurrentWeek)

	// Scenario 1: Customer makes regular payments
	fmt.Println("=== Scenario 1: Regular Payments ===")

	// Week 1 payment
	fmt.Println("Making payment for Week 1...")
	err = billingService.MakePayment(loan.ID, domain.NewMoney(110000), 1)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Println("✓ Payment successful")
	}

	outstanding, _ = billingService.GetOutstanding(loan.ID)
	fmt.Printf("Outstanding after Week 1: %s\n", outstanding)

	// Week 2 payment
	fmt.Println("\nMaking payment for Week 2...")
	err = billingService.MakeNextPayment(loan.ID, domain.NewMoney(110000))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Println("✓ Payment successful")
	}

	outstanding, _ = billingService.GetOutstanding(loan.ID)
	isDelinquent, _ = billingService.IsDelinquent(loan.ID)
	fmt.Printf("Outstanding after Week 2: %s\n", outstanding)
	fmt.Printf("Is Delinquent: %v\n\n", isDelinquent)

	// Scenario 2: Customer tries to pay wrong amount
	fmt.Println("=== Scenario 2: Invalid Payment Amount ===")
	fmt.Println("Attempting to pay Rp 100,000 (incorrect amount)...")
	err = billingService.MakeNextPayment(loan.ID, domain.NewMoney(100000))
	if err != nil {
		fmt.Printf("✗ Error: %v\n\n", err)
	}

	// Scenario 3: Customer tries to skip weeks
	fmt.Println("=== Scenario 3: Out of Sequence Payment ===")
	fmt.Println("Attempting to pay Week 5 (skipping Weeks 3 and 4)...")
	err = billingService.MakePayment(loan.ID, domain.NewMoney(110000), 5)
	if err != nil {
		fmt.Printf("✗ Error: %v\n\n", err)
	}

	// Scenario 4: Customer continues paying
	fmt.Println("=== Scenario 4: Continuing Regular Payments ===")
	for week := 3; week <= 5; week++ {
		fmt.Printf("Making payment for Week %d...\n", week)
		err = billingService.MakeNextPayment(loan.ID, domain.NewMoney(110000))
		if err != nil {
			fmt.Printf("✗ Error: %v\n", err)
		} else {
			fmt.Println("✓ Payment successful")
		}
	}

	outstanding, _ = billingService.GetOutstanding(loan.ID)
	isDelinquent, _ = billingService.IsDelinquent(loan.ID)
	nextDue := loan.GetNextDueWeek()
	fmt.Printf("\nCurrent Status:\n")
	fmt.Printf("  Outstanding: %s\n", outstanding)
	fmt.Printf("  Is Delinquent: %v\n", isDelinquent)
	fmt.Printf("  Next Due Week: %d\n", nextDue)
	fmt.Printf("  Payments Made: %d / %d\n\n", len(loan.GetPaymentHistory()), domain.LoanDurationWeeks)

	// Scenario 5: Simulate delinquency (create new loan)
	fmt.Println("=== Scenario 5: Delinquency Example ===")
	loan2, _ := billingService.CreateLoan("loan-101", "borrower-456", principal)

	fmt.Println("Week 1: New loan created, no payments made yet...")
	loan2.SetCurrentWeek(1)
	isDelinquent2, _ := billingService.IsDelinquent(loan2.ID)
	fmt.Printf("Is Delinquent: %v (current week: %d, last paid: 0, behind by: 1)\n\n", isDelinquent2, loan2.CurrentWeek)

	// Simulate time passing to week 3 without payment
	loan2.SetCurrentWeek(3)
	fmt.Println("Week 3: Still no payments made...")
	isDelinquent2, _ = billingService.IsDelinquent(loan2.ID)
	fmt.Printf("Is Delinquent: %v (current week: %d, last paid: 0, behind by: 3)\n\n", isDelinquent2, loan2.CurrentWeek)

	// Pay week 1 only
	billingService.MakePayment(loan2.ID, domain.NewMoney(110000), 1)
	fmt.Println("Paid Week 1, but still in Week 3...")
	isDelinquent2, _ = billingService.IsDelinquent(loan2.ID)
	fmt.Printf("Is Delinquent: %v (current week: %d, last paid: 1, behind by: 2) ← Still DELINQUENT!\n\n", isDelinquent2, loan2.CurrentWeek)

	// Catch up by paying week 2
	billingService.MakePayment(loan2.ID, domain.NewMoney(110000), 2)
	fmt.Println("Caught up! Paid Week 2, still in Week 3...")
	isDelinquent2, _ = billingService.IsDelinquent(loan2.ID)
	fmt.Printf("Is Delinquent: %v (current week: %d, last paid: 2, behind by: 1) ← No longer delinquent!\n\n", isDelinquent2, loan2.CurrentWeek)

	// Scenario 6: Payment History
	fmt.Println("=== Scenario 6: Payment History ===")
	history, _ := billingService.GetPaymentHistory(loan.ID)
	fmt.Printf("Total payments made: %d\n", len(history))
	fmt.Println("Recent payments:")
	for i, payment := range history {
		if i >= 5 {
			fmt.Println("  ...")
			break
		}
		fmt.Printf("  Week %d: %s (paid at %s)\n",
			payment.WeekNumber,
			payment.Amount,
			payment.PaidAt.Format("2006-01-02 15:04:05"))
	}

	fmt.Println("\n=== Demo Complete ===")
}
