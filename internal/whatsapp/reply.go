package whatsapp

import (
	"fmt"
	"strings"

	"github.com/derispewss/gonami-projects/internal/application"
	"github.com/derispewss/gonami-projects/internal/domain"
	"github.com/derispewss/gonami-projects/internal/format"
)

func replySaved(tx *domain.Transaction) string {
	var b strings.Builder
	b.WriteString("✅ Dicatat.\n\n")

	emoji := format.CategoryEmoji(tx.CategoryName)
	if tx.CategoryName == "" {
		emoji = "📦"
	}

	b.WriteString(fmt.Sprintf("%s %s\n", emoji, tx.Description))

	if tx.Type == domain.TypeExpense {
		b.WriteString(fmt.Sprintf("💸 %s\n", format.Rupiah(tx.Amount)))
	} else if tx.Type == domain.TypeIncome {
		b.WriteString(fmt.Sprintf("💰 %s\n", format.Rupiah(tx.Amount)))
	} else {
		b.WriteString(fmt.Sprintf("🔄 %s\n", format.Rupiah(tx.Amount)))
	}

	if tx.CategoryName != "" {
		b.WriteString(fmt.Sprintf("📂 %s", tx.CategoryName))
	} else {
		b.WriteString("📂 Lainnya")
	}

	return b.String()
}

func replyDraftConfirm(out *application.RecordOutcome) string {
	var b strings.Builder
	b.WriteString("Saya menemukan transaksi:\n\n")

	emoji := format.CategoryEmoji(out.Parsed.Category)
	if out.Parsed.Category == "" {
		emoji = "📦"
	}

	b.WriteString(fmt.Sprintf("%s %s — %s\n\n", emoji, out.Parsed.Description, format.Rupiah(out.Parsed.Amount)))
	b.WriteString("Simpan transaksi ini? (Balas: iya / tidak)")

	return b.String()
}

func replyReport(r *application.ReportOutcome) string {
	var b strings.Builder
	title := "Laporan"
	switch r.Type {
	case application.ReportDaily:
		title = "Pengeluaran Hari Ini"
	case application.ReportWeekly:
		title = "Pengeluaran Minggu Ini"
	case application.ReportMonthly:
		title = fmt.Sprintf("Pengeluaran %s", format.MonthYearID(r.From))
	}

	b.WriteString(fmt.Sprintf("📊 %s\n\n", title))

	if len(r.Expenses) == 0 {
		b.WriteString("Belum ada pengeluaran.\n")
	} else {
		for _, ex := range r.Expenses {
			emoji := format.CategoryEmoji(ex.CategoryName)
			b.WriteString(fmt.Sprintf("%s %-14s %s\n", emoji, format.Truncate(ex.CategoryName, 14), format.Rupiah(ex.Total)))
		}
	}

	b.WriteString(fmt.Sprintf("\nTotal: %s", format.Rupiah(r.Expense)))
	return b.String()
}

func replyRekap(r *application.ReportOutcome) string {
	var b strings.Builder
	title := "Rekap"
	switch r.Type {
	case application.ReportDaily:
		title = "Rekap Hari Ini"
	case application.ReportWeekly:
		title = "Rekap Minggu Ini"
	case application.ReportMonthly:
		title = fmt.Sprintf("Rekap %s", format.MonthYearID(r.From))
	}

	b.WriteString(fmt.Sprintf("📋 %s\n\n", title))
	b.WriteString(fmt.Sprintf("Masuk : %s\n", format.Rupiah(r.Income)))
	b.WriteString(fmt.Sprintf("Keluar: %s\n", format.Rupiah(r.Expense)))
	b.WriteString("──────────────\n")
	b.WriteString(fmt.Sprintf("Net   : %s", format.Rupiah(r.Income-r.Expense)))

	return b.String()
}

func replyBalance(r *application.BalanceOutcome) string {
	var b strings.Builder
	b.WriteString("💰 Saldo Saat Ini\n\n")
	b.WriteString(fmt.Sprintf("Masuk : %s\n", format.Rupiah(r.TotalIncome)))
	b.WriteString(fmt.Sprintf("Keluar: %s\n", format.Rupiah(r.TotalExpense)))
	b.WriteString("──────────────\n")
	b.WriteString(fmt.Sprintf("Saldo : %s", format.Rupiah(r.NetBalance)))
	return b.String()
}

func replyLastTransactions(txs []*domain.Transaction) string {
	if len(txs) == 0 {
		return "Belum ada transaksi."
	}

	var b strings.Builder
	b.WriteString("🕒 5 Transaksi Terakhir\n\n")

	for i, tx := range txs {
		emoji := format.CategoryEmoji(tx.CategoryName)
		if tx.CategoryName == "" {
			emoji = "📦"
		}

		amt := format.Rupiah(tx.Amount)
		if tx.Type == domain.TypeExpense {
			amt = "-" + amt
		} else if tx.Type == domain.TypeIncome {
			amt = "+" + amt
		}

		dateStr := format.DateShortID(tx.TransactionDate)
		b.WriteString(fmt.Sprintf("%d. %s %s\n", i+1, emoji, format.Truncate(tx.Description, 15)))
		b.WriteString(fmt.Sprintf("   %s  (%s)\n\n", amt, dateStr))
	}

	return strings.TrimSpace(b.String())
}

func helpMessage() string {
	return `*gonami — Personal Finance Assistant* 💰

Catat transaksi dengan cara ngobrol biasa:
• _beli kopi 15k_
• _aku beli ketoprak 25k dan esteh 3k_ → total 28rb
• _transfer 50k ke Budi_
• _bayar listrik 2jt kemarin_
• kirim *foto struk*, *voice note*, atau *PDF*

Tanya apa aja, bebas gaya bahasanya:
• _berapa pengeluaran hari ini?_
• _rekap minggu ini dong_
• _saldo gue sisa berapa_
• _lihat transaksi terakhir_
• _hapus transaksi terakhir_

Balas "iya" untuk simpan draft, "tidak" untuk batal.
Ketik *help* untuk melihat pesan ini lagi.`
}
