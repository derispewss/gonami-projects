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
• kirim *foto struk*, *voice note*, atau *PDF* (bisa banyak transaksi sekaligus)

Tanya apa aja, bebas gaya bahasanya:
• _berapa pengeluaran hari ini?_
• _rekap minggu ini dong_
• _saldo gue sisa berapa_
• _lihat transaksi terakhir_
• _hapus transaksi terakhir_

Fitur lanjutan:
• _budget makan 500rb_ → atur budget, cek: _budget_
• _ubah budget makan 700rb_ → sesuaikan budget yang sudah ada
• _insight_ → anomali & pengeluaran rutin
• _export_ / _export pdf_ → unduh laporan bulan ini
• _dompet_ → multi-dompet (buat dompet BCA, pakai dompet BCA)
• _kategori_ → daftar kategori, tambah kategori skincare
• _hapus semua data_ → reset data dari awal (dengan konfirmasi)

Balas "iya" untuk simpan draft, "tidak" untuk batal.
Ketik *help* untuk melihat pesan ini lagi.`
}

func replyBudgetStatus(items []application.BudgetStatus) string {
	if len(items) == 0 {
		return "Belum ada budget.\nAtur dengan: *budget [kategori] [nominal]*\nContoh: budget makan 500rb"
	}

	var b strings.Builder
	b.WriteString("🎯 Budget Bulan Ini\n\n")
	for _, it := range items {
		emoji := format.CategoryEmoji(it.Name)
		bar := progressBar(it.Ratio)
		status := ""
		switch {
		case it.Breach:
			status = " ⚠️ lewat batas"
		case it.Warned:
			status = " ⏳ mendekati batas"
		}
		b.WriteString(fmt.Sprintf("%s %s\n", emoji, format.Truncate(it.Name, 14)))
		b.WriteString(fmt.Sprintf("%s %s / %s (%d%%)%s\n\n",
			bar, format.Rupiah(it.Spent), format.Rupiah(it.Limit),
			int(it.Ratio*100), status))
	}
	return strings.TrimRight(b.String(), "\n")
}

func progressBar(ratio float64) string {
	if ratio < 0 {
		ratio = 0
	}
	filled := int(ratio * 10)
	if filled > 10 {
		filled = 10
	}
	return strings.Repeat("▓", filled) + strings.Repeat("░", 10-filled)
}

func replyInsights(res *application.InsightsResult) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("💡 Insight %s\n\n", format.MonthYearID(res.Month)))

	if !res.HasInsights {
		b.WriteString("Belum ada pola menarik dari transaksimu.\nCatat lebih banyak transaksi dulu ya!")
		return b.String()
	}

	if len(res.Anomalies) > 0 {
		b.WriteString("⚠️ *Anomali pengeluaran*\n")
		for _, a := range res.Anomalies {
			b.WriteString(fmt.Sprintf("%s %s: %s (biasanya ~%s, %.1fx lipat)\n",
				format.CategoryEmoji(a.Category),
				format.Truncate(a.Category, 14),
				format.Rupiah(a.Current),
				format.Rupiah(a.Average),
				a.Multiplier))
		}
		b.WriteString("\n")
	}

	if len(res.Recurring) > 0 {
		b.WriteString("🔁 *Deteksi pengeluaran rutin*\n")
		for _, r := range res.Recurring {
			name := r.Description
			if len(name) > 24 {
				name = name[:21] + "..."
			}
			b.WriteString(fmt.Sprintf("• %s — ~%s/bulan (%dx)\n",
				strings.Title(name), format.Rupiah(r.AvgAmount), r.Count))
		}
	}

	return b.String()
}

func replyWallets(res *application.WalletSummaryResult) string {
	if len(res.Wallets) == 0 {
		return "Belum ada dompet tambahan.\nBuat dengan: *buat dompet [nama]*\nContoh: buat dompet BCA"
	}

	var b strings.Builder
	b.WriteString("👛 Dompet Kamu\n\n")
	for _, w := range res.Wallets {
		marker := "○"
		if res.ActiveID != nil && *res.ActiveID == w.ID {
			marker = "●"
		}
		b.WriteString(fmt.Sprintf("%s %s — keluar bulan ini %s\n",
			marker, w.Name, format.Rupiah(res.Spend[w.Name])))
	}

	b.WriteString("\n● = aktif. Ganti dengan: *pakai dompet [nama]*\nKembali ke utama: *dompet umum*")
	return b.String()
}

func replyCategories(res *application.CategoryListResult) string {
	var b strings.Builder
	b.WriteString("📂 Kategori Tersedia\n\n")

	b.WriteString("*Bawaan:*\n")
	if len(res.Default) == 0 {
		b.WriteString("(kosong)\n")
	} else {
		for _, c := range res.Default {
			b.WriteString(fmt.Sprintf("%s %s\n", format.CategoryEmoji(c), c))
		}
	}

	if len(res.Custom) > 0 {
		b.WriteString("\n*Kustom:*\n")
		for _, c := range res.Custom {
			b.WriteString(fmt.Sprintf("%s %s\n", format.CategoryEmoji(c), c))
		}
		b.WriteString("\nTambah lagi: *tambah kategori [nama]*")
	} else {
		b.WriteString("\nTambah kategori sendiri: *tambah kategori [nama]*")
	}
	return b.String()
}
