package application

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/derispewss/gonami-projects/internal/domain"
	"github.com/derispewss/gonami-projects/internal/format"
	"github.com/derispewss/gonami-projects/internal/parser"
	"github.com/derispewss/gonami-projects/internal/repository"

	"github.com/go-pdf/fpdf"
)

type ExportUC struct {
	users *repository.UserRepo
	txs   *repository.TransactionRepo
}

func NewExportUC(u *repository.UserRepo, t *repository.TransactionRepo) *ExportUC {
	return &ExportUC{users: u, txs: t}
}

type ExportResult struct {
	Filename string
	MimeType string
	Data     []byte
	From     time.Time
	To       time.Time
	Count    int
}

func (uc *ExportUC) Run(ctx context.Context, jid string, period parser.Period, pdf bool) (*ExportResult, error) {
	user, err := uc.users.GetOrCreateByJID(ctx, jid, "")
	if err != nil {
		return nil, err
	}

	loc, _ := time.LoadLocation("Asia/Jakarta")
	now := time.Now().In(loc)
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)

	var from, to time.Time
	switch period {
	case parser.PeriodDaily:
		from = today
		to = today.AddDate(0, 0, 1)
	case parser.PeriodWeekly:
		wd := int(today.Weekday())
		if wd == 0 {
			wd = 7
		}
		from = today.AddDate(0, 0, -(wd - 1))
		to = from.AddDate(0, 0, 7)
	default:
		from = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc)
		to = from.AddDate(0, 1, 0)
	}

	txs, err := uc.txs.ListByRange(ctx, user.ID, from, to)
	if err != nil {
		return nil, err
	}
	if len(txs) == 0 {
		return nil, domain.ErrNotFound
	}

	var label strings.Builder
	label.WriteString(periodLabel(period))
	label.WriteString(" ")
	label.WriteString(format.MonthYearID(from))

	if pdf {
		data, err := buildPDF(txs, label.String(), from, to)
		if err != nil {
			return nil, err
		}
		return &ExportResult{
			Filename: fmt.Sprintf("gonami-%s.pdf", fileSlug(label.String())),
			MimeType: "application/pdf",
			Data:     data,
			From:     from, To: to, Count: len(txs),
		}, nil
	}

	return &ExportResult{
		Filename: fmt.Sprintf("gonami-%s.txt", fileSlug(label.String())),
		MimeType: "text/plain",
		Data:     buildText(txs, label.String()),
		From:     from, To: to, Count: len(txs),
	}, nil
}

func buildText(txs []*domain.Transaction, label string) []byte {
	var b strings.Builder
	b.WriteString("GONAMI — LAPORAN TRANSAKSI\n")
	b.WriteString(label + "\n")
	b.WriteString(strings.Repeat("=", 40) + "\n\n")

	var income, expense int64
	for _, tx := range txs {
		sign := "-"
		if tx.Type == domain.TypeIncome {
			sign = "+"
			income += tx.Amount
		} else {
			expense += tx.Amount
		}
		desc := format.Truncate(tx.Description, 24)
		if desc == "" {
			desc = "(tanpa deskripsi)"
		}
		b.WriteString(fmt.Sprintf("%s %s\n", tx.TransactionDate.Format("02 Jan"), desc))
		b.WriteString(fmt.Sprintf("      %s%s | %s\n\n", sign, format.Rupiah(tx.Amount), tx.CategoryName))
	}

	b.WriteString(strings.Repeat("-", 40) + "\n")
	b.WriteString(fmt.Sprintf("Masuk : %s\n", format.Rupiah(income)))
	b.WriteString(fmt.Sprintf("Keluar: %s\n", format.Rupiah(expense)))
	b.WriteString(fmt.Sprintf("Net   : %s\n", format.Rupiah(income-expense)))
	return []byte(b.String())
}

func buildPDF(txs []*domain.Transaction, label string, from, to time.Time) ([]byte, error) {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(15, 15, 15)
	pdf.AddPage()

	pdf.SetFont("Helvetica", "B", 16)
	pdf.Cell(0, 10, "Gonami - Laporan Transaksi")
	pdf.Ln(8)
	pdf.SetFont("Helvetica", "", 11)
	pdf.Cell(0, 6, label+"  ("+from.Format("02/01/2006")+" - "+to.AddDate(0, 0, -1).Format("02/01/2006")+")")
	pdf.Ln(10)

	pdf.SetFont("Helvetica", "B", 9)
	pdf.SetFillColor(235, 235, 235)
	pdf.CellFormat(22, 7, "Tanggal", "1", 0, "", true, 0, "")
	pdf.CellFormat(78, 7, "Deskripsi", "1", 0, "", true, 0, "")
	pdf.CellFormat(30, 7, "Kategori", "1", 0, "", true, 0, "")
	pdf.CellFormat(40, 7, "Jumlah", "1", 0, "", true, 0, "")
	pdf.Ln(-1)

	pdf.SetFont("Helvetica", "", 9)
	var income, expense int64
	for _, tx := range txs {
		if pdf.GetY() > 270 {
			pdf.AddPage()
		}
		desc := tx.Description
		if desc == "" {
			desc = "(tanpa deskripsi)"
		}
		cat := tx.CategoryName
		if cat == "" {
			cat = "Lainnya"
		}
		amount := tx.Amount
		if tx.Type == domain.TypeIncome {
			income += amount
		} else {
			expense += amount
		}
		pdf.CellFormat(22, 7, tx.TransactionDate.Format("02/01"), "1", 0, "", false, 0, "")
		pdf.CellFormat(78, 7, truncatePDF(desc, 38), "1", 0, "", false, 0, "")
		pdf.CellFormat(30, 7, truncatePDF(cat, 14), "1", 0, "", false, 0, "")
		align := "R"
		prefix := "-"
		if tx.Type == domain.TypeIncome {
			prefix = "+"
		}
		pdf.CellFormat(40, 7, prefix+format.Rupiah(amount), "1", 0, align, false, 0, "")
		pdf.Ln(-1)
	}

	pdf.Ln(4)
	pdf.SetFont("Helvetica", "B", 10)
	pdf.Cell(0, 6, fmt.Sprintf("Masuk : %s", format.Rupiah(income)))
	pdf.Ln(6)
	pdf.Cell(0, 6, fmt.Sprintf("Keluar: %s", format.Rupiah(expense)))
	pdf.Ln(6)
	pdf.Cell(0, 6, fmt.Sprintf("Net   : %s", format.Rupiah(income-expense)))

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func truncatePDF(s string, max int) string {
	r := []rune(s)
	if len(r) > max {
		return string(r[:max-3]) + "..."
	}
	return s
}

func periodLabel(p parser.Period) string {
	switch p {
	case parser.PeriodDaily:
		return "harian"
	case parser.PeriodWeekly:
		return "mingguan"
	default:
		return "bulanan"
	}
}

func fileSlug(label string) string {
	replacer := strings.NewReplacer(" ", "-", "/", "-")
	return replacer.Replace(strings.ToLower(label))
}
