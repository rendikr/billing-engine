package domain

import "errors"

var (
	// ErrInvalidPaymentAmount indicates the payment amount doesn't match the expected amount
	ErrInvalidPaymentAmount = errors.New("invalid payment amount: must match the weekly payment amount")

	// ErrNegativeAmount indicates a negative amount was provided
	ErrNegativeAmount = errors.New("amount cannot be negative")

	// ErrInvalidWeekNumber indicates an invalid week number was provided
	ErrInvalidWeekNumber = errors.New("invalid week number")
)
