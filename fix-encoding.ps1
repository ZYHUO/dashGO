# UTF-8 编码修复脚本
# 用于修复 Go 文件中的中文字符编码问题

Write-Host "开始修复 UTF-8 编码错误..." -ForegroundColor Green

# 定义字符替换映射表
$replacements = @{
    '�?' = '和'
    '�?' = '或'
    '�?' = '从'
    '�?' = '件'
    '�?' = '态'
    '�?' = '息'
    '�?' = '机'
    '�?' = '量'
    '�?' = '计'
    '�?' = '组'
    '�?' = '细'
    '�?' = '本'
    '�?' = '端'
    '�?' = '务'
    '�?' = '钥'
    '�?' = '为'
    '�?' = '回'
    '�?' = '口'
    '�?' = '符'
    '�?' = '户'
    '�?' = '餐'
    '�?' = '版'
    '�?' = '令'
    '�?' = '输'
    '�?' = '看'
    '�?' = '找'
    '�?' = '他'
    '�?' = '定'
    '�?' = '败'
    '�?' = '功'
    '�?' = '先'
    '�?' = '绑'
    '�?' = '未'
    '�?' = '请'
    '�?' = '确'
    '�?' = '取'
    '�?' = '解'
    '�?' = '吗'
    '�?' = '正'
    '�?' = '无'
    '�?' = '已'
    '�?' = '总'
    '�?' = '元'
    '�?' = '理'
    '�?' = '表'
    '�?' = '转'
    '�?' = '换'
    '�?' = '用'
    '�?' = '于'
    '�?' = '客'
    '�?' = '非'
    '�?' = '返'
    '�?' = '列'
    '�?' = '尝'
    '�?' = '试'
    '�?' = '获'
    '�?' = '处'
    '�?' = '每'
    '�?' = '个'
    '�?' = '的'
    '�?' = '前'
    '�?' = '使'
    '�?' = '匹'
    '�?' = '配'
    '�?' = '记'
    '�?' = '录'
    '�?' = '日'
    '�?' = '节'
    '�?' = '点'
    '�?' = '当'
    '�?' = '最'
    '�?' = '新'
    '�?' = '数'
    '�?' = '据'
    '�?' = '库'
    '�?' = '查'
    '�?' = '询'
    '�?' = '参'
    '�?' = '更'
    '�?' = '检'
    '�?' = '在'
    '�?' = '线'
    '�?' = '上'
    '�?' = '报'
    '�?' = '生'
    '�?' = '成'
    '�?' = '服'
    '�?' = '移'
    '�?' = '除'
    '�?' = '中'
    '�?' = '连'
    '�?' = '字'
    '�?' = '完'
    '�?' = '整'
    '�?' = '加'
    '�?' = '密'
    '�?' = '方'
    '�?' = '式'
    '�?' = '创'
    '�?' = '建'
    '�?' = '时'
    '�?' = '间'
    '�?' = '戳'
    '�?' = '格'
    '�?' = '仅'
    '�?' = '随'
    '�?' = '范'
    '�?' = '围'
    '�?' = '选'
    '�?' = '择'
    '�?' = '证'
    '�?' = '码'
    '�?' = '详'
    '�?' = '情'
    '�?' = '创'
    '�?' = '更'
    '�?' = '删'
    '�?' = '发'
    '�?' = '送'
    '�?' = '消'
    '�?' = '带'
    '�?' = '键'
    '�?' = '盘'
    '�?' = '提'
    '�?' = '供'
    '�?' = '邮'
    '�?' = '箱'
    '�?' = '该'
    '�?' = '被'
    '�?' = '其'
    '�?' = '失'
    '�?' = '账'
    '�?' = '认'
    '�?' = '取'
    '�?' = '消'
    '�?' = '知'
    '�?' = '命'
    '�?' = '入'
    '�?' = '帮'
    '�?' = '助'
    '�?' = '封'
    '�?' = '禁'
    '�?' = '过'
    '�?' = '期'
    '�?' = '套'
    '�?' = '余'
    '�?' = '额'
}

# 获取所有 Go 文件
$goFiles = Get-ChildItem -Path . -Filter *.go -Recurse | Where-Object { 
    $_.FullName -notmatch '\\vendor\\' -and 
    $_.FullName -notmatch '\\node_modules\\' -and
    $_.FullName -notmatch '\\.git\\'
}

$totalFiles = $goFiles.Count
$fixedFiles = 0
$totalReplacements = 0

Write-Host "找到 $totalFiles 个 Go 文件" -ForegroundColor Cyan

foreach ($file in $goFiles) {
    $content = Get-Content -Path $file.FullName -Raw -Encoding UTF8
    $originalContent = $content
    $fileReplacements = 0
    
    # 应用所有替换
    foreach ($key in $replacements.Keys) {
        $oldContent = $content
        $content = $content -replace [regex]::Escape($key), $replacements[$key]
        if ($oldContent -ne $content) {
            $count = ([regex]::Matches($oldContent, [regex]::Escape($key))).Count
            $fileReplacements += $count
        }
    }
    
    # 如果有修改，保存文件
    if ($originalContent -ne $content) {
        # 使用 UTF-8 without BOM 保存
        $utf8NoBom = New-Object System.Text.UTF8Encoding $false
        [System.IO.File]::WriteAllText($file.FullName, $content, $utf8NoBom)
        
        $fixedFiles++
        $totalReplacements += $fileReplacements
        Write-Host "✓ 修复: $($file.FullName) ($fileReplacements 处)" -ForegroundColor Yellow
    }
}

Write-Host "`n修复完成!" -ForegroundColor Green
Write-Host "总文件数: $totalFiles" -ForegroundColor Cyan
Write-Host "修复文件数: $fixedFiles" -ForegroundColor Cyan
Write-Host "总替换次数: $totalReplacements" -ForegroundColor Cyan

# 验证编译
Write-Host "`n正在验证编译..." -ForegroundColor Green
$buildResult = go build ./... 2>&1
if ($LASTEXITCODE -eq 0) {
    Write-Host "✓ 编译成功!" -ForegroundColor Green
} else {
    Write-Host "✗ 编译失败，请检查错误:" -ForegroundColor Red
    Write-Host $buildResult -ForegroundColor Red
}
