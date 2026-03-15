Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

param(
    [string]$Repo = $env:REPO,
    [string]$ServiceName = $env:SERVICE_NAME,
    [string]$InstallDir = $env:INSTALL_DIR,
    [string]$Version = $env:VERSION
)

if (-not $Repo) {
    $Repo = "jcastilloa/context-distill"
}
if (-not $ServiceName) {
    $ServiceName = "context-distill"
}
if (-not $InstallDir) {
    $InstallDir = "$env:LOCALAPPDATA\context-distill\bin"
}

function Resolve-Arch {
    $arch = [System.Runtime.InteropServices.RuntimeInformation]::ProcessArchitecture.ToString().ToLowerInvariant()
    switch ($arch) {
        "x64" { return "amd64" }
        "arm64" { return "arm64" }
        default { throw "Unsupported architecture '$arch' (supported: amd64, arm64)." }
    }
}

function Resolve-Version {
    if ($Version) {
        return $Version
    }

    $apiUrl = "https://api.github.com/repos/$Repo/releases/latest"
    $release = Invoke-RestMethod -Uri $apiUrl
    if (-not $release.tag_name) {
        throw "Could not resolve latest release tag from $apiUrl. Set -Version manually (for example v0.1.0)."
    }

    return [string]$release.tag_name
}

$resolvedVersion = Resolve-Version
$arch = Resolve-Arch
$asset = "${ServiceName}_${resolvedVersion}_windows_${arch}.zip"
$url = "https://github.com/$Repo/releases/download/$resolvedVersion/$asset"

$tmpDir = Join-Path ([System.IO.Path]::GetTempPath()) ("context-distill-install-" + [guid]::NewGuid().ToString("N"))
New-Item -ItemType Directory -Path $tmpDir | Out-Null

try {
    $archivePath = Join-Path $tmpDir $asset
    Write-Host "Downloading $url"
    Invoke-WebRequest -Uri $url -OutFile $archivePath

    Write-Host "Extracting $asset"
    Expand-Archive -Path $archivePath -DestinationPath $tmpDir -Force

    $binary = Get-ChildItem -Path $tmpDir -Recurse -Filter "$ServiceName.exe" | Select-Object -First 1
    if (-not $binary) {
        throw "Binary $ServiceName.exe not found inside archive."
    }

    New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
    $target = Join-Path $InstallDir "$ServiceName.exe"
    Copy-Item -Path $binary.FullName -Destination $target -Force

    Write-Host "Installed: $target"
    if (-not ($env:Path -split ";" | Where-Object { $_ -eq $InstallDir })) {
        Write-Warning "$InstallDir is not in PATH."
    }

    & $target version
}
finally {
    Remove-Item -Path $tmpDir -Recurse -Force -ErrorAction SilentlyContinue
}
