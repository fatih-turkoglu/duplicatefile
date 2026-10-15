# DuplicateFile – Çalışma Planı

Wails v2 (Go) + React/TypeScript + MUI ile kopya ve benzer dosya bulucu.
Tamamlanan adımlar `[x]` ile işaretlenir. Faz başlıklarındaki durum:
✅ tamamlandı · 🟨 kısmen · ⬜ başlanmadı

| Faz | Konu | Durum |
|---|---|---|
| 0 | Geliştirme Ortamı | ✅ |
| 1 | Veritabanı Katmanı | ✅ |
| 2 | Tarama Motoru | ✅ |
| 3 | Dosya İşlemleri | ✅ |
| 4 | Arayüz Temeli | ✅ |
| 5 | Tarama Ekranı | ✅ |
| 6 | Sonuç Ekranı | ✅ |
| 7 | İşlem Arayüzü | ✅ |
| 8 | Benzer Görseller | ✅ |
| 9 | Ayarlar ve Kalıcılık | ✅ |
| 10 | Paketleme ve Yayın | 🟨 3/4 |
| 11 | Dil Desteği | ✅ |

---

## ✅ Faz 0 – Geliştirme Ortamı
- [x] Go 1.27.2 portable kurulumu (`.tools\go`, SHA256 doğrulandı)
- [x] Node.js v24.21.0 LTS portable kurulumu (`.tools\node`, SHA256 doğrulandı)
- [x] `env.ps1` – araçları yalnızca oturum için etkinleştiren betik (APPDATA yönlendirme, telemetri kapalı)
- [x] Wails CLI v2.16.0 kurulumu (`.tools\gopath`)
- [x] Wails + React + TypeScript proje iskeleti
- [x] MUI, Material ikonları, Roboto fontu eklendi
- [x] Go bağımlılıkları: `modernc.org/sqlite`, `zeebo/xxh3`
- [x] İlk üretim derlemesi başarılı (`build\bin\DuplicateFile.exe`)

## ✅ Faz 1 – Veritabanı Katmanı (Go)
- [x] Proje klasör yapısı: `internal/db` (`internal/scanner` ve `internal/fileops` kendi fazlarında)
- [x] DB dosyasının konumu (`%LOCALAPPDATA%\DuplicateFile\index.db`) ve açılış ayarları (WAL, busy_timeout)
- [x] Şema: `files` tablosu (id, path, size, mtime, partial_hash, full_hash, last_seen_scan)
- [x] Şema: `scans` tablosu (id, root_path, başlangıç/bitiş, dosya sayısı, durum)
- [x] İndeksler: `path` (unique, NOCASE), `size`, `full_hash`
- [x] Şema sürümleme / migration altyapısı (`PRAGMA user_version`)
- [x] CRUD fonksiyonları ve toplu (transaction) yazma
- [x] Birim testleri
- [x] DB'nin uygulama açılış/kapanışına bağlanması (`app.go`)

