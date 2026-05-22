# 发行版 Profile 指南

Bootimus 使用发行版 profile 来检测 ISO 类型并生成正确的引导参数。Profile 由数据驱动 — 无需修改代码即可添加对新发行版的支持。

## 目录

- [概述](#概述)
- [工作原理](#工作原理)
- [查看 Profile](#查看-profile)
- [更新 Profile](#更新-profile)
- [创建自定义 Profile](#创建自定义-profile)
- [Profile 字段](#profile-字段)
- [占位符](#占位符)
- [示例](#示例)
- [故障排查](#故障排查)

## 概述

发行版 profile 定义:
- **如何检测** ISO 属于哪个发行版(文件名模式匹配)
- **从哪里查找** ISO 内部的 kernel、initrd 和 squashfs
- **使用什么引导参数** 进行 PXE 引导
- **支持哪种无人值守安装类型**(preseed、kickstart、autoinstall 等)

### Profile 类型

| 类型 | 说明 |
|------|-------------|
| **内建** | 随 Bootimus 一起发布,可从中央仓库更新 |
| **自定义** | 由用户创建,永不被更新覆盖 |

匹配 ISO 文件名时,自定义 profile 始终优先于内建 profile。

## 工作原理

1. 当一个 ISO 被上传或提取时,Bootimus 把文件名与 profile 的模式进行匹配
2. 使用匹配 profile 的 kernel/initrd 路径来定位 ISO 内部的引导文件
3. profile 的引导参数成为默认值(可在镜像 Properties 中编辑)
4. 引导时,参数中的占位符被解析为实际 URL

### Profile 生命周期

```
Build time:    distro-profiles.json embedded in binary
                        ↓
First startup:  Profiles seeded into database
                        ↓
"Check for Updates":  Latest profiles fetched from GitHub
                        ↓
User creates:   Custom profiles stored in database (never overwritten)
```

## 查看 Profile

在管理面板中进入 **Boot > Distro Profiles**,即可看到所有已加载的 profile 及其文件名模式、引导参数、类型(Built-in/Custom)和版本。

## 更新 Profile

### 自动(推荐)

在 Distro Profiles 标签页点击 **"Check for Updates"**。这会从以下地址抓取最新 profile:

```
https://raw.githubusercontent.com/garybowers/bootimus/main/distro-profiles.json
```

- 自动添加新 profile
- 把现有内建 profile 更新到最新版本
- 自定义 profile 从不被修改

### 通过 API

```bash
curl -H "Authorization: Bearer $TOKEN" -X POST http://localhost:8081/api/profiles/update
```

响应:
```json
{
  "success": true,
  "message": "Updated to version 0.1.21 (2 added, 5 updated)"
}
```

## 创建自定义 Profile

### 通过 Web 界面

1. 进入 **Boot > Distro Profiles**
2. 点击 **"+ Add Custom Profile"**
3. 填写 profile 字段
4. 点击 **"Create Profile"**

### 通过 API

```bash
curl -H "Authorization: Bearer $TOKEN" -X POST http://localhost:8081/api/profiles/save \
  -H "Content-Type: application/json" \
  -d '{
    "profile_id": "my-distro",
    "display_name": "My Custom Distro",
    "family": "debian",
    "filename_patterns": ["mydistro", "my-distro"],
    "kernel_paths": ["/live/vmlinuz", "/boot/vmlinuz"],
    "initrd_paths": ["/live/initrd.img", "/boot/initrd"],
    "squashfs_paths": ["/live/filesystem.squashfs"],
    "default_boot_params": "boot=live initrd=initrd ip=dhcp",
    "boot_params_with_squashfs": "boot=live initrd=initrd fetch={{SQUASHFS}}",
    "auto_install_type": "preseed"
  }'
```

### 删除自定义 Profile

只能删除自定义 profile。内建 profile 在下次更新时会被恢复。

```bash
curl -H "Authorization: Bearer $TOKEN" -X DELETE "http://localhost:8081/api/profiles/delete?id=my-distro"
```

## Profile 字段

| 字段 | 必填 | 说明 |
|-------|----------|-------------|
| `profile_id` | 是 | 唯一标识符(例如 `ubuntu`、`my-distro`) |
| `display_name` | 是 | UI 中显示的可读名称 |
| `family` | 否 | 发行版家族(例如 `debian`、`arch`、`redhat`)— 用于分组 |
| `filename_patterns` | 是 | 用于匹配 ISO 文件名的子串(不区分大小写) |
| `kernel_paths` | 否 | ISO 内部 kernel 的尝试路径(例如 `/casper/vmlinuz`) |
| `initrd_paths` | 否 | ISO 内部 initrd 的尝试路径 |
| `squashfs_paths` | 否 | squashfs 根文件系统的尝试路径 |
| `default_boot_params` | 否 | 默认的 kernel 引导参数(支持占位符) |
| `boot_params_with_squashfs` | 否 | 检测到 squashfs 时使用的备选引导参数 |
| `auto_install_type` | 否 | 无人值守安装格式:`preseed`、`kickstart`、`autoinstall`、`autounattend` |
| `boot_method` | 否 | 覆盖引导方法(例如 Windows 用 `wimboot`) |

## 占位符

引导参数支持以下占位符,在引导时解析:

| 占位符 | 解析为 | 示例 |
|-------------|-------------|---------|
| `{{BASE_URL}}` | 服务器 HTTP URL | `http://192.168.1.10:8080` |
| `{{CACHE_DIR}}` | 提取文件目录 | `ubuntu-24.04-server-amd64` |
| `{{FILENAME}}` | ISO 文件名(URL 编码) | `ubuntu-24.04-server-amd64.iso` |
| `{{SQUASHFS}}` | squashfs 文件的完整 URL | `http://192.168.1.10:8080/boot/ubuntu.../casper/filesystem.squashfs` |

### 带占位符的示例

```
boot=live initrd=initrd fetch={{SQUASHFS}} ip=dhcp
```

解析为:
```
boot=live initrd=initrd fetch=http://192.168.1.10:8080/boot/debian-live-13/live/filesystem.squashfs ip=dhcp
```

## 示例

### 基于 Debian 的 Live ISO

```json
{
  "profile_id": "my-debian-live",
  "display_name": "My Debian Live Spin",
  "family": "debian",
  "filename_patterns": ["my-debian"],
  "kernel_paths": ["/live/vmlinuz"],
  "initrd_paths": ["/live/initrd.img"],
  "squashfs_paths": ["/live/filesystem.squashfs"],
  "default_boot_params": "initrd=initrd boot=live priority=critical",
  "boot_params_with_squashfs": "initrd=initrd boot=live priority=critical fetch={{SQUASHFS}}"
}
```

### 基于 Arch 的发行版

```json
{
  "profile_id": "my-arch-spin",
  "display_name": "My Arch Spin",
  "family": "arch",
  "filename_patterns": ["myarch"],
  "kernel_paths": ["/arch/boot/x86_64/vmlinuz-linux", "/boot/vmlinuz-linux"],
  "initrd_paths": ["/arch/boot/x86_64/initramfs-linux.img", "/boot/initramfs-linux.img"],
  "squashfs_paths": ["/arch/x86_64/airootfs.sfs"],
  "default_boot_params": "archisobasedir=arch archiso_http_srv={{BASE_URL}}/boot/{{CACHE_DIR}}/iso/ ip=dhcp"
}
```

### 基于 RHEL 的安装器

```json
{
  "profile_id": "my-rhel-clone",
  "display_name": "My RHEL Clone",
  "family": "redhat",
  "filename_patterns": ["myrhel"],
  "kernel_paths": ["/images/pxeboot/vmlinuz"],
  "initrd_paths": ["/images/pxeboot/initrd.img"],
  "default_boot_params": "root=live:{{BASE_URL}}/isos/{{FILENAME}} rd.live.image inst.repo={{BASE_URL}}/boot/{{CACHE_DIR}}/iso/ rd.neednet=1 ip=dhcp",
  "auto_install_type": "kickstart"
}
```

## 故障排查

### ISO 未被识别为正确的发行版

检查 ISO 文件名是否匹配任何 profile 模式:

1. 进入 **Distro Profiles** 标签
2. 查看 "Filename Patterns" 列
3. 如果没有模式匹配你的 ISO 文件名,创建一个自定义 profile

### 提取后引导参数错误

1. 打开镜像 **Properties**
2. 在 Boot Parameters 旁点击 **"Re-detect"**
3. 或手动编辑引导参数 — 它们支持占位符

### "Check for Updates" 失败

更新从 GitHub 拉取。检查:
- 服务器是否能访问互联网
- `raw.githubusercontent.com` 没有被屏蔽
- 如果 GitHub 宕机,稍后重试

### 自定义 Profile 未匹配

自定义 profile 优先于内建。请确认:
- `filename_patterns` 包含匹配你 ISO 文件名的子串(不区分大小写)
- profile ID 是唯一的
- profile 已成功保存

### 贡献 Profile

要把 profile 添加到面向所有用户的官方列表:
1. Fork [Bootimus 仓库](https://github.com/garybowers/bootimus)
2. 编辑仓库根目录下的 `distro-profiles.json`
3. 把你的 profile 添加到 `profiles` 数组
4. 提交 pull request

这样,所有 Bootimus 用户都能通过 "Check for Updates" 获得新 profile。
