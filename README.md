# DuplicateFile

Windows için kopya ve benzer dosya bulucu. Klasörleri tarar, içeriği birebir aynı
dosyaları ya da birbirine benzeyen fotoğrafları gruplar; seçtiklerinizi güvenle
silmenizi, taşımanızı ya da yeniden adlandırmanızı sağlar.

- **Birebir kopyalar:** boyut → dosyanın başı ve sonu → tamamı (xxh3) sırasıyla
  karşılaştırılır; yalnızca gerektiği kadar okunur.
- **Benzer görseller:** yeniden boyutlandırılmış, sıkıştırılmış ya da hafif
  düzenlenmiş fotoğraflar algısal hash (dHash) ile bulunur.
- **Hızlı tekrar tarama:** taranan dosyaların bilgisi bir veritabanında saklanır;
  değişmeyen dosyalar bir sonraki taramada yeniden okunmaz.
- Açık/koyu tema, pencere boyutuna uyan (responsive) Material arayüz.
- Çok dilli: Türkçe, İngilizce, Almanca, Fransızca, Portekizce, Romence, Bulgarca
  (Windows diline göre otomatik; yenileri dil dosyası eklenerek).

---

## Kullanım

### Kurulum

Kurulum gerekmez: `DuplicateFile.exe`'yi ve yanındaki `lang` klasörünü istediğiniz
klasöre koyup çalıştırın (`lang` olmadan da çalışır; arayüz İngilizce olur).

- Windows 10 ya da 11 (64 bit).
- Microsoft Edge WebView2 Runtime. Windows 11'de ve güncel Windows 10'da zaten
  yüklüdür; yoksa uygulama ilk açılışta indirmeyi önerir.

### Tarama

1. **Tara** sayfasında **Klasör ekle** ile bir ya da daha fazla klasör seçin.
   Klasörler arasındaki kopyalar da bulunur.
2. **Ne aranacak?** — *Birebir kopyalar* ya da *Benzer görseller*.
   Benzer görsellerde **benzerlik eşiği** ayarlanabilir (düşük = daha kesin).
3. İsterseniz dosya boyutu, dosya türü, gizli dosyalar ve hariç tutulan
   klasörler için seçenekleri değiştirin. Seçenekler bir sonraki açılışta hatırlanır.
4. **Taramayı başlat.** Tarama sırasında başka sayfalara geçebilirsiniz.
   **Durdur** derseniz o ana kadar hesaplananlar kaydedilir; sonraki tarama
   kaldığı yerden hızla devam eder.

### Kopyaları inceleme ve işlem yapma

**Kopyalar** sayfasında her grup bir kart olarak listelenir.

- Dosyaları tek tek işaretleyin ya da **Seç** menüsünü kullanın:
  *En yeniyi tut*, *En eskiyi tut*, *En yüksek çözünürlüklüyü tut* (benzer
  görseller), *Klasördeki kopyaları seç*. Bu yardımcılar hiçbir grubun bütün
  dosyalarını seçmez.
- Benzer görsel gruplarında **Karşılaştır**, görselleri büyük önizlemeyle yan yana gösterir.
- Klasör simgesi dosyayı **Gezgin'de** gösterir.
- Seçili dosyalar için:
  - **Sil** — varsayılan olarak **Geri Dönüşüm Kutusu'na** gönderir; kalıcı
    silme ayrıca işaretlenmelidir.
  - **Taşı** — hedefte aynı adlı dosya varsa yeni ad verir ya da o dosyayı atlar.
  - **Yeniden adlandır** — tek dosyada yeni ad; birden çok dosyada `{ad}`,
    `{uzantı}`, `{n}` şablonu ve önizleme.

### Güvenlik

- **Son kopya koruması:** bir dosya, diskte değişmeden duran bir eşi
  (birebir kopyası ya da benzer görsellerde eşik içindeki benzeri) kalmıyorsa
  **silinmez**. Bir grubun bütün dosyaları seçilse bile hepsi silinemez.
- Hiçbir işlem var olan bir dosyanın üzerine yazmaz.
- Geri Dönüşüm Kutusu olmayan sürücülerde (ağ sürücüsü, USB bellek) dosyalar
  sessizce kalıcı silinmez; uygulama bunu sorar.
- Farklı sürücüye taşımada dosya önce kopyalanır, doğrulanır, sonra kaynak silinir.

### Ayarlar

- Tema (sistem / açık / koyu).
- **Aynı anda okunan dosya:** klasik sabit disklerde (HDD) **1** daha hızlı olabilir.
- **Veritabanı:** konumu görüntüleme, başka klasöre (ör. başka sürücüye) taşıma,
  varsayılana döndürme.
