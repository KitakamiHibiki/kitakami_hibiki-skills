# install

Download the pre-built CLI tools (`pixiv`, `wechat`) and skill definitions from GitHub Releases, then install them to the appropriate directories.

## Overview

This skill downloads two artifacts from the latest GitHub Release:

1. **Binary archive** — platform-specific ZIP containing `pixiv` and `wechat` CLI tools, extracted to `~/.local/bin` and added to PATH.
2. **Skills definitions** — `skills-definitions.zip` containing all skill `.md` files, extracted to the agent's own skill directory.

No Go toolchain or compiler is required.

## Prerequisites

- Internet access (to reach GitHub Releases)
- `curl` (Linux / macOS) or `curl` / `powershell` (Windows)
- `unzip` (Linux / macOS) — pre-installed on most systems
- Know your agent's skill directory path (denoted as `{SKILL_DIR}` below)

## Workflow

### Step 1 — Determine OS and Architecture

The release ZIPs follow the naming convention:
`skills-<version>-<os>-<arch>.zip`

Supported combinations:

| OS      | Arch        | Suffix               |
|---------|-------------|----------------------|
| Linux   | amd64       | `linux-amd64`        |
| Windows | amd64       | `windows-amd64`      |
| macOS   | amd64 (Intel) | `darwin-amd64`     |
| macOS   | arm64 (Apple Silicon) | `darwin-arm64` |

### Step 2 — Fetch the Latest Release

Use the GitHub API to get the latest release tag:

```bash
REPO="KitakamiHibiki/kitakami_hibiki-skills"
LATEST=$(curl -sSfL "https://api.github.com/repos/$REPO/releases/latest" \
  | grep '"tag_name":' | cut -d'"' -f4)
echo "Latest release: $LATEST"
```

### Step 3 — Download and Install CLI Tools

**Linux / macOS:**

```bash
# Auto-detect platform
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
case "$ARCH" in
  x86_64)  ARCH="amd64" ;;
  aarch64) ARCH="arm64"  ;;
esac

# Build download URL
URL="https://github.com/$REPO/releases/download/$LATEST/skills-${LATEST#v}-${OS}-${ARCH}.zip"

# Download, extract, and install
mkdir -p ~/.local/bin
curl -sSfL "$URL" -o /tmp/skills.zip
unzip -o -q /tmp/skills.zip -d ~/.local/bin/
rm /tmp/skills.zip
chmod +x ~/.local/bin/pixiv ~/.local/bin/wechat

# Add to PATH if not already present
case :$PATH: in
  *:$HOME/.local/bin:*) ;;
  *) echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
     echo "Added ~/.local/bin to PATH in ~/.bashrc — restart your shell or run:"
     echo "  export PATH=\"\$HOME/.local/bin:\$PATH\"" ;;
esac
```

**Windows (PowerShell):**

```powershell
# Detect platform
$arch = "amd64"
$os = "windows"

# Get latest release tag
$repo = "KitakamiHibiki/kitakami_hibiki-skills"
$latest = (Invoke-RestMethod "https://api.github.com/repos/$repo/releases/latest").tag_name
$ver = $latest -replace '^v', ''

# Download and extract
$url = "https://github.com/$repo/releases/download/$latest/skills-$ver-$os-$arch.zip"
$zip = "$env:TEMP\skills.zip"
Invoke-WebRequest -Uri $url -OutFile $zip

$installDir = "$env:USERPROFILE\.local\bin"
New-Item -ItemType Directory -Force -Path $installDir | Out-Null
Expand-Archive -Path $zip -DestinationPath $installDir -Force
Remove-Item $zip

# Add to PATH
$currentPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($currentPath -notlike "*$installDir*") {
  [Environment]::SetEnvironmentVariable("Path", "$currentPath;$installDir", "User")
  Write-Host "Added $installDir to user PATH — restart your terminal."
}
```

### Step 4 — Download and Install Skill Definitions

Download `skills-definitions.zip` and extract it to your agent's skill directory (`{SKILL_DIR}`). Replace `{SKILL_DIR}` with the actual path where your agent loads skills from (e.g. `~/.kitakami_hibiki/skills`, `~/.myagent/skills`, or a project-relative path).

**Linux / macOS:**

```bash
SKILL_DIR="{SKILL_DIR}"   # ← replace with your agent's skill directory

# Download skills definitions
SKILLS_URL="https://github.com/$REPO/releases/download/$LATEST/skills-definitions.zip"
curl -sSfL "$SKILLS_URL" -o /tmp/skills-definitions.zip

# Extract to skill directory
mkdir -p "$SKILL_DIR"
unzip -o -q /tmp/skills-definitions.zip -d "$SKILL_DIR"
rm /tmp/skills-definitions.zip
```

**Windows (PowerShell):**

```powershell
$skillDir = "{SKILL_DIR}"   # ← replace with your agent's skill directory

# Download skills definitions
$skillsUrl = "https://github.com/$repo/releases/download/$latest/skills-definitions.zip"
$skillsZip = "$env:TEMP\skills-definitions.zip"
Invoke-WebRequest -Uri $skillsUrl -OutFile $skillsZip

# Extract to skill directory
New-Item -ItemType Directory -Force -Path $skillDir | Out-Null
Expand-Archive -Path $skillsZip -DestinationPath $skillDir -Force
Remove-Item $skillsZip
```

### Step 5 — Verify Installation

```bash
pixiv help
wechat help
```

Both commands should print their usage information. Also verify that the skill `.md` files are present in `{SKILL_DIR}`.

## Updating

To update to the latest release, repeat Steps 2–5. New binaries will overwrite old ones in `~/.local/bin`, and new skill definitions will overwrite old ones in `{SKILL_DIR}`.

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
| 2 | `curl: command not found` | curl not installed | Install curl, or use `wget` / `Invoke-WebRequest` |
| 2 | `unzip: command not found` | unzip not installed | Install unzip (`apt install unzip`, `brew install unzip`) |
| 3 | `404` or download fails | Platform not supported | Check supported platforms in Step 1; run on amd64 Linux/Windows/macOS |
| 3 | `No releases found` | No tags pushed yet | Push a version tag (`git tag v0.1.0 && git push --tags`) to trigger a release build |
| 4 | `command not found` | Install dir not on PATH | Add the directory to PATH and restart shell |
| 4 | `404` on skills-definitions.zip | Release was built before this zip was added | Use a newer release tag |
