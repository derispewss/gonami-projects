package parser

import "strings"

type categoryRule struct {
	Name  string
	Words []string
}

var expenseCategoryRules = []categoryRule{
	{"Food & Beverage", []string{
		"makan", "minum", "makanan", "minuman", "kopi", "ngopi", "cafe", "kafe",
		"warung", "resto", "restoran", "indomaret", "alfamart", "pizza",
		"burger", "ayam", "geprek", "bakso", "mie", "mi ", "nasi", "soto",
		"sate", "seblak", "seblac", "boba", "milk tea", "juice", "jus",
		"rokok", "snack", "cemilan", "martabak", "warkop", "starbucks",
		"mcd", "kfc", "hokben", "domino", "gofood", "grabfood", "shopeefood",
		"coca cola", "cola", "teh", "air minum", "sarapan", "brunch",
		"makan siang", "makan malam", "dessert", "es krim", "roti",
	}},
	{"Transportation", []string{
		"bensin", "pertalite", "pertamax", "solar", "bbm", "parkir", "tol",
		"ojek", "gojek", "grab", "grabbike", "maxim", "taxi", "taksi",
		"angkot", "busway", "transjakarta", "kereta", "krl", "mrt", "lrt",
		"e-toll", "etoll", "servis motor", "ganti oli", "oli",
	}},
	{"Shopping", []string{
		"shopee", "tokopedia", "lazada", "tiktok shop", "blibli", "zalora",
		"belanja", "mall", "toko", "marketplace", "baju", "celana", "sepatu",
		"tas", "kosmetik", "skincare", "parfum",
	}},
	{"Bills", []string{
		"listrik", "token listrik", "pln", "pdam", "air", "telkom",
		"internet", "wifi", "indihome", "biznet", "first media", "myrepublic",
		"pulsa", "paket data", "kuota", "cicilan", "kredit", "paylater",
		"iuran", "kos", "kost", "kontrakan", "sewa", "apartemen", "bpjs",
		"asuransi", "pajak", "tagihan",
	}},
	{"Entertainment", []string{
		"netflix", "spotify", "disney", "vidio", "viu", "youtube premium",
		"bioskop", "cgv", "xxi", "cinema", "game", "steam", "mobile legends",
		"genshin", "top up game", "konser", "hiburan", "karaoke", "billiard",
	}},
	{"Health", []string{
		"dokter", "obat", "apotek", "klinik", "rumah sakit", "rs ",
		"vitamin", "suplemen", "gym", "fitness", "yoga", "dokter gigi",
		"vaksin", "medical checkup",
	}},
	{"Education", []string{
		"kursus", "les", "privat", "buku", "sekolah", "kuliah", "spp",
		"training", "workshop", "seminar", "udemy", "coursera", "sertifikasi",
	}},
	{"Travel", []string{
		"hotel", "penginapan", "tiket pesawat", "garuda", "lion air", "citilink",
		"airasia", "liburan", "wisata", "villa", "homestay", "airbnb",
		"traveloka", "tiket.com", "wisata",
	}},
	{"Subscription", []string{
		"subscribe", "langganan", "langganan bulanan", "renewal", "premium",
		"membership", "paket bundling",
	}},
}

var incomeCategoryRules = []categoryRule{
	{"Salary", []string{"gaji", "salary", "payroll", "thr"}},
	{"Freelance", []string{"freelance", "project", "proyek", "client", "klien", "fee", "honor"}},
	{"Business", []string{"bisnis", "usaha", "jualan", "jualan", "omzet", "dagang", "warung"}},
	{"Gift", []string{"hadiah", "kado", "angpao", "amplop", "bonus dari"}},
	{"Investment", []string{"dividen", "deviden", "bunga", "cuan", "saham", "reksadana", "crypto"}},
}

var brands = []struct {
	Keyword  string
	Category string
	Merchant string
}{
	{"indomaret", "Food & Beverage", "Indomaret"},
	{"alfamart", "Food & Beverage", "Alfamart"},
	{"alfamidi", "Food & Beverage", "Alfamidi"},
	{"starbucks", "Food & Beverage", "Starbucks"},
	{"hokben", "Food & Beverage", "HokBen"},
	{"mcd", "Food & Beverage", "McDonald's"},
	{"kfc", "Food & Beverage", "KFC"},
	{"gojek", "Transportation", "Gojek"},
	{"grabbike", "Transportation", "Grab"},
	{"grab", "Transportation", "Grab"},
	{"maxim", "Transportation", "Maxim"},
	{"shopee", "Shopping", "Shopee"},
	{"tokopedia", "Shopping", "Tokopedia"},
	{"netflix", "Subscription", "Netflix"},
	{"spotify", "Subscription", "Spotify"},
}

type categoryResult struct {
	Name     string
	Merchant string
	Matched  bool
}

func DetectCategory(lower string, txType string) categoryResult {
	if txType == "expense" {

		for _, b := range brands {
			if strings.Contains(lower, b.Keyword) || fuzzyMatchWord(lower, b.Keyword, 1) {
				return categoryResult{Name: b.Category, Merchant: b.Merchant, Matched: true}
			}
		}
	}

	var rules []categoryRule
	switch txType {
	case "expense":
		rules = expenseCategoryRules
	case "income":
		rules = incomeCategoryRules
	default:
		return categoryResult{}
	}

	for _, rule := range rules {
		for _, w := range rule.Words {
			if strings.Contains(lower, w) || fuzzyMatchWord(lower, w, 1) {
				return categoryResult{Name: rule.Name, Matched: true}
			}
		}
	}
	return categoryResult{}
}
