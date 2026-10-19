# DuplicateFile

Windows ve macOS için kopya ve benzer dosya bulucu. Klasörleri tarar, içeriği birebir aynı
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
  (sistemin diline göre otomatik; yenileri dil dosyası eklenerek).

---

## Kullanım

### macOS kurulumu

- **macOS 12 Monterey veya üzeri.** Apple Silicon (M serisi) ve Intel desteklenir;
  Universal paket her ikisini de içerir. Alt sürüm sınırı [Go 1.26 gereksiniminden](https://go.dev/wiki/MinimumRequirements) gelir.
- ZIP’i açın, **DuplicateFile.app** dosyasını **Uygulamalar** klasörüne sürükleyin ve açın.
  Dil dosyaları uygulamanın içindedir; ayrıca `lang` klasörü taşımanız gerekmez.
- WKWebView macOS ile gelir; WebView2 kurulumuna gerek yoktur.
- Yerel derleme ad hoc imzalanır. Başka bilgisayarlara dağıtılacak paket için Apple
  Developer ID imzası ve notarizasyon gerekir; yerel imza Apple onayı yerine geçmez.
- macOS, seçtiğiniz Masaüstü, Belgeler, İndirilenler, harici disk veya ağ klasörüne
  erişim için izin sorabilir. Klasör seçimi sistemin kendi penceresiyle yapılır.
- Dosya gösterme düğmesi **Finder’da göster**, varsayılan silme **Çöp Sepeti’ne taşı**
  olarak çalışır. Çöp Sepeti’ne taşıma başarısızsa dosya kalıcı silinmez.
- Mac sistem yazı tipi, sürüklenebilir bütünleşik başlık, standart pencere düğmeleri,
  yerel uygulama/Düzen menüleri ve **⌘C / ⌘V / ⌘X / ⌘Z / ⌘A / ⌘Q** kısayolları kullanılır.

### Windows kurulumu

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
- Klasör simgesi dosyayı **Finder’da** (macOS) ya da **Gezgin’de** (Windows) gösterir.
- Seçili dosyalar için:
  - **Sil** — varsayılan olarak **Geri Dönüşüm Kutusu'na** gönderir; kalıcı
    silme ayrıca işaretlenmelidir.
  - **Taşı** — hedefte aynı adlı dosya varsa yeni ad verir ya da o dosyayı atlar.
  - **Yeniden adlandır** — tek dosyada yeni ad; birden çok dosyada `{ad}`,
    `{uzantı}`, `{n}` şablonu ve önizleme.

### Güvenlik

- **Tarama kapsamı:** `/` gibi tüm diski kapsayan taramalar sistem ve uygulama
  dosyalarını da sonuçlara dahil edebilir. Aynı içerik, dosyanın gereksiz olduğu
  anlamına gelmez; farklı uygulamalar kendi kopyalarına ihtiyaç duyabilir.
  Toplu temizliği Masaüstü, Belgeler, İndirilenler ve Resimler gibi kişisel
  klasörlerle sınırlayın; `.app/Contents` ve sistem klasörlerindeki dosyaları
  kopya olarak görünmeleri nedeniyle silmeyin. macOS’ta **Gizli ve sistem
  dosyalarını atla** açıkken bilinen sistem klasörleri, kullanıcı Library klasörü
  ve uygulama/arşiv paketleri tarama dışında bırakılır. Bu koruma tüm olası
  uygulama veri dosyalarını tanımaz; kişisel klasörlerle sınırlı tarama önerilir.
- **Mac uygulama koruması:** bilinen sistem yolları ve `.app`, `.framework`,
  `.bundle`, Fotoğraflar arşivi gibi paketlerin içeriği, eski tarama sonuçlarından
  seçilse dahi silinemez, taşınamaz veya yeniden adlandırılamaz. Bu konumlar
  son kopya korumasında kişisel dosyanın yedeği sayılmaz. Aynı fiziksel dosyaya
  çıkan farklı yollar da bağımsız bir yedek olarak kabul edilmez.
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

Windows’ta arayüz dilleri exe’nin yanındaki **`lang`** klasöründen okunur:

```
DuplicateFile.exe
lang\
  tr.lng  de.lng  fr.lng  pt.lng  ro.lng  bg.lng  en.lng
```

- Dil, işletim sisteminin dil sırasına göre **otomatik** seçilir (macOS’ta Sistem Ayarları → Dil ve Bölge).
  Ayarlar → Görünüm → **Dil** ile elle de seçilebilir.
- `lang` klasörü yoksa ya da boşsa uygulama **İngilizce** çalışır (İngilizce programın içindedir).
- Bir dil dosyasında eksik metin varsa o metin İngilizce gösterilir.
- Yeni dil eklemek: `en.lng`'yi kopyalayıp dil koduyla adlandırın (ör. `it.lng`), `@name`,
  `@locale` ve metinleri çevirip `lang` klasörüne koyun. Biçim `en.lng`'nin başında anlatılır.

Çeviriler (Türkçe dışındakiler) ana dili o dil olan biri tarafından gözden geçirilmemiştir.

macOS’ta çeviriler `DuplicateFile.app/Contents/Resources/lang` içindedir. İmzalı bir
uygulamanın içeriğini değiştirmek imzayı bozacağından, özel çeviriler için
`DUPLICATEFILE_LANG_DIR` ile ayrı klasör kullanın.

### Verilerin konumu

macOS:

| Ne | Nerede |
|---|---|
| Veritabanı | `~/Library/Application Support/DuplicateFile/index.db` (Ayarlar’dan taşınabilir) |
| Ayar dosyası | `~/Library/Application Support/DuplicateFile/config.json` |
| WKWebView tercihleri | `~/Library/WebKit/com.duplicatefile.app` ve `~/Library/Preferences/com.duplicatefile.app.plist` |

Veritabanı ve ayarlar sistemin silebildiği `Library/Caches` yerine kalıcı uygulama
verileri klasöründedir.

Windows:

| Ne | Nerede |
|---|---|
| Veritabanı (önbellek) | `%LOCALAPPDATA%\DuplicateFile\index.db` (Ayarlar'dan taşınabilir) |
| Ayar dosyası | `%LOCALAPPDATA%\DuplicateFile\config.json` |
| Arayüz tercihleri (tema, tarama seçenekleri) | WebView2 profili: `%APPDATA%\DuplicateFile.exe\` |

`DUPLICATEFILE_DATA_DIR` ortam değişkeni tanımlıysa veritabanı ve ayar dosyası o
klasörde tutulur (taşınabilir kullanım ve testler için).

`DUPLICATEFILE_LANG_DIR` tanımlıysa dil dosyaları varsayılan konum yerine o klasörden okunur.

Uygulamayı tamamen kaldırmak için `.app` / `.exe` dosyasını (Windows’ta yanındaki `lang` klasörünü de) ve ilgili platformun yukarıdaki veri klasörlerini silmeniz yeterlidir.

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
SQLite (saf Go, `modernc.org/sqlite`). Masaüstü kabuğu ve macOS’un yerel Foundation bağlantısı CGO kullanır.

### macOS geliştirme

Gerekenler: **Go 1.26.x**, **Node.js 22.13+** (24 LTS önerilir), npm ve Xcode
Command Line Tools. Geliştirme araçları kullanıcı ortamında ya da `.tools/go`,
`.tools/bin` altında bulunabilir; sistem ayarları değiştirilmez.

```bash
xcode-select --install  # araçlar kurulu değilse, bir kez
source ./env.sh
go install github.com/wailsapp/wails/v2/cmd/wails@v2.16.0
wails dev
```

`env.sh`, Bash ve zsh ile kaynaklanabilir. Go/npm önbelleklerini proje içinde
saklar ve geliştirmede dil dosyalarını doğrudan `lang/` klasöründen okutur.

```bash
./build.sh                 # bu Mac’in mimarisi, testler dahil
./build.sh --universal     # Apple Silicon + Intel, tek .app
./build.sh --arch amd64    # yalnızca Intel
./build.sh --skip-tests    # testleri atla
```

Çıktı: **`build/bin/DuplicateFile.app`** ve sürüm/mimari adlı ZIP.
Betik kilitli npm bağımlılıklarını kurar, arayüzü derler, Go vet/test ve arayüz
testlerini çalıştırır; dilleri pakete koyar, son paketi imzalayıp doğrular.
`wails.json` sürümün tek kaynağıdır. Universal derleme macOS üzerinde yapılır.
Yalnızca yerel deneme için ad hoc imza kullanılır; Developer ID ile imzalamak için
`DUPLICATEFILE_SIGN_IDENTITY` tanımlanabilir. Dağıtım için ayrıca Apple notarizasyonu gerekir.

Temiz klonda `go test ./...` öncesi `cd frontend && npm ci && npm run build`
çalışmalıdır; Go arayüz dosyalarını içine gömer. `build.sh` bu sırayı otomatik uygular.

### Windows ortamı

Araçlar sisteme kurulmaz; proje içindeki `.tools\` klasöründe taşınabilir olarak durur
(Go, Node.js, Wails CLI, Git, önbellekler). Her terminalde önce:

```powershell
. .\env.ps1
```

Başındaki nokta ve boşluk önemlidir; `.\env.ps1` olarak çalıştırılırsa ayarlar terminalde kalmaz.

Windows betiği *"is not digitally signed"* / `PSSecurityException` hatasıyla engellerse
(varsayılan çalıştırma politikası ya da zip'ten gelen "internetten indirildi" işareti),
aynı terminalde önce şunları çalıştırın:

```powershell
Set-ExecutionPolicy -Scope Process -ExecutionPolicy Bypass -Force   # yalnızca bu terminal için; yönetici gerekmez
Get-ChildItem -Recurse -Filter *.ps1 | Unblock-File                 # zip'ten gelen engel işaretini kaldırır (bir kez yeterli)
. .\env.ps1
```

İlk satır her yeni terminalde gerekir. Kalıcı olarak yalnızca kendi kullanıcınız için
yerel betiklere izin vermek isterseniz bir kez `Set-ExecutionPolicy -Scope CurrentUser RemoteSigned`
çalıştırabilirsiniz. Hata sürerse `Get-ExecutionPolicy -List` çıktısına bakın: `MachinePolicy` ya da
`UserPolicy` doluysa politika Group Policy ile belirlenmiştir ve bu komutlarla değiştirilemez.

Bu yalnızca o terminal oturumunun `PATH`'ini ayarlar. `.tools\` yoksa (ör. depoyu yeni
klonladıysanız): Go 1.26 ve Node.js 22.13+’ın resmi Windows **zip** sürümlerini `.tools\go` ve
`.tools\node` klasörlerine, isterseniz Git for Windows **PortableGit**'i `.tools\git`
klasörüne açın; sonra `go install github.com/wailsapp/wails/v2/cmd/wails@latest`.

### Windows çalıştırma ve derleme

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

Uçtan uca testler kurulu Chrome/Edge’i `playwright-core` ile sürer ve `wails dev`
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

macOS’ta uçtan uca testleri ayrı terminalde aynı şekilde çalıştırabilirsiniz:

```bash
# Terminal 1: yalnızca geçici test verisi
source ./env.sh
DUPLICATEFILE_DATA_DIR="$PWD/.tools/e2e-data" wails dev
# Terminal 2
node frontend/e2e/fileops.mjs
node frontend/e2e/similar.mjs
node frontend/e2e/language.mjs with-lang
node frontend/e2e/macos.mjs  # Mac tema, klavye, Finder/Çöp Sepeti metinleri
node frontend/e2e/settings.mjs "$PWD/.tools/e2e-data"
```

Chrome/Edge standart konumda değilse `DUPLICATEFILE_BROWSER` ile çalıştırılabilir
programın tam yolunu verin. Ekran görüntüleri `e2e-out/` içine yazılır.
Gerçek Çöp Sepeti testi yalnızca istenince çalışır ve geçici bir test dosyasını
Çöp Sepeti’nde bırakır:
`DUPLICATEFILE_TEST_RECYCLE=1 go test ./internal/fileops -run MacRecycleIntegration`.

Gerçek Windows Geri Dönüşüm Kutusu testi yalnızca istenince çalışır:
`$env:DUPLICATEFILE_TEST_RECYCLE = "1"; go test ./internal/fileops -run Recycle`

### Dil desteği (geliştirici)

- İngilizce metinler `frontend/src/i18n/en.ts` içindedir (tek kaynak). Arayüzde metinler
  `t('anahtar', { yer: değer })` ile kullanılır; sayıya bağlı metinlerde `{ count }` verilir ve
  dilin çoğul kuralına göre `.one` / `.few` / `.other` … eki seçilir.
- Yeni metin eklendiğinde: `en.ts`'e ekleyin → `cd frontend; npm run lang:en` (`lang/en.lng`'yi
  üretir) → diğer `lang/*.lng` dosyalarına çevirisini ekleyin. `npm test` her dosyada eksik/fazla
  anahtar, yer tutucu ve çoğul eki uyumunu denetler.
- macOS’a özgü metinler `.mac` ekiyle tanımlanır (ör. `file.showInExplorer.mac`, çoğulda `outcome.recycled.mac.other`); diğer platformlar temel anahtarı kullanır.
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
internal/i18n          .lng dil dosyaları ve sistem dil tercihi
internal/macos         Foundation: Çöp Sepeti, sistem dili ve tema
internal/pathutil      platforma uygun dosya yolu karşılaştırması
internal/appdirs       platforma uygun kalıcı veri klasörü
internal/apperr        arayüze giden kodlu hatalar
lang/                  dil dosyaları (build.sh .app içine, build.ps1 exe yanına kopyalar)
cmd/scancli            tarama motorunu arayüzsüz deneme ve ölçme aracı
cmd/genicon            uygulama simgesini (build/appicon.png) çizer
frontend/src           React arayüzü (api.ts: Go ile tek bağlantı noktası)
frontend/e2e           uçtan uca testler
```

macOS taramalarında `.Trash`, `.Trashes`, `.Spotlight-V100` ve `.fseventsd`
klasörleri her zaman atlanır. Gizli dosya filtresi noktalı adlara ek olarak Finder’ın
`hidden` bayrağını da dikkate alır; seçenek açıkken sistem klasörleri ve uygulama paketleri de atlanır. Unix yollarında harf büyüklüğü korunur; harfe
duyarlı APFS birimlerinde `Foto` ve `foto` ayrı dizinlerdir. Veritabanı v3 migration’ı
mevcut önbelleği koruyarak platforma uygun karşılaştırmayı uygular.
