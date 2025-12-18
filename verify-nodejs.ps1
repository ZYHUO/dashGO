# Verify Node.js Installation

Write-Host "Verifying Node.js Installation..." -ForegroundColor Cyan

try {
    $nodeVersion = node --version
    Write-Host "✓ Node.js: $nodeVersion" -ForegroundColor Green
} catch {
    Write-Host "✗ Node.js not found" -ForegroundColor Red
    Write-Host "Please install Node.js from https://nodejs.org/" -ForegroundColor Yellow
    exit 1
}

try {
    $npmVersion = npm --version  
    Write-Host "✓ npm: $npmVersion" -ForegroundColor Green
} catch {
    Write-Host "✗ npm not found" -ForegroundColor Red
    exit 1
}

Write-Host "`nNode.js installation verified successfully!" -ForegroundColor Green
Write-Host "You can now run frontend build commands." -ForegroundColor Cyan