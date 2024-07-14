package main

import (
	"fmt"     // Paket fmt digunakan untuk mencetak output ke konsol
	"strings" // Paket strings digunakan untuk melakukan operasi terkait string
)

// Struktur RowData digunakan untuk menyimpan informasi tentang sebuah website
type RowData struct {
	RankWebsite int
	Domain      string
	TLD         string
	IDN_TLD     string
	Valid       bool
	RefIPs      int
}

// Fungsi GetTLD digunakan untuk mendapatkan TLD dan IDN_TLD dari sebuah domain
func GetTLD(domain string) (TLD string, IDN_TLD string) {
	// ListIDN_TLD adalah peta yang berisi pasangan TLD dan IDN_TLD khusus Indonesia
	var ListIDN_TLD = map[string]string{
		".com": ".co.id",
		".org": ".org.id",
		".gov": ".go.id",
	}

	// Iterasi mundur melalui domain untuk mencari titik (.) pertama dari belakang
	for i := len(domain) - 1; i >= 0; i-- {
		if domain[i] == '.' {
			TLD = domain[i:] // TLD diisi dengan bagian domain setelah titik (.)
			break
		}
	}

	// Jika TLD ada di ListIDN_TLD, maka gunakan IDN_TLD yang sesuai, jika tidak, gunakan TLD yang sama
	if _, ok := ListIDN_TLD[TLD]; ok {
		return TLD, ListIDN_TLD[TLD]
	} else {
		return TLD, TLD
	}
}

// Fungsi ProcessGetTLD digunakan untuk mengisi TLD dan IDN_TLD pada RowData dan menangani error
func ProcessGetTLD(website RowData, ch chan RowData, chErr chan error) {
	// Mengecek apakah domain kosong
	if website.Domain == "" {
		chErr <- fmt.Errorf("domain name is empty") // Mengirimkan error ke channel chErr
		return
	}
	// Mengecek apakah domain tidak valid
	if !website.Valid {
		chErr <- fmt.Errorf("domain not valid") // Mengirimkan error ke channel chErr
		return
	}
	// Mengecek apakah RefIPs tidak valid
	if website.RefIPs == -1 {
		chErr <- fmt.Errorf("domain RefIPs not valid") // Mengirimkan error ke channel chErr
		return
	}

	// Mendapatkan TLD dan IDN_TLD dari domain
	TLD, IDN_TLD := GetTLD(website.Domain)
	website.TLD = TLD         // Mengisi TLD pada website
	website.IDN_TLD = IDN_TLD // Mengisi IDN_TLD pada website

	ch <- website // Mengirimkan website yang sudah diisi TLD dan IDN_TLD ke channel ch
}

// Variabel FuncProcessGetTLD digunakan sebagai goroutine di fungsi FilterAndGetDomain
var FuncProcessGetTLD = ProcessGetTLD

// Fungsi FilterAndFillData digunakan untuk melakukan filter dan pengisian data secara konkuren
func FilterAndFillData(TLD string, data []RowData) ([]RowData, error) {
	ch := make(chan RowData, len(data)) // Membuat channel ch untuk mengirim RowData
	errCh := make(chan error)           // Membuat channel errCh untuk mengirim error

	// Iterasi melalui data dan menjalankan goroutine ProcessGetTLD
	for _, website := range data {
		go FuncProcessGetTLD(website, ch, errCh)
	}

	filteredData := make([]RowData, 0) // Slice untuk menyimpan data yang sudah difilter
	var errors []string                // Slice untuk menyimpan pesan error

	// Looping untuk menerima data dari channel ch dan error dari channel errCh
	for i := 0; i < len(data); i++ {
		select {
		case website := <-ch:
			// Jika TLD dari website memiliki suffix TLD yang diinginkan, tambahkan ke filteredData
			if strings.HasSuffix(website.TLD, TLD) {
				filteredData = append(filteredData, website)
			}
		case err := <-errCh:
			errors = append(errors, err.Error()) // Menambahkan pesan error ke dalam slice errors
		}
	}

	// Jika ada error, kembalikan error dengan menggabungkan pesan error yang terjadi
	if len(errors) > 0 {
		return nil, fmt.Errorf(strings.Join(errors, ", "))
	}

	// Jika tidak ada error, kembalikan filteredData
	return filteredData, nil
}

// Fungsi main adalah fungsi utama untuk menjalankan program
func main() {
	// Memanggil FilterAndFillData dengan TLD ".com" dan data website yang disediakan
	rows, err := FilterAndFillData(".com", []RowData{
		{1, "google.com", "", "", true, 100},
		{2, "facebook.com", "", "", true, 100},
		{3, "golang.org", "", "", true, 100},
		{4, "", "", "", true, 100},
		{5, "invalid.com", "", "", false, 100},
		{6, "noRefIPs.com", "", "", true, -1},
	})

	// Menangani error jika ada
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		// Mencetak setiap website yang sudah difilter
		for _, row := range rows {
			fmt.Println(row)
		}
	}
}
