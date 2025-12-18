# Frontend Build Diagnostics Script

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Frontend Build Diagnostics" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan

# Check Node.js
Write-Host "`n[Check] Node.js Version" -ForegroundColor Yellow
try {
    $nodeVersion = node --version
    Write-Host "  OK Node.js: $nodeVersion" -ForegroundColor Green
    
    $versionNumber = [int]($nodeVersion -replace 'v(\d+)\..*', '$1')
    if ($versionNumber -lt 18) {
        Write-Host "  WARNING: Version too low, recommend >= 18.0.0" -ForegroundColor Red
    }
} catch {
    Write-Host "  ERROR: Node.js not installed" -ForegroundColor Red
}

# Check npm
Write-Host "`n[Check] npm Version" -ForegroundColor Yellow
try {
    $npmVersion = npm --version
    Write-Host "  OK npm: $npmVersion" -ForegroundColor Green
} catch {
    Write-Host "  ERROR: npm not installed" -ForegroundColor Red
}

# Check directory structure
Write-Host "`n[Check] Directory Structure" -ForegroundColor Yellow
$dirs = @("src", "node_modules", "dist", "node_modules\.vite")
foreach ($dir in $dirs) {
    if (Test-Path $dir) {
        $size = (Get-ChildItem -Recurse $dir -ErrorAction SilentlyContinue | Measure-Object -Property Length -Sum).Sum / 1MB
        Write-Host ("  OK {0,-20} exists ({1:N2} MB)" -f $dir, $size) -ForegroundColor Green
    } else {
        Write-Host ("  MISSING {0,-20}" -f $dir) -ForegroundColor Yellow
    }
}

# Check key files
Write-Host "`n[Check] Key Files" -ForegroundColor Yellow
$files = @(
    "package.json",
    "package-lock.json", 
    "vite.config.ts",
    "index.html",
    "src\main.ts",
    "src\App.vue"
)
foreach ($file in $files) {
    if (Test-Path $file) {
        Write-Host "  OK $file" -ForegroundColor Green
    } else {
        Write-Host "  MISSING $file" -ForegroundColor Red
    }
}

# Check dist directory content
Write-Host "`n[Check] Dist Directory Content" -ForegroundColor Yellow
if (Test-Path "dist") {
    $indexHtml = Test-Path "dist\index.html"
    $assetsDir = Test-Path "dist\assets"
    
    if ($indexHtml) {
        Write-Host "  OK dist\index.html exists" -ForegroundColor Green
        
        # Check index.html modification time
        $lastModified = (Get-Item "dist\index.html").LastWriteTime
        $timeDiff = (Get-Date) - $lastModified
        Write-Host ("  INFO Last modified: {0} ({1:N0} minutes ago)" -f $lastModified, $timeDiff.TotalMinutes) -ForegroundColor Cyan
        
        if ($timeDiff.TotalHours -gt 1) {
            Write-Host "  WARNING Build output may be outdated, recommend rebuild" -ForegroundColor Yellow
        }
    } else {
        Write-Host "  MISSING dist\index.html" -ForegroundColor Red
    }
    
    if ($assetsDir) {
        $jsFiles = (Get-ChildItem "dist\assets\*.js" -ErrorAction SilentlyContinue).Count
        $cssFiles = (Get-ChildItem "dist\assets\*.css" -ErrorAction SilentlyContinue).Count
        Write-Host "  OK dist\assets exists (JS: $jsFiles, CSS: $cssFiles)" -ForegroundColor Green
    } else {
        Write-Host "  MISSING dist\assets" -ForegroundColor Red
    }
} else {
    Write-Host "  MISSING dist directory, need to build" -ForegroundColor Yellow
}

# Check Vite cache
Write-Host "`n[Check] Vite Cache" -ForegroundColor Yellow
if (Test-Path "node_modules\.vite") {
    $cacheSize = (Get-ChildItem -Recurse "node_modules\.vite" -ErrorAction SilentlyContinue | Measure-Object -Property Length -Sum).Sum / 1MB
    Write-Host ("  INFO Vite cache exists ({0:N2} MB)" -f $cacheSize) -ForegroundColor Cyan
    Write-Host "  INFO If having issues, you can delete this cache" -ForegroundColor Cyan
} else {
    Write-Host "  OK No Vite cache" -ForegroundColor Green
}

# Recommendations
Write-Host "`n========================================" -ForegroundColor Cyan
Write-Host "Recommendations" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan

$hasIssues = $false

if (-not (Test-Path "dist")) {
    Write-Host "• Need to run build: npm run build" -ForegroundColor Yellow
    $hasIssues = $true
}

if (Test-Path "node_modules\.vite") {
    Write-Host "• If cache issues, run: .\clean-build.ps1" -ForegroundColor Yellow
    $hasIssues = $true
}

if (Test-Path "dist\index.html") {
    $lastModified = (Get-Item "dist\index.html").LastWriteTime
    $timeDiff = (Get-Date) - $lastModified
    if ($timeDiff.TotalHours -gt 1) {
        Write-Host "• Build output may be outdated, recommend rebuild" -ForegroundColor Yellow
        $hasIssues = $true
    }
}

if (-not $hasIssues) {
    Write-Host "OK No obvious issues found" -ForegroundColor Green
    Write-Host "`nIf browser still shows old version:" -ForegroundColor Cyan
    Write-Host "  1. Press Ctrl+Shift+R to force refresh browser" -ForegroundColor Cyan
    Write-Host "  2. Clear browser cache" -ForegroundColor Cyan
    Write-Host "  3. Test in incognito mode" -ForegroundColor Cyan
}

Write-Host ""