# DuplicateFile – Çalışma Kuralları

## Dil

Bütün cevaplar Türkçe verilir; soru başka bir dilde sorulsa da cevap Türkçedir.

## plan.md her zaman güncel tutulur

`plan.md` projenin tek ilerleme kaydıdır. Durum yalnızca sohbette bildirilmez, dosyaya da işlenir.

- Bir adım bitince hemen `[ ]` → `[x]` yapılır; plandan sapma ya da ek iş varsa adımın metnine yazılır.
- Bir fazın durumu değişince **iki yer birden** güncellenir:
  - Faz başlığındaki simge: `## ✅ Faz N – …`
  - Dosyanın başındaki özet tablo
- Simgeler: ✅ tamamlandı · 🟨 kısmen (tabloda `tamamlanan/toplam` adım) · ⬜ başlanmadı
- Bir faz, yalnızca bütün adımları `[x]` olduğunda ✅ olur.
- Doğrulanamayan şeyler (elle denenmemiş, testle kapsanmamış) adımın metninde belirtilir; işaret ona göre konur.

## Geliştirme ortamı

Araçlar (Go, Node, Wails) `.tools\` altında taşınabilir olarak durur; sisteme kurulum yapılmaz. Her terminalde önce `. .\env.ps1` çalıştırılır.
