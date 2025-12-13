# 简单的UTF-8编码修复脚本
Write-Host "开始修复UTF-8编码错误..." -ForegroundColor Green

# 获取所有Go文件
$goFiles = Get-ChildItem -Path . -Filter *.go -Recurse | Where-Object { 
    $_.FullName -notmatch '\\vendor\\' -and 
    $_.FullName -notmatch '\\node_modules\\' -and
    $_.FullName -notmatch '\\.git\\'
}

$fixedFiles = 0

foreach ($file in $goFiles) {
    try {
        $content = Get-Content -Path $file.FullName -Raw -Encoding UTF8
        $originalContent = $content
        
        # 替换常见的编码错误字符
        $content = $content -replace '�\?', '计'
        $content = $content -replace '�\?', '机'
        $content = $content -replace '�\?', '态'
        $content = $content -replace '�\?', '行'
        $content = $content -replace '�\?', '从'
        $content = $content -replace '�\?', '组'
        $content = $content -replace '�\?', '在'
        $content = $content -replace '�\?', '序'
        $content = $content -replace '�\?', '理'
        $content = $content -replace '�\?', '息'
        $content = $content -replace '�\?', '务'
        $content = $content -replace '�\?', '用'
        $content = $content -replace '�\?', '户'
        $content = $content -replace '�\?', '败'
        $content = $content -replace '�\?', '功'
        $content = $content -replace '�\?', '成'
        $content = $content -replace '�\?', '新'
        $content = $content -replace '�\?', '更'
        $content = $content -replace '�\?', '删'
        $content = $content -replace '�\?', '除'
        $content = $content -replace '�\?', '查'
        $content = $content -replace '�\?', '找'
        $content = $content -replace '�\?', '获'
        $content = $content -replace '�\?', '取'
        $content = $content -replace '�\?', '创'
        $content = $content -replace '�\?', '建'
        $content = $content -replace '�\?', '验'
        $content = $content -replace '�\?', '证'
        $content = $content -replace '�\?', '检'
        $content = $content -replace '�\?', '测'
        $content = $content -replace '�\?', '试'
        $content = $content -replace '�\?', '连'
        $content = $content -replace '�\?', '接'
        $content = $content -replace '�\?', '数'
        $content = $content -replace '�\?', '据'
        $content = $content -replace '�\?', '库'
        $content = $content -replace '�\?', '表'
        $content = $content -replace '�\?', '记'
        $content = $content -replace '�\?', '录'
        $content = $content -replace '�\?', '存'
        $content = $content -replace '�\?', '储'
        $content = $content -replace '�\?', '保'
        $content = $content -replace '�\?', '加'
        $content = $content -replace '�\?', '载'
        $content = $content -replace '�\?', '配'
        $content = $content -replace '�\?', '置'
        $content = $content -replace '�\?', '文'
        $content = $content -replace '�\?', '件'
        $content = $content -replace '�\?', '路'
        $content = $content -replace '�\?', '径'
        $content = $content -replace '�\?', '目'
        $content = $content -replace '�\?', '录'
        $content = $content -replace '�\?', '迁'
        $content = $content -replace '�\?', '移'
        $content = $content -replace '�\?', '模'
        $content = $content -replace '�\?', '型'
        $content = $content -replace '�\?', '结'
        $content = $content -replace '�\?', '构'
        $content = $content -replace '�\?', '自'
        $content = $content -replace '�\?', '动'
        $content = $content -replace '�\?', '执'
        $content = $content -replace '�\?', '行'
        $content = $content -replace '�\?', '状'
        $content = $content -replace '�\?', '态'
        $content = $content -replace '�\?', '显'
        $content = $content -replace '�\?', '示'
        $content = $content -replace '�\?', '过'
        $content = $content -replace '�\?', '滤'
        $content = $content -replace '�\?', '排'
        $content = $content -replace '�\?', '序'
        $content = $content -replace '�\?', '跳'
        $content = $content -replace '�\?', '过'
        $content = $content -replace '�\?', '分'
        $content = $content -replace '�\?', '割'
        $content = $content -replace '�\?', '语'
        $content = $content -replace '�\?', '句'
        $content = $content -replace '�\?', '读'
        $content = $content -replace '�\?', '取'
        $content = $content -replace '�\?', '失'
        $content = $content -replace '�\?', '败'
        $content = $content -replace '�\?', '继'
        $content = $content -replace '�\?', '续'
        $content = $content -replace '�\?', '空'
        $content = $content -replace '�\?', '白'
        $content = $content -replace '�\?', '注'
        $content = $content -replace '�\?', '释'
        $content = $content -replace '�\?', '开'
        $content = $content -replace '�\?', '始'
        $content = $content -replace '�\?', '前'
        $content = $content -replace '�\?', '缀'
        $content = $content -replace '�\?', '忽'
        $content = $content -replace '�\?', '略'
        $content = $content -replace '�\?', '错'
        $content = $content -replace '�\?', '误'
        $content = $content -replace '�\?', '回'
        $content = $content -replace '�\?', '滚'
        $content = $content -replace '�\?', '事'
        $content = $content -replace '�\?', '务'
        $content = $content -replace '�\?', '提'
        $content = $content -replace '�\?', '交'
        $content = $content -replace '�\?', '标'
        $content = $content -replace '�\?', '记'
        $content = $content -replace '�\?', '已'
        $content = $content -replace '�\?', '完'
        $content = $content -replace '�\?', '成'
        $content = $content -replace '�\?', '个'
        $content = $content -replace '�\?', '待'
        
        # 如果有修改，保存文件
        if ($originalContent -ne $content) {
            # 使用 UTF-8 without BOM 保存
            $utf8NoBom = New-Object System.Text.UTF8Encoding $false
            [System.IO.File]::WriteAllText($file.FullName, $content, $utf8NoBom)
            $fixedFiles++
            Write-Host "✓ 修复: $($file.FullName)" -ForegroundColor Yellow
        }
    }
    catch {
        Write-Host "✗ 处理失败: $($file.FullName) - $($_.Exception.Message)" -ForegroundColor Red
    }
}

Write-Host "`n修复完成! 修复了 $fixedFiles 个文件" -ForegroundColor Green

# 验证编译
Write-Host "`n正在验证编译..." -ForegroundColor Green
$buildResult = go build ./... 2>&1
if ($LASTEXITCODE -eq 0) {
    Write-Host "✓ 编译成功!" -ForegroundColor Green
} else {
    Write-Host "✗ 编译失败，请检查错误:" -ForegroundColor Red
    Write-Host $buildResult -ForegroundColor Red
}