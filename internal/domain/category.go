package domain

import (
	"time"

	"github.com/google/uuid"
)

type Category struct {
	ID        uuid.UUID
	UserID    *uuid.UUID
	Name      string
	Type      TransactionType
	CreatedAt time.Time
	UpdatedAt time.Time
}

var DefaultExpenseCategories = []string{
	"Food & Beverage",
	"Transportation",
	"Shopping",
	"Bills",
	"Entertainment",
	"Health",
	"Education",
	"Travel",
	"Subscription",
	"Other",
}

var DefaultIncomeCategories = []string{
	"Salary",
	"Freelance",
	"Business",
	"Gift",
	"Investment",
	"Other",
}

func IsValidDefaultCategory(name string, t TransactionType) bool {
	var list []string
	switch t {
	case TypeExpense:
		list = DefaultExpenseCategories
	case TypeIncome:
		list = DefaultIncomeCategories
	default:
		return false
	}
	for _, n := range list {
		if n == name {
			return true
		}
	}
	return false
}
