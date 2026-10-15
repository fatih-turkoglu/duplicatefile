# Yayın derlemesi: build\bin\DuplicateFile.exe
#   - sürüm wails.json'daki info.productVersion'dan alınır (exe özellikleri ve uygulama içi aynı)
#   - -trimpath: exe'ye derleme makinesinin yerel yolları gömülmez
#   - önce Go ve arayüz testleri çalışır; biri başarısızsa derleme yapılmaz
# Kullanım:  .\build.ps1            (testlerle)
#            .\build.ps1 -SkipTests
param([switch]$SkipTests)
$ErrorActionPreference = 'Stop'
Set-Location $PSScriptRoot
. .\env.ps1 | Out-Null

$version = (Get-Content wails.json -Raw | ConvertFrom-Json).info.productVersion
if (-not $version) { throw 'wails.json içinde info.productVersion yok' }

if (-not $SkipTests) {
    go vet ./...
    if ($LASTEXITCODE) { throw 'go vet başarısız' }
    go test ./... -count=1
    if ($LASTEXITCODE) { throw 'Go testleri başarısız' }
    Push-Location frontend
    try {
        npm test
        if ($LASTEXITCODE) { throw 'arayüz testleri başarısız' }
    } finally { Pop-Location }
}

# Simge PNG'si değiştiyse .ico yeniden üretilsin (Wails yalnızca yoksa üretir).
$ico = 'build\windows\icon.ico'
if ((Test-Path $ico) -and (Get-Item 'build\appicon.png').LastWriteTime -gt (Get-Item $ico).LastWriteTime) {
    Remove-Item $ico
}

wails build -clean -trimpath -ldflags "-X main.version=$version"
if ($LASTEXITCODE) { throw 'wails build başarısız' }

# Dil dosyaları exe'nin yanındaki lang klasörüne (uygulama oradan okur).
$langOut = 'build\bin\lang'
New-Item -ItemType Directory -Force $langOut | Out-Null
Copy-Item 'lang\*.lng' $langOut -Force
Write-Host ("Dil dosyaları: {0}" -f ((Get-ChildItem $langOut -Filter *.lng).BaseName -join ', '))

$exe = Get-Item 'build\bin\DuplicateFile.exe'
Write-Host ("DuplicateFile {0}: {1} ({2:N1} MB)" -f $version, $exe.FullName, ($exe.Length / 1MB)) -ForegroundColor Green
