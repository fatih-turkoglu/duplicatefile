// Package apperr, arayüze giden hatalara makine tarafından okunabilir bir kod
// ekler. Wails hatayı yalnızca metin olarak iletir; metin "[kod] İngilizce
// açıklama" biçimindedir ve arayüz kodu seçili dile çevirir. Kodu tanımayan
// arayüz İngilizce açıklamayı gösterir.
package apperr

import (
	"errors"
	"fmt"
	"regexp"
)

// Error kodlu bir hatadır.
type Error struct {
	Code string
	Err  error
}

func (e *Error) Error() string { return fmt.Sprintf("[%s] %s", e.Code, e.Err.Error()) }
func (e *Error) Unwrap() error { return e.Err }

// New kodlu yeni bir hata döner.
func New(code, msg string) error {
	return &Error{Code: code, Err: errors.New(msg)}
}

// Errorf kodlu, biçimlendirilmiş bir hata döner (%w desteklenir).
func Errorf(code, format string, args ...any) error {
	return &Error{Code: code, Err: fmt.Errorf(format, args...)}
}

// Code, hata zincirindeki ilk kodu döner; kod yoksa boş metin.
func Code(err error) string {
	var e *Error
	if errors.As(err, &e) {
		return e.Code
	}
	return ""
}

var prefix = regexp.MustCompile(`^\[[a-z_]+\] `)

// Message, kodu atılmış açıklamayı döner (kod iç içe yazılmışsa hepsi atılır).
func Message(err error) string {
	s := err.Error()
	for prefix.MatchString(s) {
		s = prefix.ReplaceAllString(s, "")
	}
	return s
}
