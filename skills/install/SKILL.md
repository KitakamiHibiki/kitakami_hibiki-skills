# install

Build and install the CLI tools (`pixiv`, `wechat`) from source, then configure the system PATH.

## Overview

This skill compiles the Go source code into standalone executables and makes them
available system-wide. After installation, the `pixiv` and `wechat` commands are
accessible from any directory.

## Prerequisites

- Go 1.23+ installed and `go` available on PATH
- Git (to clone or pull the latest source)

## Workflow

### Step 1 — Clone or Pull Latest Source

```bash
git clone <repo-url>          # first time
git pull                      # subsequent updates
```

Then change to the project root and enter the `bin/` directory where `go.mod` lives.

### Step 2 — Build Binaries

Build all CLI tools:

```bash
cd bin
go mod tidy
go build ./pixiv/
go build ./wechat/
```

This produces:
- `pixiv` (or `pixiv.exe` on Windows)
- `wechat` (or `wechat.exe` on Windows)

in the current directory.

Alternatively, use the build script which outputs to `bin/build/` and
cross-compiles for both Windows and Linux:

```bash
# Windows
bin\script\build.bat

# Linux / macOS
bash bin/script/build.sh
```

### Step 3 — Install to PATH

Copy the binaries to a directory that's on the system PATH.

**Option A — User-local bin directory** (recommended):

```bash
# Linux / macOS
mkdir -p ~/.local/bin
cp bin/build/* ~/.local/bin/
# Verify ~/.local/bin is in PATH; add to ~/.bashrc if not:
# export PATH="$HOME/.local/bin:$PATH"
```

```cmd
:: Windows
mkdir %USERPROFILE%\.local\bin
copy bin\build\*.exe %USERPROFILE%\.local\bin\
:: Add %USERPROFILE%\.local\bin to PATH via System Environment Variables
```

**Option B — System-wide** (requires admin / sudo):

```bash
# Linux / macOS
sudo cp bin/build/pixiv bin/build/wechat /usr/local/bin/
```

```cmd
:: Windows (run as Administrator)
copy bin\build\pixiv.exe bin\build\wechat.exe C:\Windows\System32\
```

**Option C — Keep in-place and add to PATH**:

```bash
# Add the build output directory directly to PATH:
export PATH="$PATH:/path/to/Skills/bin/build"
```

```cmd
:: Windows
set PATH=%PATH%;D:\path\to\Skills\bin\build
```

### Step 4 — Verify Installation

```bash
pixiv help
wechat help
```

Both commands should print their usage information.

## Post-Installation

After installation, configure credentials:

```bash
# Pixiv (requires PHPSESSID from browser)
pixiv login

# WeChat Official Account (requires AppID and AppSecret)
wechat login
```

## Error Handling

| Step | Error | Likely Cause | Resolution |
|------|-------|-------------|------------|
| 2 | `go: command not found` | Go not installed | Install Go 1.23+ from https://go.dev |
| 2 | build fails | Outdated dependencies | Run `go mod tidy` first |
| 3 | `cp: permission denied` | Insufficient privileges | Use Option A (user-local) instead of system-wide |
| 4 | `command not found` | Install dir not on PATH | Add the directory to PATH and restart shell |
