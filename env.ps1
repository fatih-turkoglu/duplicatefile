# Proje araçlarını yalnızca bu terminal oturumu için etkinleştirir.
# Kullanım:  . .\env.ps1   (başındaki nokta önemli)
# Sistem PATH'ine, kayıt defterine veya kullanıcı profiline hiçbir şey yazılmaz.

$tools = Join-Path $PSScriptRoot '.tools'

# Go/npm'in kullanıcı profiline yazdığı ayar ve telemetri dosyaları da .tools içinde kalsın
$env:APPDATA      = Join-Path $tools 'appdata'
New-Item -ItemType Directory -Force $env:APPDATA | Out-Null

$env:GOROOT       = Join-Path $tools 'go'
$env:GOPATH       = Join-Path $tools 'gopath'
$env:GOCACHE      = Join-Path $tools 'cache\go-build'
$env:GOMODCACHE   = Join-Path $tools 'gopath\pkg\mod'
$env:GOTOOLCHAIN  = 'local'
$env:npm_config_cache  = Join-Path $tools 'cache\npm'
$env:npm_config_prefix = Join-Path $tools 'node'

$paths = @(
    (Join-Path $env:GOROOT 'bin'),
    (Join-Path $env:GOPATH 'bin'),
    (Join-Path $tools 'node'),
    (Join-Path $tools 'git\cmd')
)
$env:PATH = ($paths -join ';') + ';' + $env:PATH

Write-Host "DuplicateFile ortamı etkin: $(go version), node $(node --version)" -ForegroundColor Green