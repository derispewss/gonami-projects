package application

import (
	"context"
	"fmt"
	"strings"

	"github.com/derispewss/gonami-projects/internal/domain"
	"github.com/derispewss/gonami-projects/internal/repository"
)

type ManageCategory struct {
	users *repository.UserRepo
	cats  *repository.CategoryRepo
}

func NewManageCategory(u *repository.UserRepo, c *repository.CategoryRepo) *ManageCategory {
	return &ManageCategory{users: u, cats: c}
}

type CategoryListResult struct {
	Custom  []string
	Default []string
}

func (uc *ManageCategory) Add(ctx context.Context, jid, rawName string) (*domain.Category, error) {
	name := cleanCategoryName(rawName)
	if name == "" {
		return nil, domain.ErrInvalidInput
	}

	txType := domain.TypeExpense
	for _, w := range strings.Fields(strings.ToLower(rawName)) {
		if w == "pemasukan" || w == "income" {
			txType = domain.TypeIncome
		}
	}

	user, err := uc.users.GetOrCreateByJID(ctx, jid, "")
	if err != nil {
		return nil, err
	}
	return uc.cats.CreateCustom(ctx, user.ID, name, txType)
}

func (uc *ManageCategory) List(ctx context.Context, jid string) (*CategoryListResult, error) {
	user, err := uc.users.GetOrCreateByJID(ctx, jid, "")
	if err != nil {
		return nil, err
	}

	cats, err := uc.cats.ListAvailable(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	res := &CategoryListResult{}
	for _, c := range cats {
		if c.UserID != nil {
			res.Custom = append(res.Custom, c.Name)
		} else {
			res.Default = append(res.Default, c.Name)
		}
	}
	return res, nil
}

func cleanCategoryName(raw string) string {
	var words []string
	for _, w := range strings.Fields(raw) {
		low := strings.ToLower(w)
		switch low {
		case "tambah", "buat", "bikin", "kategori", "baru",
			"pemasukan", "income", "pengeluaran", "expense",
			"dong", "deh", "aja", "ya":
			continue
		}
		words = append(words, strings.ToLower(w))
	}
	if len(words) == 0 || len(words) > 3 {
		return ""
	}
	name := strings.Join(words, " ")
	if len(name) > 100 {
		return fmt.Sprint(name[:100])
	}
	return name
}
