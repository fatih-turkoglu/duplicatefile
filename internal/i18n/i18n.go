// Package i18n, exe'nin yanındaki lang klasöründeki .lng dil dosyalarını
// okur. İngilizce metinler arayüze gömülüdür; dil dosyası yoksa ya da bir
// anahtar eksikse İngilizce kullanılır.
//
// Biçim (UTF-8, BOM olabilir):
//
//	# yorum (; ile başlayan satırlar da yorumdur)
//	@name = Türkçe          dilin kendi adı (seçim listesinde görünür)
//	@locale = tr-TR         sayı/tarih biçimi için
//	nav.scan = Tara
//	scan.found.one = {count} file found     dile özgü çoğul ekleri: zero/one/two/few/many/other
//	help.text = Birinci satır\nİkinci satır  kaçışlar: \n \t \\
package i18n

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// File, okunmuş bir dil dosyasıdır.
type File struct {
	Code    string            `json:"code"`   // dosya adından: tr, pt-br …
	Name    string            `json:"name"`   // @name; yoksa kod
	Locale  string            `json:"locale"` // @locale; yoksa kod
	Entries map[string]string `json:"entries"`
	// Warnings, dosyanın yüklenmesini engellemeyen sorunlardır (çevirmen için).
	Warnings []string `json:"warnings"`
}

// Info, dil listesinde gösterilen özettir.
type Info struct {
	Code   string `json:"code"`
	Name   string `json:"name"`
	Locale string `json:"locale"`
}

// codePattern, dosya adı olarak kabul edilen dil kodları: tr, en, pt-br …
// (yol ayracı vb. içeremez; LoadFile'a dışarıdan gelen kod için de kullanılır).
var codePattern = regexp.MustCompile(`^[a-z]{2,3}(-[a-z0-9]{2,8})?$`)

// ValidCode, kodun geçerli bir dil kodu olup olmadığını söyler.
func ValidCode(code string) bool {
	return codePattern.MatchString(code)
}

// Parse bir dil dosyasını okur. Bozuk satırlar atlanır ve uyarı olarak
// bildirilir; hata yalnızca okuma başarısız olursa döner.
func Parse(code string, r io.Reader) (File, error) {
	f := File{Code: code, Name: code, Locale: code, Entries: map[string]string{}, Warnings: []string{}}
	data, err := io.ReadAll(r)
	if err != nil {
		return f, err
	}
	data = bytes.TrimPrefix(data, []byte("\xef\xbb\xbf"))

	sc := bufio.NewScanner(bytes.NewReader(data))
	sc.Buffer(make([]byte, 64*1024), 1024*1024)
	for n := 1; sc.Scan(); n++ {
		line := strings.TrimSpace(sc.Text())
		if line == "" || line[0] == '#' || line[0] == ';' {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		key = strings.TrimSpace(key)
		if !ok || key == "" {
			f.Warnings = append(f.Warnings, fmt.Sprintf("line %d: expected 'key = value'", n))
			continue
		}
		value = unescape(strings.TrimSpace(value))

		if meta, isMeta := strings.CutPrefix(key, "@"); isMeta {
			switch meta {
			case "name":
				f.Name = value
			case "locale":
				f.Locale = value
			default:
				f.Warnings = append(f.Warnings, fmt.Sprintf("line %d: unknown setting @%s", n, meta))
			}
			continue
		}
		if _, dup := f.Entries[key]; dup {
			f.Warnings = append(f.Warnings, fmt.Sprintf("line %d: duplicate key %s (last one is used)", n, key))
		}
		f.Entries[key] = value
	}
	return f, sc.Err()
}

func unescape(s string) string {
	if !strings.Contains(s, `\`) {
		return s
	}
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == '\\' && i+1 < len(s) {
			switch s[i+1] {
			case 'n':
				b.WriteByte('\n')
				i++
				continue
			case 't':
				b.WriteByte('\t')
				i++
				continue
			case '\\':
				b.WriteByte('\\')
				i++
				continue
			}
		}
		b.WriteByte(s[i])
	}
	return b.String()
}

// Dir, dil dosyalarının klasörüdür: DUPLICATEFILE_LANG_DIR tanımlıysa o,
// değilse exe'nin yanındaki "lang".
func Dir() string {
	if d := os.Getenv("DUPLICATEFILE_LANG_DIR"); d != "" {
		return d
	}
	exe, err := os.Executable()
	if err != nil {
		return "lang"
	}
	return languageDir(exe)
}

// macOS bundles keep writable/customisable translations in Contents/Resources.
func languageDir(exe string) string {
	dir := filepath.Dir(exe)
	if filepath.Base(dir) == "MacOS" && filepath.Base(filepath.Dir(dir)) == "Contents" {
		return filepath.Join(filepath.Dir(dir), "Resources", "lang")
	}
	return filepath.Join(dir, "lang")
}

// List, klasördeki geçerli dil dosyalarını ada göre sıralı döner. Klasör
// yoksa boş liste döner (uygulama İngilizce çalışır).
func List(dir string) []Info {
	matches, _ := filepath.Glob(filepath.Join(dir, "*.lng"))
	out := []Info{}
	for _, p := range matches {
		code := strings.ToLower(strings.TrimSuffix(filepath.Base(p), filepath.Ext(p)))
		if !ValidCode(code) {
			continue
		}
		f, err := LoadFile(dir, code)
		if err != nil {
			continue
		}
		out = append(out, Info{Code: f.Code, Name: f.Name, Locale: f.Locale})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// LoadFile, klasördeki <code>.lng dosyasını okur.
func LoadFile(dir, code string) (File, error) {
	code = strings.ToLower(code)
	if !ValidCode(code) {
		return File{}, fmt.Errorf("invalid language code %q", code)
	}
	fh, err := os.Open(filepath.Join(dir, code+".lng"))
	if err != nil {
		return File{}, err
	}
	defer fh.Close()
	return Parse(code, fh)
}
