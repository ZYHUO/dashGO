# Encoding Fixes Summary

## Fixed Files

All UTF-8 encoding issues (garbled Chinese characters `�`) have been fixed in the following files:

### 1. internal/service/scheduler.go
- Fixed: `每小时执�?` → `每小时执行`
- Fixed: `每分钟执�?` → `每分钟任务`
- Fixed: `每�?号` → `每月1号`
- Fixed: `发送到期提�?` → `发送到期提醒`
- Fixed: `发送邮�?` → `发送邮件`
- Fixed: `发�?Telegram` → `发送 Telegram`
- Fixed: `重置月流�?` → `重置月流量`
- Fixed: `发送流量预�?` → `发送流量预警`
- Fixed: `获取流量使用超过 80% 的用�?` → `获取流量使用超过 80% 的用户`
- Fixed: `删除 90 天前的流量日�?` → `删除 90 天前的流量日志`
- Fixed: `�?v2_server_log 表聚合数�?` → `从 v2_server_log 表聚合数据`
- Fixed: `数据归档和汇�?` → `数据归档和汇总`
- Removed unused variable `recordAt` (compilation error)

### 2. internal/service/server_group.go
- Fixed: `用户组服�?` → `服务器组服务`
- Fixed: `获取用户�?` → `获取服务器组`
- Fixed: `创建用户�?` → `创建服务器组`
- Fixed: `更新用户�?` → `更新服务器组`
- Fixed: `删除用户�?` → `删除服务器组`
- Fixed import paths from `dashgo` to `xboard`

### 3. internal/service/coupon.go
- Fixed: `检查时�?` → `检查时间`
- Fixed: `检查使用次�?` → `检查使用次数`
- Fixed: `检查用户使用次�?` → `检查用户使用次数`
- Fixed: `检查套餐限�?` → `检查套餐限制`
- Fixed: `检查周期限�?` → `检查周期限制`
- Fixed: `百分比折�?` → `百分比折扣`
- Fixed: `使用优惠�?` → `使用优惠券`
- Fixed: `获取优惠�?` → `获取优惠券`
- Fixed: `创建优惠�?` → `创建优惠券`
- Fixed: `更新优惠�?` → `更新优惠券`
- Fixed: `删除优惠�?` → `删除优惠券`
- Fixed: `生成随机�?` → `生成随机码`

### 4. internal/model/user_group.go
- Fixed: `保留这些字段用于向后兼容，但标记为废�?` → `保留这些字段用于向后兼容，但标记为废弃`
- Fixed: `速度限制应该由套餐决�?` → `速度限制应该由套餐决定`
- Fixed: `设备限制应该由套餐决�?` → `设备限制应该由套餐决定`
- Fixed: `获取 server_ids �?int64 数组` → `获取 server_ids 为 int64 数组`
- Fixed: `获取 plan_ids �?int64 数组` → `获取 plan_ids 为 int64 数组`
- Fixed: `检查该组是否可以访问指定节�?` → `检查该组是否可以访问指定节点`
- Fixed: `检查该组是否可以购买指定套�?` → `检查该组是否可以购买指定套餐`

### 5. internal/protocol/base64.go
- Fixed: `获取加密方式，默�?aes-256-gcm` → `获取加密方式，默认 aes-256-gcm`
- Fixed: `密码：对�?SS2022，使�?server.Password（已包含服务器密�?用户密钥格式�?` → `密码：对于 SS2022，使用 server.Password（已包含服务器密钥+用户密钥格式）`
- Fixed: `对于普�?SS，使用用�?UUID` → `对于普通 SS，使用用户 UUID`
- Fixed: `握手服务�?` → `握手服务器`

## Verification

All files have been verified:
- ✅ No compilation errors (verified with getDiagnostics)
- ✅ No remaining garbled characters (verified with grepSearch)
- ✅ All Chinese comments are now properly encoded in UTF-8

## Build Instructions

To compile the project, use the build scripts:

**Windows:**
```powershell
.\build-all.ps1 -Target all -Version 1.0.0
```

**Linux/macOS:**
```bash
./build-all.sh all 1.0.0
```

Or build specific components:
- `.\build-all.ps1 -Target server` - Build server only
- `.\build-all.ps1 -Target agent` - Build agent only
- `.\build-all.ps1 -Target frontend` - Build frontend only
