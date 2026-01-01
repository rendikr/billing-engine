package domain

import "errors"

var (
	// ErrInvalidPaymentAmount indicates the payment amount doesn't match the expected amount
	ErrInvalidPaymentAmount = errors.New("invalid payment amount: must match the weekly payment amount")

	// ErrNegativeAmount indicates a negative amount was provided
	ErrNegativeAmount = errors.New("amount cannot be negative")

	// ErrInvalidWeekNumber indicates an invalid week number was provided
	ErrInvalidWeekNumber = errors.New("invalid week number")

	// ErrLoanFullyPaid indicates attempting to pay an already fully paid loan
	ErrLoanFullyPaid = errors.New("loan is already fully paid")

	// ErrWeekAlreadyPaid indicates attempting to pay for a week that's already paid
	ErrWeekAlreadyPaid = errors.New("this week has already been paid")

	// ErrPaymentOutOfSequence indicates attempting to pay a week out of sequence
	ErrPaymentOutOfSequence = errors.New("payments must be made in sequence (cannot skip unpaid weeks)")
)
