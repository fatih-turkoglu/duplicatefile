package i18n

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	src := "\xef\xbb\xbf# yorum\n; yorum\n\n@name = Türkçe\n@locale = tr-TR\n" +
		"nav.scan = Tara\n" +
		"  boşluklu.anahtar   =   değer  \n" +
		"satır.sonu = Bir\\nİki \\\\ ters\n" +
		"eşittir = a = b\n" +
		"bozuk satır\n" +
		"nav.scan = Tara (ikinci)\n" +
		"@bilinmeyen = x\n" +
		"boş =\n"
	f, err := Parse("tr", strings.NewReader(src))
	if err != nil {
		t.Fatal(err)
	}
	if f.Name != "Türkçe" || f.Locale != "tr-TR" || f.Code != "tr" {
		t.Errorf("meta: %+v", f)
	}
	want := map[string]string{
		"nav.scan":         "Tara (ikinci)", // son tekrar geçerli
		"boşluklu.anahtar": "değer",
		"satır.sonu":       "Bir\nİki \\ ters",
		"eşittir":          "a = b", // yalnızca ilk = ayırır
		"boş":              "",
	}
	for k, v := range want {
		if f.Entries[k] != v {
			t.Errorf("%s = %q, beklenen %q", k, f.Entries[k], v)
		}
	}
	if len(f.Entries) != len(want) {
		t.Errorf("anahtar sayısı = %d: %v", len(f.Entries), f.Entries)
	}
	// bozuk satır, tekrar eden anahtar, bilinmeyen ayar
	if len(f.Warnings) != 3 {
		t.Errorf("uyarılar = %v", f.Warnings)
	}
}

func TestParseWindowsLineEndings(t *testing.T) {
	f, _ := Parse("de", strings.NewReader("@name = Deutsch\r\na = b\r\n"))
	if f.Name != "Deutsch" || f.Entries["a"] != "b" {
		t.Errorf("CRLF: %+v", f)
	}
}

func TestListAndLoad(t *testing.T) {
	dir := t.TempDir()
	write := func(name, content string) {
		os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644)
	}
	write("tr.lng", "@name = Türkçe\na = b\n")
	write("DE.lng", "@name = Deutsch\n") // büyük harfli ad: de
	write("pt-br.lng", "@name = Português\n")
	write("notlar.txt", "dil dosyası değil\n")
	write("../x.lng", "")                    // klasör dışında
	write("geçersiz kod.lng", "@name = X\n") // geçersiz kod: atlanır

	got := List(dir)
	var codes []string
	for _, i := range got {
		codes = append(codes, i.Code)
	}
	if strings.Join(codes, ",") != "de,pt-br,tr" {
		t.Errorf("List = %v", codes)
	}
	if f, err := LoadFile(dir, "TR"); err != nil || f.Entries["a"] != "b" {
		t.Errorf("LoadFile: %+v %v", f, err)
	}
	for _, bad := range []string{"../tr", `..\tr`, "tr/../de", "", "türkçe"} {
		if _, err := LoadFile(dir, bad); err == nil {
			t.Errorf("geçersiz kod kabul edildi: %q", bad)
		}
	}
	if l := List(filepath.Join(dir, "yok")); l == nil || len(l) != 0 {
		t.Errorf("olmayan klasör: %v", l)
	}
}

// Projedeki gerçek dil dosyaları Go ayrıştırıcısıyla uyarısız okunmalı
// (arayüzdeki tutarlılık testi kendi ayrıştırıcısını kullanır; ikisi uyuşmalı).
func TestRealLanguageFiles(t *testing.T) {
	dir := filepath.Join("..", "..", "lang")
	list := List(dir)
	if len(list) < 7 {
		t.Fatalf("lang klasöründe %d dil var: %+v", len(list), list)
	}
	en, err := LoadFile(dir, "en")
	if err != nil {
		t.Fatal(err)
	}
	for _, info := range list {
		f, err := LoadFile(dir, info.Code)
		if err != nil {
			t.Errorf("%s: %v", info.Code, err)
			continue
		}
		if len(f.Warnings) > 0 {
			t.Errorf("%s uyarıları: %v", info.Code, f.Warnings)
		}
		if f.Name == f.Code || f.Locale == f.Code {
			t.Errorf("%s: @name/@locale eksik (%q, %q)", info.Code, f.Name, f.Locale)
		}
		if _, ok := f.Entries["scan.start"]; !ok {
			t.Errorf("%s: scan.start yok", info.Code)
		}
	}
	if en.Entries["rename.error.chars"] != `The name cannot contain: < > : " / \ | ? *` {
		t.Errorf("kaçış çözümü: %q", en.Entries["rename.error.chars"])
	}
}

func TestDirEnv(t *testing.T) {
	t.Setenv("DUPLICATEFILE_LANG_DIR", `C:\diller`)
	if Dir() != `C:\diller` {
		t.Errorf("Dir = %s", Dir())
	}
	t.Setenv("DUPLICATEFILE_LANG_DIR", "")
	if !strings.HasSuffix(Dir(), "lang") {
		t.Errorf("varsayılan Dir = %s", Dir())
	}
}
