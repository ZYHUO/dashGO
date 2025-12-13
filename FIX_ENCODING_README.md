# UTF-8 编码修复工具使用说明

## 问题描述
项目中的 Go 文件存在 UTF-8 编码错误，中文注释显示为乱码（�?）。

## 解决方案
我们提供了两个自动修复脚本：
- `fix-encoding.ps1` - Windows PowerShell 脚本
- `fix-encoding.sh` - Linux/macOS Bash 脚本

## 使用方法

### Windows (PowerShell)

1. 打开 PowerShell（以管理员身份运行）

2. 如果是第一次运行脚本，需要允许脚本执行：
```powershell
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
```

3. 运行修复脚本：
```powershell
.\fix-encoding.ps1
```

### Linux/macOS (Bash)

1. 给脚本添加执行权限：
```bash
chmod +x fix-encoding.sh
```

2. 运行修复脚本：
```bash
./fix-encoding.sh
```

## 脚本功能

1. **自动扫描** - 扫描项目中所有 Go 文件（排除 vendor、node_modules、.git 目录）
2. **批量替换** - 使用预定义的字符映射表批量替换损坏的字符
3. **UTF-8 保存** - 使用 UTF-8 without BOM 格式保存文件
4. **编译验证** - 修复完成后自动运行 `go build ./...` 验证
5. **详细报告** - 显示修复的文件数量和替换次数

## 输出示例

```
开始修复 UTF-8 编码错误...
找到 156 个 Go 文件
✓ 修复: internal/handler/agent.go (23 处)
✓ 修复: internal/service/telegram.go (45 处)
✓ 修复: pkg/utils/crypto.go (8 处)

修复完成!
总文件数: 156
修复文件数: 3
总替换次数: 76

正在验证编译...
✓ 编译成功!
```

## 支持的字符替换

脚本包含 100+ 个常见的损坏字符映射，包括：
- �? → 和
- �? → 或
- �? → 从
- �? → 件
- �? → 态
- ... 等等

完整列表请查看脚本源代码。

## 注意事项

1. **备份建议** - 运行脚本前建议先提交 Git 或创建备份
2. **编码设置** - 确保你的编辑器使用 UTF-8 编码
3. **手动检查** - 脚本完成后建议手动检查关键文件
4. **编译测试** - 修复后运行完整的测试套件

## 如果脚本无法解决问题

如果自动脚本无法完全修复编码问题，可以：

1. **使用 VS Code**
   - 打开文件
   - 点击右下角的编码（可能显示为 GBK 或其他）
   - 选择 "通过编码重新打开"
   - 选择 UTF-8
   - 手动修复乱码
   - 保存为 UTF-8

2. **使用 iconv 命令**
```bash
iconv -f GBK -t UTF-8 input.go > output.go
```

3. **重新输入注释**
   - 对于严重损坏的文件，直接重新输入中文注释

## 预防措施

为避免将来出现编码问题：

1. **配置 Git**
```bash
git config --global core.autocrlf false
git config --global core.safecrlf true
```

2. **配置编辑器**
   - VS Code: 设置 `"files.encoding": "utf8"`
   - GoLand: Settings → Editor → File Encodings → UTF-8

3. **使用 .editorconfig**
```ini
[*.go]
charset = utf-8
end_of_line = lf
```

## 故障排除

### 问题：PowerShell 提示无法运行脚本
**解决**：运行 `Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser`

### 问题：Bash 提示权限不足
**解决**：运行 `chmod +x fix-encoding.sh`

### 问题：修复后仍有编译错误
**解决**：
1. 检查 `go build` 的具体错误信息
2. 手动检查报错的文件
3. 可能需要手动修复一些特殊字符

### 问题：某些字符没有被替换
**解决**：
1. 在脚本中添加新的字符映射
2. 或手动修复这些字符

## 技术支持

如有问题，请查看：
- `ENCODING_FIX_NEEDED.md` - 详细的编码问题说明
- 项目 Issues 页面

## 许可证

这些脚本与项目使用相同的许可证。
