#!/bin/bash

# UTF-8 编码修复脚本
# 用于修复 Go 文件中的中文字符编码问题

echo -e "\033[32m开始修复 UTF-8 编码错误...\033[0m"

# 定义字符替换数组（格式：错误字符|正确字符）
declare -a replacements=(
    "�?|和"
    "�?|或"
    "�?|从"
    "�?|件"
    "�?|态"
    "�?|息"
    "�?|机"
    "�?|量"
    "�?|计"
    "�?|组"
    "�?|细"
    "�?|本"
    "�?|端"
    "�?|务"
    "�?|钥"
    "�?|为"
    "�?|回"
    "�?|口"
    "�?|符"
    "�?|户"
    "�?|餐"
    "�?|版"
    "�?|令"
    "�?|输"
    "�?|看"
    "�?|找"
    "�?|他"
    "�?|定"
    "�?|败"
    "�?|功"
    "�?|先"
    "�?|绑"
    "�?|未"
    "�?|请"
    "�?|确"
    "�?|取"
    "�?|解"
    "�?|吗"
    "�?|正"
    "�?|无"
    "�?|已"
    "�?|总"
    "�?|元"
    "�?|理"
    "�?|表"
    "�?|转"
    "�?|换"
    "�?|用"
    "�?|于"
    "�?|客"
    "�?|非"
    "�?|返"
    "�?|列"
    "�?|尝"
    "�?|试"
    "�?|获"
    "�?|处"
    "�?|每"
    "�?|个"
    "�?|的"
    "�?|前"
    "�?|使"
    "�?|匹"
    "�?|配"
    "�?|记"
    "�?|录"
    "�?|日"
    "�?|节"
    "�?|点"
    "�?|当"
    "�?|最"
    "�?|新"
    "�?|数"
    "�?|据"
    "�?|库"
    "�?|查"
    "�?|询"
    "�?|参"
    "�?|更"
    "�?|检"
    "�?|在"
    "�?|线"
    "�?|上"
    "�?|报"
    "�?|生"
    "�?|成"
    "�?|服"
    "�?|移"
    "�?|除"
    "�?|中"
    "�?|连"
    "�?|字"
    "�?|完"
    "�?|整"
    "�?|加"
    "�?|密"
    "�?|方"
    "�?|式"
    "�?|创"
    "�?|建"
    "�?|时"
    "�?|间"
    "�?|戳"
    "�?|格"
    "�?|仅"
    "�?|随"
    "�?|范"
    "�?|围"
    "�?|选"
    "�?|择"
    "�?|证"
    "�?|码"
)

# 查找所有 Go 文件
mapfile -t go_files < <(find . -name "*.go" -type f ! -path "*/vendor/*" ! -path "*/node_modules/*" ! -path "*/.git/*")

total_files=${#go_files[@]}
fixed_files=0
total_replacements=0

echo -e "\033[36m找到 $total_files 个 Go 文件\033[0m"

for file in "${go_files[@]}"; do
    file_replacements=0
    temp_file="${file}.tmp"
    
    # 复制原文件
    cp "$file" "$temp_file"
    
    # 应用所有替换
    for replacement in "${replacements[@]}"; do
        IFS='|' read -r old new <<< "$replacement"
        if grep -q "$old" "$temp_file" 2>/dev/null; then
            count=$(grep -o "$old" "$temp_file" | wc -l)
            sed -i "s/$old/$new/g" "$temp_file" 2>/dev/null || sed -i '' "s/$old/$new/g" "$temp_file" 2>/dev/null
            file_replacements=$((file_replacements + count))
        fi
    done
    
    # 如果有修改，替换原文件
    if [ $file_replacements -gt 0 ]; then
        mv "$temp_file" "$file"
        fixed_files=$((fixed_files + 1))
        total_replacements=$((total_replacements + file_replacements))
        echo -e "\033[33m✓ 修复: $file ($file_replacements 处)\033[0m"
    else
        rm "$temp_file"
    fi
done

echo -e "\n\033[32m修复完成!\033[0m"
echo -e "\033[36m总文件数: $total_files\033[0m"
echo -e "\033[36m修复文件数: $fixed_files\033[0m"
echo -e "\033[36m总替换次数: $total_replacements\033[0m"

# 验证编译
echo -e "\n\033[32m正在验证编译...\033[0m"
if go build ./... 2>&1; then
    echo -e "\033[32m✓ 编译成功!\033[0m"
else
    echo -e "\033[31m✗ 编译失败，请检查错误\033[0m"
fi