- **Bakım:** sıkıştırma, diskte olmayan dosyaların kayıtlarını temizleme,
  önbelleği tamamen temizleme. Bunların hiçbiri dosyalarınıza dokunmaz.

### Dil

Arayüz dilleri exe'nin yanındaki **`lang`** klasöründen okunur:

```
DuplicateFile.exe
lang\
  tr.lng  de.lng  fr.lng  pt.lng  ro.lng  bg.lng  en.lng
```

- Dil, Windows'un dil sırasına göre **otomatik** seçilir (ör. Windows Türkçeyse Türkçe).
  Ayarlar → Görünüm → **Dil** ile elle de seçilebilir.
- `lang` klasörü yoksa ya da boşsa uygulama **İngilizce** çalışır (İngilizce programın içindedir).
- Bir dil dosyasında eksik metin varsa o metin İngilizce gösterilir.
- Yeni dil eklemek: `en.lng`'yi kopyalayıp dil koduyla adlandırın (ör. `it.lng`), `@name`,
  `@locale` ve metinleri çevirip `lang` klasörüne koyun. Biçim `en.lng`'nin başında anlatılır.

Çeviriler (Türkçe dışındakiler) ana dili o dil olan biri tarafından gözden geçirilmemiştir.

### Verilerin konumu

| Ne | Nerede |
|---|---|
| Veritabanı (önbellek) | `%LOCALAPPDATA%\DuplicateFile\index.db` (Ayarlar'dan taşınabilir) |
| Ayar dosyası | `%LOCALAPPDATA%\DuplicateFile\config.json` |
| Arayüz tercihleri (tema, tarama seçenekleri) | WebView2 profili: `%APPDATA%\DuplicateFile.exe\` |

`DUPLICATEFILE_DATA_DIR` ortam değişkeni tanımlıysa veritabanı ve ayar dosyası o
klasörde tutulur (taşınabilir kullanım ve testler için).

`DUPLICATEFILE_LANG_DIR` tanımlıysa dil dosyaları exe'nin yanı yerine o klasörden okunur.

Uygulamayı tamamen kaldırmak için exe'yi (ve yanındaki `lang` klasörünü) ve yukarıdaki klasörleri silmeniz yeterlidir.

### Bilinen sınırlar

- **HEIC** (iPhone fotoğrafları) desteklenmez; benzer görsellerde JPG, PNG, GIF,
  BMP, TIFF ve WebP taranır.
- **EXIF yönü** dikkate alınmaz: telefonun yalnızca etiketle döndürdüğü bir fotoğraf
  ile gerçekten döndürülmüş kopyası eşleşmeyebilir.
- Benzer görsel gruplama tüm çiftleri karşılaştırır; çok büyük arşivlerde
  (100.000+ görsel) yavaşlayabilir.
- Uygulama yeniden açıldığında yalnızca birebir kopya sonuçları veritabanından
  yüklenir; benzer görsel sonuçları için yeniden tarama gerekir.

---

## Geliştirme

Go + [Wails v2](https://wails.io) arka uç, React + TypeScript + [MUI](https://mui.com) arayüz,
SQLite (saf Go, `modernc.org/sqlite`; CGO gerekmez).

### Ortam

Araçlar sisteme kurulmaz; proje içindeki `.tools\` klasöründe taşınabilir olarak durur
(Go, Node.js, Wails CLI, Git, önbellekler). Her terminalde önce:

```powershell
. .\env.ps1
```

Bu yalnızca o terminal oturumunun `PATH`'ini ayarlar. `.tools\` yoksa (ör. depoyu yeni
klonladıysanız): Go ve Node.js'in resmi Windows **zip** sürümlerini `.tools\go` ve
`.tools\node` klasörlerine, isterseniz Git for Windows **PortableGit**'i `.tools\git`
klasörüne açın; sonra `go install github.com/wailsapp/wails/v2/cmd/wails@latest`.

### Çalıştırma ve derleme

```powershell
wails dev          # canlı geliştirme (arayüz değişiklikleri anında yansır)
.\build.ps1        # testler + yayın derlemesi → build\bin\DuplicateFile.exe
```

`build.ps1` önce bütün testleri çalıştırır; sürümü `wails.json` içindeki
`info.productVersion`'dan alır (exe özellikleri ve uygulamadaki *Hakkında* aynı
kaynaktan) ve `-trimpath` ile derler. Sürüm değiştirmek için yalnızca `wails.json`
güncellenir.

### Testler

```powershell
go test ./...                       # Go birim testleri
cd frontend; npm test               # arayüz mantığı (node --test)
```

Uçtan uca testler kurulu Edge'i `playwright-core` ile sürer ve `wails dev`
sunucusuna bağlanır (ayrı terminalde `wails dev` çalışırken):

```powershell
node frontend\e2e\smoke.mjs <taranacak-klasör>   # tarama, sonuç ekranı, tema
node frontend\e2e\fileops.mjs                    # sil / taşı / yeniden adlandır (geçici klasörde)
node frontend\e2e\similar.mjs                    # benzer görseller (geçici klasörde)
node frontend\e2e\language.mjs with-lang         # dil seçimi ve değiştirme (bkz. Dil desteği)

# Ayarlar testi önbelleği siler ve veritabanını taşır: YALNIZCA boş bir veri klasörüyle
$env:DUPLICATEFILE_DATA_DIR = "<boş klasör>"; wails dev
node frontend\e2e\settings.mjs "<aynı klasör>"
```

Gerçek Geri Dönüşüm Kutusu testi yalnızca istenince çalışır:
`$env:DUPLICATEFILE_TEST_RECYCLE = "1"; go test ./internal/fileops -run Recycle`

### Dil desteği (geliştirici)

- İngilizce metinler `frontend/src/i18n/en.ts` içindedir (tek kaynak). Arayüzde metinler
  `t('anahtar', { yer: değer })` ile kullanılır; sayıya bağlı metinlerde `{ count }` verilir ve
  dilin çoğul kuralına göre `.one` / `.few` / `.other` … eki seçilir.
- Yeni metin eklendiğinde: `en.ts`'e ekleyin → `cd frontend; npm run lang:en` (`lang/en.lng`'yi
  üretir) → diğer `lang/*.lng` dosyalarına çevirisini ekleyin. `npm test` her dosyada eksik/fazla
  anahtar, yer tutucu ve çoğul eki uyumunu denetler.
- Go'nun arayüzde gösterilen hataları `[kod] İngilizce açıklama` biçimindedir
  (`internal/apperr`); arayüz kodu `error.<kod>` anahtarıyla çevirir.
- `wails dev` sırasında dil dosyaları `build\bin\lang`'dan okunur (`build.ps1` kopyalar) ya da
  `$env:DUPLICATEFILE_LANG_DIR = "<proje>\lang"` ile doğrudan kaynak klasörden.
- Uçtan uca: `node frontend\e2e\language.mjs no-lang|with-lang` (kullanımı dosyanın başında).

### Temiz Windows'ta test

`tools\sandbox\DuplicateFile.wsb` dosyası uygulamayı **Windows Korumalı Alanı**'nda
(her açılışta sıfırdan başlayan geçici Windows) açar. Özelliğin açık olması gerekir:
*Windows özelliklerini aç veya kapat → Windows Korumalı Alanı* (yönetici yetkisi, yeniden başlatma).

1. `.\build.ps1`
2. `tools\sandbox\DuplicateFile.wsb` dosyasına çift tıklayın (ağ kapalıdır).
3. Kontrol listesi:
   - Uygulama açılıyor, pencere ve görev çubuğu simgesi doğru.
   - Masaüstündeki `DuplicateFile` klasörünü taramaya ekleyin; tarama tamamlanıyor.
   - Ayarlar → Hakkında'da sürüm doğru, veritabanı `%LOCALAPPDATA%\DuplicateFile` altında.
   - Bir klasörde birkaç kopya oluşturup **Geri Dönüşüm Kutusu'na** silin; dosyalar
     Geri Dönüşüm Kutusu'nda görünüyor.

### Proje yapısı

```
main.go, app.go        Wails uygulaması ve arayüze açılan Go fonksiyonları
settings.go            yapılandırma, veritabanı taşıma ve bakım
thumbs.go              /thumb: küçük resim sunucusu (asset middleware)
internal/db            SQLite şeması, migration'lar, sorgular
internal/scanner       tarama motoru (birebir + benzer görsel)
internal/imaging       dHash ve küçük resim
internal/fileops       silme / taşıma / yeniden adlandırma + son kopya koruması
internal/i18n          .lng dil dosyaları ve Windows dil tercihi
internal/apperr        arayüze giden kodlu hatalar
lang/                  dil dosyaları (build.ps1 exe'nin yanına kopyalar)
cmd/scancli            tarama motorunu arayüzsüz deneme ve ölçme aracı
cmd/genicon            uygulama simgesini (build/appicon.png) çizer
frontend/src           React arayüzü (api.ts: Go ile tek bağlantı noktası)
frontend/e2e           uçtan uca testler
plan.md                faz faz çalışma planı ve durumu
```
