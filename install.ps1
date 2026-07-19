<#
.SYNOPSIS
Installs the latest release of git-swap for Windows.
.DESCRIPTION
Downloads the latest Windows zip from GitHub Releases, extracts it, and copies the binary to a folder in your PATH.
.EXAMPLE
Invoke-WebRequest -Uri https://raw.githubusercontent.com/jokot/git-swap/main/install.ps1 -OutFile install.ps1; .\install.ps1
#>

$ErrorActionPreference = "Stop"

$Repo = "jokot/git-swap"
$BinName = "git-swap.exe"

$Arch = if ($env:PROCESSOR_ARCHITECTURE -eq "AMD64") { "x86_64" } elseif ($env:PROCESSOR_ARCHITECTURE -eq "ARM64") { "arm64" } else { "i386" }

Write-Host "Fetching latest release for Windows_$Arch..."
$ReleaseUrl = "https://api.github.com/repos/$Repo/releases/latest"
$Release = Invoke-RestMethod -Uri $ReleaseUrl -UseBasicParsing

$DownloadUrl = ($Release.assets | Where-Object { $_.name -match "Windows_$Arch\.zip" }).browser_download_url

if (-not $DownloadUrl) {
    Write-Error "Could not find a release for Windows_$Arch"
}

$TempDir = Join-Path $env:TEMP "git-swap-install"
if (Test-Path $TempDir) { Remove-Item -Recurse -Force $TempDir }
New-Item -ItemType Directory -Path $TempDir | Out-Null

$ZipPath = Join-Path $TempDir "git-swap.zip"

Write-Host "Downloading from $DownloadUrl..."
Invoke-WebRequest -Uri $DownloadUrl -OutFile $ZipPath

Write-Host "Extracting..."
Expand-Archive -Path $ZipPath -DestinationPath $TempDir -Force

$InstallDir = Join-Path $env:USERPROFILE "AppData\Local\Microsoft\WindowsApps"
if (-not (Test-Path $InstallDir)) {
    New-Item -ItemType Directory -Path $InstallDir | Out-Null
}

$SourceExe = Join-Path $TempDir $BinName
$DestExe = Join-Path $InstallDir $BinName

Write-Host "Installing to $InstallDir..."
Move-Item -Path $SourceExe -Destination $DestExe -Force

Remove-Item -Recurse -Force $TempDir

Write-Host "✅ Successfully installed $BinName to $InstallDir"
Write-Host "Run 'git-swap import' or 'git-swap add' to get started."