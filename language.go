package main

import "DuplicateFile/internal/i18n"

// LanguageSetup, arayüzün dil seçmesi için gerekenlerdir.
type LanguageSetup struct {
	Dir       string      `json:"dir"`       // dil dosyalarının klasörü
	Available []i18n.Info `json:"available"` // lang klasöründeki diller
	System    []string    `json:"system"`    // sistemin tercih ettiği diller, öncelik sırasıyla
}

// LanguageSetup, mevcut dil dosyalarını ve sistemin dil tercihlerini döner.
func (a *App) LanguageSetup() LanguageSetup {
	dir := i18n.Dir()
	return LanguageSetup{Dir: dir, Available: i18n.List(dir), System: i18n.SystemLanguages()}
}

// LoadLanguage, lang klasöründeki <code>.lng dosyasını okur.
func (a *App) LoadLanguage(code string) (i18n.File, error) {
	return i18n.LoadFile(i18n.Dir(), code)
}
