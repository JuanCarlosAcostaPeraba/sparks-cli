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
$ChecksumsUrl = "https://github.com/$Repo/releases/download/$Tag/checksums.txt"
$TempDir = Join-Path ([System.IO.Path]::GetTempPath()) ("sparks-install-" + [System.Guid]::NewGuid().ToString())
$ArchivePath = Join-Path $TempDir $Asset
$ChecksumsPath = Join-Path $TempDir "checksums.txt"

function Download-File {
    param(
        [Parameter(Mandatory = $true)][string]$Uri,
        [Parameter(Mandatory = $true)][string]$Destination
    )

    $Curl = Get-Command curl.exe -ErrorAction SilentlyContinue
    if ($Curl) {
        & $Curl.Source -fsSL -4 --retry 3 --connect-timeout 20 --max-time 120 --http1.1 --tlsv1.2 --ssl-no-revoke -o $Destination $Uri
        if ($LASTEXITCODE -ne 0) {
            throw "sparks installer: curl failed to download $Uri"
        }
        return
    }

    Invoke-WebRequest -Uri $Uri -OutFile $Destination
}

New-Item -ItemType Directory -Path $TempDir | Out-Null

try {
    Write-Host "Installing sparks $Tag for windows/$Arch..."
    Download-File -Uri $Url -Destination $ArchivePath
    Download-File -Uri $ChecksumsUrl -Destination $ChecksumsPath

    $ChecksumLine = Get-Content -LiteralPath $ChecksumsPath | Where-Object {
        $Fields = $_ -split "\s+"
        $Fields.Count -ge 2 -and $Fields[-1].TrimStart("*") -eq $Asset
    } | Select-Object -First 1
    if (-not $ChecksumLine) {
        throw "sparks installer: checksum for $Asset was not published."
    }

    $Expected = (($ChecksumLine -split "\s+")[0]).ToLowerInvariant()
    $Actual = (Get-FileHash -LiteralPath $ArchivePath -Algorithm SHA256).Hash.ToLowerInvariant()
    if ($Actual -ne $Expected) {
        throw "sparks installer: checksum mismatch for $Asset."
    }
    Write-Host "Checksum verified."

    Expand-Archive -Path $ArchivePath -DestinationPath $TempDir -Force
    $BinaryPath = Join-Path $TempDir "sparks.exe"
    if (-not (Test-Path -LiteralPath $BinaryPath)) {
        throw "sparks installer: sparks.exe was not found in $Asset."
    }

    New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
    Copy-Item -LiteralPath $BinaryPath -Destination (Join-Path $InstallDir "sparks.exe") -Force

    $UserPath = [Environment]::GetEnvironmentVariable("Path", "User")
    $PathEntries = $UserPath -split ";" | Where-Object { $_ }
    if ($env:SPARKS_SKIP_PATH_UPDATE -ne "1" -and $PathEntries -notcontains $InstallDir) {
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
