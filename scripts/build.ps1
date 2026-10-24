# Cross-compiles Cryptomon for Linux, macOS and Windows from Windows PowerShell.
# Usage:  powershell -ExecutionPolicy Bypass -File scripts\build.ps1
$ErrorActionPreference = "Stop"
Set-Location (Join-Path $PSScriptRoot "..")

$Version = if ($env:VERSION) { $env:VERSION } else { "1.0.0" }
$Out = "dist"
$LdFlags = "-s -w -X cryptomon/internal/app.Version=$Version"
$Targets = @("windows/amd64", "windows/arm64", "linux/amd64", "linux/arm64", "darwin/amd64", "darwin/arm64")

$env:CGO_ENABLED = "0"
foreach ($t in $Targets) {
    $os, $arch = $t.Split("/")
    $ext = if ($os -eq "windows") { ".exe" } else { "" }
    $dir = Join-Path $Out "cryptomon-$Version-$os-$arch"
    New-Item -ItemType Directory -Force -Path $dir | Out-Null
    Write-Host "building $os/$arch"
    $env:GOOS = $os
    $env:GOARCH = $arch
    foreach ($cmd in @("cryptomon", "currency-service", "gateway")) {
        go build -trimpath -ldflags $LdFlags -o (Join-Path $dir "$cmd$ext") "./cmd/$cmd"
        if ($LASTEXITCODE -ne 0) { throw "build failed for $t" }
    }
    Copy-Item README.md $dir
}
Remove-Item Env:GOOS, Env:GOARCH
Write-Host "done: $Out\"
