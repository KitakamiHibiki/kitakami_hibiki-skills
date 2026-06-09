# Skills

A modular CLI toolkit composed of reusable capabilities ("skills"). Each skill is
a self-contained feature documented in `skills/`, with the corresponding CLI
implementation under `bin/`.

## Project Structure

```
├── bin/              # Go CLI tools
│   ├── pixiv/        # Pixiv illustration CLI
│   ├── wechat/       # WeChat Official Account CLI
│   ├── internal/     # Shared packages (config, db, pixiv client, wechat client, proxy)
│   ├── script/       # Build scripts
│   └── go.mod        # Go module
├── skills/           # Skill documentation (one directory per skill)
│   ├── pixiv-search/
│   └── pixiv-wechat-push/
└── .github/          # CI workflows
```

## Prerequisites

- Go 1.23+

## Build

Build all CLI tools as standalone executables:

```bash
cd bin

# Build all tools at once
go build ./pixiv/      # produces pixiv (or pixiv.exe)
go build ./wechat/     # produces wechat (or wechat.exe)
```

Or use the build scripts (output goes to `bin/build/`):

```bash
# Windows
script\build.bat

# Linux / macOS
bash script/build.sh
```

## PATH Setup

After building, add the output directory to your `PATH` so the CLI commands
are available from anywhere:

**Windows (cmd):**
```cmd
set PATH=%PATH%;D:\Dongby\Program\AI\Skills\bin\build
```
Or add the path permanently via System Environment Variables.

**Linux / macOS:**
```bash
export PATH=$PATH:/path/to/Skills/bin/build
```
Add the line to `~/.bashrc` or `~/.zshrc` to make it permanent.

Once `pixiv` and `wechat` are on PATH, all skill workflows work without
hardcoded paths.

## Usage

### pixiv

Search and download illustrations from Pixiv using the web API.

```bash
# Login (save PHPSESSID)
pixiv login

# Fetch recommended illustrations
pixiv recommand -limit 10
```

Flags:
| Flag   | Default | Description                           |
| ------ | ------- | ------------------------------------- |
| -limit | 10      | Number of illustrations to fetch      |
| -r18   | 0       | 0=exclude, 1=include all, 2=only R18 |
| -mode  | daily   | daily, weekly, monthly, random        |

### wechat

Manage WeChat Official Account drafts and articles.

```bash
# Configure credentials
wechat login

# List drafts
wechat draft list

# Create draft from JSON
wechat draft create article.json

# Publish draft
wechat publish submit <media_id>

# Upload image
wechat media upload image.jpg
```

## Skills

Each directory under `skills/` documents a specific capability provided by the CLI.

| Skill              | Description                                                   |
|--------------------|---------------------------------------------------------------|
| install            | Build and install CLI tools, configure PATH                   |
| pixiv-search       | Search and download Pixiv illustrations (PHPSESSID auth)      |
| pixiv-wechat-push  | Fetch Pixiv daily picks → generate → publish WeChat article   |
