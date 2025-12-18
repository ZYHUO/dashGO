# Frontend Setup Script - Run after installing Node.js

Write-Host "========================================" -ForegroundColor Green
Write-Host "Frontend Setup" -ForegroundColor Green  
Write-Host "========================================" -ForegroundColor Green

# Step 1: Verify Node.js
Write-Host "`n[1/3] Verifying Node.js installation..." -ForegroundColor Yellow
try {
    $nodeVersion = node --version
    $npmVersion = npm --version
    Write-Host "  ✓ Node.js: $nodeVersion" -ForegroundColor Green
    Write-Host "  ✓ npm: $npmVersion" -ForegroundColor Green
} catch {
    Write-Host "  ✗ Node.js/npm not found" -ForegroundColor Red
    Write-Host "  Please install Node.js first from https://nodejs.org/" -ForegroundColor Yellow
    exit 1
}

# Step 2: Install dependencies
Write-Host "`n[2/3] Installing dependencies..." -ForegroundColor Yellow
npm install

if ($LASTEXITCODE -eq 0) {
    Write-Host "  ✓ Dependencies installed successfully" -ForegroundColor Green
} else {
    Write-Host "  ✗ Failed to install dependencies" -ForegroundColor Red
    exit 1
}

# Step 3: Build frontend
Write-Host "`n[3/3] Building frontend..." -ForegroundColor Yellow
npm run build

if ($LASTEXITCODE -eq 0) {
    Write-Host "`n========================================" -ForegroundColor Green
    Write-Host "✓ Frontend setup completed successfully!" -ForegroundColor Green
    Write-Host "========================================" -ForegroundColor Green
    
    # Show build info
    if (Test-Path "dist") {
        $distSize = (Get-ChildItem -Recurse dist | Measure-Object -Property Length -Sum).Sum / 1MB
        Write-Host "`nBuild output:" -ForegroundColor Cyan
        Write-Host ("  Size: {0:N2} MB" -f $distSize) -ForegroundColor Cyan
        Write-Host "  Location: web/dist/" -ForegroundColor Cyan
    }
} else {
    Write-Host "`n========================================" -ForegroundColor Red
    Write-Host "✗ Build failed" -ForegroundColor Red
    Write-Host "========================================" -ForegroundColor Red
    Write-Host "Please check the error messages above" -ForegroundColor Yellow
}