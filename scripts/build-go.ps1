param(
    [string]$Version = "dev",
    [string]$BuildTime = "",
    [switch]$SkipTest
)

$ErrorActionPreference = "Stop"

if ([string]::IsNullOrWhiteSpace($BuildTime)) {
    $BuildTime = (Get-Date -Format o)
}

New-Item -ItemType Directory -Path dist -Force | Out-Null

Write-Host "[1/3] go mod tidy"
go mod tidy

if (-not $SkipTest) {
    Write-Host "[2/3] go test"
    go test ./...
} else {
    Write-Host "[2/3] go test (skipped)"
}

Write-Host "[3/3] go build (windows amd64)"
$env:GOOS = "windows"
$env:GOARCH = "amd64"
go build -ldflags="-s -w -X main.Version=$Version -X main.BuildTime=$BuildTime" -o dist/bitcraft-tsbc-windows-amd64.exe ./cmd/bitcraft-tsbc

Write-Host "Done: dist/bitcraft-tsbc-windows-amd64.exe"