## ✅ Faz 2 – Tarama Motoru (Go)
- [x] Klasör gezinme (`filepath.WalkDir`), hariç tutulacak klasörler ve gizli/sistem dosyası filtresi
- [x] Uzantı ve minimum boyut filtresi, iç içe köklerin elenmesi
- [x] Önbellek kontrolü: path + size + mtime aynıysa hash'i DB'den al
- [x] Aşama 1: boyuta göre gruplama (tekil boyutları ele)
- [x] Aşama 2: kısmi hash (ilk/son 4 KB)
- [x] Aşama 3: tam hash (xxh3-128), worker pool ile paralel
- [x] Silinmiş dosyaları DB'den temizleme (görülmeyen kayıtlar diskte kontrol edilir; filtreli taramada önbellek korunur)
- [x] Yalnızca değişen kayıtları DB'ye yazma
- [x] İlerleme olayları (Wails events): aşama, işlenen/toplam, bayt, geçerli dosya
- [x] Tarama iptali (`context.Context`; iptalde hesaplanan hash'ler yine kaydedilir)
- [x] Kopya gruplarını döndüren sorgu (tarama sonucu + DB'den yükleme)
- [x] Birim testleri ve büyük klasörde performans denemesi (`cmd/scancli`)
- [x] Taramanın uygulamaya bağlanması (`StartScan`, `CancelScan`, `IsScanning`, `DuplicateGroups`)

## ✅ Faz 3 – Dosya İşlemleri (Go)
- [x] Silme: Geri Dönüşüm Kutusu'na gönderme (varsayılan; Geri Dönüşüm Kutusu olmayan sürücülerde reddedilir)
- [x] Kalıcı silme seçeneği (onay arayüzü Faz 7'de)
- [x] Taşıma (farklı sürücüler arası kopyala + doğrula + sil; değişiklik zamanı korunur, hata olursa geri alınır)
- [x] Taşımada ad çakışması: atla ya da "ad (2).uzantı"
- [x] Yeniden adlandırma (çakışma kontrolü, Windows ad kuralları, harf büyüklüğü değişimi)
- [x] Her işlemden sonra DB güncelleme
- [x] Toplu işlem sonucu raporu (başarılı/başarısız liste, ilerleme olayı)
- [x] Birim testleri (Geri Dönüşüm Kutusu testi isteğe bağlı: `DUPLICATEFILE_TEST_RECYCLE=1`)
- [x] Uygulamaya bağlama (`DeleteFiles`, `MoveFiles`, `RenameFiles`, `ValidateFileName`; tarama ile aynı anda çalışmaz)

## ✅ Faz 4 – Arayüz Temeli (React + MUI)
- [x] Wails örnek ekranının temizlenmesi (şablon dosyaları, `Greet`)
- [x] MUI tema: açık/koyu mod, sistem temasını takip, seçim kalıcı, Türkçe yerelleştirme
- [x] Pencere başlık çubuğunun temaya uyması; açılışta tema rengiyle arka plan (parlama yok)
- [x] Ana yerleşim: üst çubuk, içerik alanı, responsive breakpoint'ler (geniş: yan menü, orta: dar şerit, dar: alt çubuk)
- [x] Go bağlamalarının (bindings) tipli kullanımı (`src/api.ts`: fonksiyonlar + olaylar)
- [x] Sayfa iskeletleri: Tara, Kopyalar, Ayarlar (tema seçimi ve DB konumu çalışıyor)
- [x] Açık temanın ve Ayarlar sayfasının görsel kontrolü (Faz 5 duman testinde doğrulandı)

## ✅ Faz 5 – Tarama Ekranı
- [x] Klasör seçme (yerel diyalog), birden çok kök klasör
- [x] Tarama seçenekleri (min. boyut, uzantı filtresi + hazır gruplar, gizli dosyalar, hariç klasörler); seçenekler hatırlanır
- [x] İlerleme çubuğu ve aşama göstergesi (geçen süre, işlenen dosya/bayt, geçerli dosya)
- [x] İptal butonu
- [x] Tarama özeti (kopya grubu, kazanılabilir alan, önbellek istatistikleri, erişilemeyen dosyalar)
- [x] Tarama durumunun sayfalar arası korunması; başka sayfadayken menüde gösterge
- [x] Arayüz duman testi (`frontend/e2e/smoke.mjs`, `wails dev` + kurulu Edge)

## ✅ Faz 6 – Sonuç Ekranı
- [x] Kopya grupları listesi (sanal kaydırma: 9.364 dosyada yalnızca ~16 satır çiziliyor)
- [x] Grup başına bilgi: dosya sayısı, boyut, kazanılacak alan, seçim durumu ("Tümü seçili" uyarısı)
- [x] Dosya satırı: ad, klasör, değişiklik tarihi (boyut grup başlığında)
- [x] Seçim yardımcıları: "en yeniyi tut", "en eskiyi tut", "şu klasördekileri seç" (hiçbir grubu tamamen seçmez)
- [x] Sıralama (kazanılabilir alan, boyut, kopya sayısı) ve arama/filtre
- [x] Dosyayı Gezgin'de gösterme (otomatik testte tıklanmadı; elle denenmeli)
- [x] Sonuçların ve seçimin sayfalar arası korunması; açılışta kayıtlı sonuçların yüklenmesi
- [x] Seçim mantığı birim testleri (`npm test`) ve duman testine sonuç ekranı adımları

## ✅ Faz 7 – İşlem Arayüzü
- [x] Arka uç güvenlik katmanı: içeriğin diskte değişmemiş başka bir kopyası kalmayacaksa silme reddedilir (taşıma içerik kaybettirmediği için serbest)
- [x] İşlem sonuçlarına hata kodu (`no_recycle_bin`, `last_copy`, `exists`…)
- [x] Toplu işlem çubuğu (sil / taşı / yeniden adlandır; dar ekranda simge)
- [x] Onay diyalogları (silmede Geri Dönüşüm Kutusu varsayılan, kalıcı silme ayrıca işaretlenir)
- [x] Taşıma için hedef klasör seçimi ve ad çakışması seçeneği
- [x] Yeniden adlandırma diyaloğu (tekli ve şablonlu toplu: `{ad}`, `{uzantı}`, `{n}`; önizleme ve doğrulama)
- [x] İlerleme penceresi, sonuç bildirimi (snackbar) ve hata listesi; Geri Dönüşüm Kutusu olmayan sürücüde "kalıcı sil" seçeneği
- [x] İşlem sonrası listenin güncellenmesi (silinenler çıkar, taşınan/yeniden adlandırılanların yolu değişir)
- [x] Testler: Go (son kopya koruması), `npm test` (şablon), uçtan uca `frontend/e2e/fileops.mjs` (diskte doğrulamalı)

## ✅ Faz 8 – Benzer Görseller (Genişletme)
- [x] Perceptual hash: 64 bit dHash (`internal/imaging`); JPG, PNG, GIF, BMP, TIFF, WebP — HEIC desteklenmez (saf Go çözücü yok), EXIF yönü dikkate alınmaz
- [x] DB şemasına `phash`, `width`, `height` sütunları (migration v2; v1 veritabanı kayıtlarıyla taşınır)
- [x] Tarama modu: "Birebir kopyalar" / "Benzer görseller"; algısal hash önbelleğe alınır, iki mod birbirinin önbelleğini silmez
- [x] Gruplama: Hamming mesafesi + birleşim-bulma (geçişli); referans = en yüksek çözünürlük
- [x] Benzerlik eşiği ayarı (2–16, varsayılan 8)
- [x] Benzer görsel gruplarını önizlemeli gösterme: satırlarda küçük resim, çözünürlük, boyut, benzerlik %; "Karşılaştır" penceresi
- [x] Küçük resim sunucusu (`/thumb`, Wails asset middleware, 64 MB bellek önbelleği; yalnızca görsel uzantılı tam yollar)
- [x] "Her grupta en yüksek çözünürlüklüyü tut" seçim yardımcısı
- [x] Silme koruması benzer görsellerde: silinen her görselin eşik içinde benzer, diskte duran bir eşi kalmalı
- [x] Testler: Go (dHash, küçük resim, migration, benzer tarama, önbellek koruma, silme koruması, /thumb), `npm test`, uçtan uca `frontend/e2e/similar.mjs`
- Doğrulanmayanlar: yalnızca yapay görsellerle test edildi (gerçek fotoğraf arşivinde eşik isabeti ölçülmedi); tüm çift karşılaştırması (n²) büyük arşivde (100.000+) ölçülmedi

## ✅ Faz 9 – Ayarlar ve Kalıcılık
- [x] Ayarlar ekranı: tema; tarama varsayılanları (paralel okuma sayısı, seçenekleri varsayılana döndürme); DB konumu (göster, başka klasöre taşı, varsayılana dön)
- [x] DB taşıma güvenli sırayla (checkpoint → kopya → yeni konumda açılış → eski silme; hata olursa eskiyle devam); konum `config.json`'da
- [x] Ayarlanan DB konumu açılamazsa (ör. sürücü takılı değil) varsayılana dönülür ve Ayarlar'da uyarı gösterilir; ayar korunur
- [x] Son taranan klasörleri hatırlama (Faz 5'te tarama seçenekleriyle birlikte)
- [x] DB bakım: istatistikler, sıkıştırma (VACUUM), silinmiş dosyaların kayıtlarını temizleme, önbelleği tamamen temizleme (onaylı; ekrandaki sonuçlar da kalkar)
- [x] Tarama, dosya işlemi ve bakım birbirini dışlar; veritabanı erişimi kilitli (taşıma sırasında güvenli)
- [x] `DUPLICATEFILE_DATA_DIR` ortam değişkeni: veritabanı ve ayarlar bu klasörde (testler gerçek veriye dokunmasın diye; taşınabilir kullanım için de)
- [x] Testler: Go (bakım, taşıma, geri alma, var olan hedefi reddetme, varsayılana dönüş), uçtan uca `frontend/e2e/settings.mjs` (yalnızca geçici veri klasöründe çalışır)
- Doğrulanmayanlar: farklı sürücüye (ör. C: → D:) DB taşıma elle denenmedi (test aynı sürücüde); "Klasörde göster" düğmesi elle denenmedi

## 🟨 Faz 10 – Paketleme ve Yayın
- [x] Uygulama ikonu (`cmd/genicon` ile çizilir → `build/appicon.png`; exe, pencere ve görev çubuğunda doğrulandı) ve sürüm bilgileri (tek kaynak `wails.json` → exe özellikleri + Ayarlar > Hakkında; Gezgin'de ve uygulamada 1.0.0 doğrulandı)
- [x] Üretim derlemesi ve boyut optimizasyonu: `build.ps1` (önce testler, `-trimpath`, sürüm); exe 17,5 → 16,7 MB; arayüz paketi 1.615 → 863 KB (Roboto yalnızca latin + latin-ext, woff2); exe'de yerel yol kalmadı. Wails zaten `-s -w` kullanıyordu; UPX bilinçli olarak kullanılmadı (antivirüs yanlış alarmları)
- [ ] Temiz bir Windows'ta test — **yapılmadı**: Windows Korumalı Alanı bu bilgisayarda açık değil (açmak yönetici yetkisi + yeniden başlatma ister). Hazır yapılandırma `tools/sandbox/DuplicateFile.wsb` ve kontrol listesi README'de
- [x] README: kullanım, güvenlik, veri konumları, bilinen sınırlar, geliştirme ortamı, testler, proje yapısı
- [x] Kaynak kodu GitHub'a yüklendi: https://github.com/fatih-turkoglu/duplicatefile (git, `.tools\git`'te taşınabilir; derleme çıktıları ve araçlar `.gitignore` ile dışarıda)

## ✅ Faz 11 – Dil Desteği
- [x] `.lng` biçimi (UTF-8 `anahtar = değer`, `@name`/`@locale`, yorum, `{yer tutucu}`, dile özgü çoğul ekleri) ve Go ayrıştırıcısı (`internal/i18n`) + testler; dil kodu doğrulanır (klasör dışına çıkılamaz)
- [x] Exe'nin yanındaki `lang` klasöründen dil listesi ve dosya okuma (`DUPLICATEFILE_LANG_DIR` ile değiştirilebilir); Windows tercih edilen dillerinin okunması
- [x] İngilizce metinler gömülü (`frontend/src/i18n/en.ts`, 314 anahtar): `lang` yoksa/boşsa İngilizce; dosyada eksik anahtar İngilizceye düşer
- [x] Otomatik dil seçimi (Windows dil sırası; tam kod → ana dil → aynı dilin varyantı) + Ayarlar'da elle seçim
- [x] Arayüzdeki bütün metinlerin anahtarlara taşınması; sayı/tarih/yüzde/boyut biçimleri ve birimler (Fr: o/Ko, Bg: КБ) ve MUI bileşen metinleri dile göre
- [x] Dil değişince tarama, sonuçlar ve açık sayfa korunur
- [x] Go'nun kullanıcıya görünen hata mesajları İngilizceye çevrildi ve kodlandı (`internal/apperr`, dosya işlemi ve tarama hata kodları); arayüz koddan çevirir
- [x] Yeniden adlandırma şablonu dilden bağımsız (`{name}`, `{ext}`, `{n}`; eski `{ad}`, `{uzantı}` da çalışır)
- [x] Dil dosyaları: tr, de, fr, pt (pt-BR), ro (one/few/other), bg + şablon `en.lng` (`npm run lang:en` ile üretilir)
- [x] Tutarlılık testi (`lang-files.test.ts`): eksik/fazla anahtar, yer tutucu, dile uygun çoğul ekleri, `@locale`, `en.lng` = `en.ts`; Go tarafı da gerçek dosyaları uyarısız okur
- [x] `build.ps1` dil dosyalarını `build\bin\lang`'a kopyalar; README
- [x] Uçtan uca `frontend/e2e/language.mjs`: dil dosyası yokken İngilizce; otomatik Türkçe; de/ro/fr/bg/pt'ye geçiş, Romence "few", Fransızca birim, sonuçların korunması. Diğer e2e testleri dili Türkçeye sabitler
- Doğrulanmayanlar: Türkçe dışındaki çeviriler ana dili o dil olan biri tarafından gözden geçirilmedi; Go'nun geliştirici günlükleri ve test mesajları Türkçe kaldı (kullanıcıya görünmez); Windows klasör diyaloğunun düğmeleri Windows dilindedir (uygulama dilinden bağımsız)

---

## Hata düzeltmeleri

- [x] **2026-10-15 — Kopyalar silindikten sonra yeniden tarama boş ekranda kalıyordu.**
  Sebep: hiç kopya grubu kalmayınca Go boş grup listesini JSON'da `null` gönderiyordu; özet kartı
  `null.reduce` ile çöküyor, React bütün arayüzü kaldırıyordu. Düzeltme: Go'da gruplar her zaman `[]`
  (`scanner`, `db.DuplicateGroups`); arayüz Go'dan gelen sonucu normalleştiriyor; beklenmeyen hatalarda
  boş ekran yerine mesaj + "Yeniden yükle" gösteren hata sınırı (`ErrorBoundary`).
  Testler: Go `TestScanNoDuplicatesGivesEmptyGroups`; uçtan uca `similar.mjs` ve `fileops.mjs`'e
  "silme sonrası yeniden tarama" adımı. Hata sınırının kendisi elle tetiklenip denenmedi.
