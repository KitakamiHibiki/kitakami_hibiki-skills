# update

Update the CLI tools (`pixiv`, `wechat`) to the latest GitHub Release.

## Overview

Downloads the latest pre-built binaries from GitHub Releases and replaces the
currently installed executables. No Go toolchain required.

## Quick Update

### Linux / macOS

```bash
INSTALL_DIR="${HOME}/.local/bin"

curl -sL https://github.com/KitakamiHibiki/kitakami_hibiki-skills/releases/latest/download/linux.zip \
  -o /tmp/skills.zip

unzip -o /tmp/skills.zip -d /tmp/skills/
cp /tmp/skills/linux/* "$INSTALL_DIR/"
rm -rf /tmp/skills.zip /tmp/skills/

# Verify
pixiv --version
wechat --version
```

### Windows (PowerShell)

```powershell
$installDir = "$env:USERPROFILE\.local\bin"

Invoke-WebRequest -Uri "https://github.com/KitakamiHibiki/kitakami_hibiki-skills/releases/latest/download/windows.zip" -OutFile "$env:TEMP\skills.zip"

Expand-Archive -Path "$env:TEMP\skills.zip" -DestinationPath "$env:TEMP\skills" -Force
Copy-Item "$env:TEMP\skills\windows\*.exe" -Destination $installDir
Remove-Item -Recurse -Force "$env:TEMP\skills.zip", "$env:TEMP\skills"

# Verify
pixiv --version
wechat --version
```

## One-liner

**Linux / macOS:**
```bash
curl -sL https://github.com/KitakamiHibiki/kitakami_hibiki-skills/releases/latest/download/linux.zip -o /tmp/s.zip && unzip -o /tmp/s.zip -d /tmp/s/ && cp /tmp/s/linux/* ~/.local/bin/ && rm -rf /tmp/s.zip /tmp/s/ && pixiv --version
```

**Windows (PowerShell):**
```powershell
iwr -Uri https://github.com/KitakamiHibiki/kitakami_hibiki-skills/releases/latest/download/windows.zip -OutFile $env:TEMP\s.zip; Expand-Archive $env:TEMP\s.zip -Dest $env:TEMP\s -Force; cp $env:TEMP\s\windows\*.exe $env:USERPROFILE\.local\bin\; ri -Recurse -Force $env:TEMP\s.zip,$env:TEMP\s
```

## Alternative — Build from Source

If you have the source repository cloned and Go 1.23+ installed:

```bash
cd /path/to/kitakami_hibiki-skills/bin
git pull --tags
go build ./pixiv/
go build ./wechat/
cp pixiv ~/.local/bin/
cp wechat ~/.local/bin/
pixiv --version
```

## Error Handling

| Error | Likely Cause | Resolution |
|-------|-------------|------------|
| `curl: command not found` | curl not installed | Install curl, or use `wget` / `Invoke-WebRequest` |
| 404 on download URL | No releases yet | Check https://github.com/KitakamiHibiki/kitakami_hibiki-skills/releases |
| `cp: permission denied` | Insufficient privileges | Use `sudo` or install to user-local directory |
| Version unchanged | Binary not replaced | Verify `INSTALL_DIR` points to the correct PATH directory |
