# 前端清理和重新构建脚本
# 解决新旧版本冲突问题

Write-Host "========================================" -ForegroundColor Green
Write-Host "前端清理和重新构建" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green

# 1. 清理 dist 目录
Write-Host "`n[1/5] 清理 dist 目录..." -ForegroundColor Yellow
if (Test-Path "dist") {
    Remove-Item -Recurse -Force "dist"
    Write-Host "✓ dist 目录已删除" -ForegroundColor Green
} else {
    Write-Host "✓ dist 目录不存在，跳过" -ForegroundColor Green
}

# 2. 清理 Vite 缓存
Write-Host "`n[2/5] 清理 Vite 缓存..." -ForegroundColor Yellow
if (Test-Path "node_modules\.vite") {
    Remove-Item -Recurse -Force "node_modules\.vite"
    Write-Host "✓ Vite 缓存已清理" -ForegroundColor Green
} else {
    Write-Host "✓ Vite 缓存不存在，跳过" -ForegroundColor Green
}

# 3. 清理 node_modules（可选，但推荐）
Write-Host "`n[3/5] 清理 node_modules..." -ForegroundColor Yellow
$cleanNodeModules = Read-Host "是否清理 node_modules？这会重新安装所有依赖 (y/N)"
if ($cleanNodeModules -eq "y" -or $cleanNodeModules -eq "Y") {
    if (Test-Path "node_modules") {
        Write-Host "正在删除 node_modules（这可能需要一些时间）..." -ForegroundColor Yellow
        Remove-Item -Recurse -Force "node_modules"
        Write-Host "✓ node_modules 已删除" -ForegroundColor Green
    }
    
    Write-Host "正在重新安装依赖..." -ForegroundColor Yellow
    npm install
    Write-Host "✓ 依赖安装完成" -ForegroundColor Green
} else {
    Write-Host "✓ 跳过 node_modules 清理" -ForegroundColor Green
}

# 4. 清理 package-lock.json 缓存
Write-Host "`n[4/5] 检查 package-lock.json..." -ForegroundColor Yellow
if (Test-Path "package-lock.json") {
    Write-Host "✓ package-lock.json 存在" -ForegroundColor Green
} else {
    Write-Host "! package-lock.json 不存在，将在安装时生成" -ForegroundColor Yellow
}

# 5. 重新构建
Write-Host "`n[5/5] 重新构建前端..." -ForegroundColor Yellow
npm run build

if ($LASTEXITCODE -eq 0) {
    Write-Host "`n========================================" -ForegroundColor Green
    Write-Host "✓ 构建成功！" -ForegroundColor Green
    Write-Host "========================================" -ForegroundColor Green
    
    # 显示构建产物信息
    if (Test-Path "dist") {
        Write-Host "`n构建产物信息:" -ForegroundColor Cyan
        $distSize = (Get-ChildItem -Recurse dist | Measure-Object -Property Length -Sum).Sum / 1MB
        Write-Host ("  大小: {0:N2} MB" -f $distSize) -ForegroundColor Cyan
        Write-Host ("  文件数: {0}" -f (Get-ChildItem -Recurse dist | Measure-Object).Count) -ForegroundColor Cyan
    }
    
    Write-Host "`n提示:" -ForegroundColor Yellow
    Write-Host "  1. 如果浏览器仍显示旧版本，请按 Ctrl+Shift+R 强制刷新" -ForegroundColor Yellow
    Write-Host "  2. 或者清除浏览器缓存后重新访问" -ForegroundColor Yellow
} else {
    Write-Host "`n========================================" -ForegroundColor Red
    Write-Host "✗ 构建失败" -ForegroundColor Red
    Write-Host "========================================" -ForegroundColor Red
    Write-Host "请检查上面的错误信息" -ForegroundColor Red
}
