package domain

import (
	"fmt"

	"github.com/shopspring/decimal"
)

type Money struct {
	amount decimal.Decimal
}

// NewMoney creates a new Money instance from an int64 value
func NewMoney(amount int64) Money {
	return Money{
		amount: decimal.NewFromInt(amount),
	}
}

// NewMoneyFromDecimal creates Money from a decimal value
func NewMoneyFromDecimal(amount decimal.Decimal) Money {
	return Money{amount: amount}
}

func (m Money) Amount() decimal.Decimal {
	return m.amount
}

func (m Money) IsZero() bool {
	return m.amount.IsZero()
}

func (m Money) IsNegative() bool {
	return m.amount.IsNegative()
}

func (m Money) Equals(other Money) bool {
	return m.amount.Equal(other.amount)
}

func (m Money) Add(other Money) Money {
	return Money{amount: m.amount.Add(other.amount)}
}

func (m Money) Subtract(other Money) Money {
	return Money{amount: m.amount.Sub(other.amount)}
}

func (m Money) Multiply(multiplier decimal.Decimal) Money {
	return Money{amount: m.amount.Mul(multiplier)}
}

func (m Money) GreaterThan(other Money) bool {
	return m.amount.GreaterThan(other.amount)
}

func (m Money) LessThan(other Money) bool {
	return m.amount.LessThan(other.amount)
}

func (m Money) String() string {
	return fmt.Sprintf("IDR %s", m.amount.StringFixed(0))
}

func (m Money) Int64() int64 {
	return m.amount.IntPart()
}
