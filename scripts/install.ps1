$ErrorActionPreference = "Stop"

[Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12

$Repo = "JuanCarlosAcostaPeraba/sparks-cli"
$Version = if ($env:SPARKS_VERSION) { $env:SPARKS_VERSION } else { "latest" }
$InstallDir = if ($env:SPARKS_INSTALL_DIR) { $env:SPARKS_INSTALL_DIR } else { Join-Path $env:LOCALAPPDATA "Programs\sparks" }

$Arch = switch ($env:PROCESSOR_ARCHITECTURE) {
    "AMD64" { "amd64" }
    "ARM64" { "arm64" }
    default { throw "sparks installer: unsupported architecture: $env:PROCESSOR_ARCHITECTURE" }
}

if ($Arch -eq "arm64") {
    throw "sparks installer: Windows arm64 binaries are not published yet."
}

if ($Version -eq "latest") {
    $Release = Invoke-RestMethod -Uri "https://api.github.com/repos/$Repo/releases/latest"
    $Version = $Release.tag_name
}

if ([string]::IsNullOrWhiteSpace($Version)) {
    throw "sparks installer: could not resolve the latest release."
}

$Tag = if ($Version.StartsWith("v")) { $Version } else { "v$Version" }
$ReleaseVersion = $Tag.TrimStart("v")
$Asset = "sparks_${ReleaseVersion}_windows_${Arch}.zip"
$Url = "https://github.com/$Repo/releases/download/$Tag/$Asset"
$TempDir = Join-Path ([System.IO.Path]::GetTempPath()) ("sparks-install-" + [System.Guid]::NewGuid().ToString())
$ArchivePath = Join-Path $TempDir $Asset

New-Item -ItemType Directory -Path $TempDir | Out-Null

try {
    Write-Host "Installing sparks $Tag for windows/$Arch..."
    Invoke-WebRequest -Uri $Url -OutFile $ArchivePath
    Expand-Archive -Path $ArchivePath -DestinationPath $TempDir -Force

    New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
    Copy-Item -Path (Join-Path $TempDir "sparks.exe") -Destination (Join-Path $InstallDir "sparks.exe") -Force

    $UserPath = [Environment]::GetEnvironmentVariable("Path", "User")
    $PathEntries = $UserPath -split ";" | Where-Object { $_ }
    if ($PathEntries -notcontains $InstallDir) {
        $NewPath = if ($UserPath) { "$UserPath;$InstallDir" } else { $InstallDir }
        [Environment]::SetEnvironmentVariable("Path", $NewPath, "User")
        $env:Path = "$env:Path;$InstallDir"
        Write-Host "Added $InstallDir to your user PATH. Open a new terminal if sparks is not found."
    }

    Write-Host "sparks installed to $(Join-Path $InstallDir "sparks.exe")"
}
finally {
    Remove-Item -LiteralPath $TempDir -Recurse -Force -ErrorAction SilentlyContinue
}
