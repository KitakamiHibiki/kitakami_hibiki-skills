# pixiv-search

Search and download illustrations from Pixiv using the pixiv web API with PHPSESSID cookie authentication.

## Usage

Browse or download Pixiv illustrations. Use this when you need to search, view metadata, or download images from Pixiv.

### Prerequisites

The user needs a **PHPSESSID cookie** from their Pixiv login session:
1. Open https://www.pixiv.net/ in browser and log in
2. Open DevTools (F12) → Application → Cookies → https://www.pixiv.net
3. Copy the value of the `PHPSESSID` cookie

### Authentication

Before first use, ask the user to provide their PHPSESSID:

```bash
pixiv login
```

The user will be prompted to paste their PHPSESSID. The tool verifies the session and saves it to the SQLite database (`~/.kitakami_hibiki/data.db`).

Alternatively, set the `PIXIV_PHPSESSID` environment variable.

### Fetching and Downloading Recommended Illustrations

```bash
pixiv recommand [flags]
```

Flags:
| Flag | Default | Description |
|------|---------|-------------|
| `-limit` | `10` | Number of illustrations to fetch |
| `-r18` | `0` | R18 filter: `0`=exclude, `1`=include all, `2`=only R18 |
| `-mode` | `daily` | Ranking mode: daily, weekly, monthly, random |

### Output

- Displays illustration metadata (title, artist, size, tags, etc.)
- Downloads high-resolution original images to `~/.kitakami_hibiki/pixiv/recommand/`
- Files are named `{illustId}_{pageNumber}.{ext}` (page numbers are zero-padded, e.g. `12345678_000.jpg`)

## Implementation Details

- Uses Pixiv web API (`www.pixiv.net/ajax/...`) — no OAuth involved
- Authentication via `PHPSESSID` cookie
- Image URLs fetched from `/ajax/illust/{id}/pages` endpoint
- Referer header set to `https://www.pixiv.net/` for hotlink protection

## Error Handling

| Error | Cause | Resolution |
|-------|-------|------------|
| `login failed` | PHPSESSID expired or invalid | User needs to re-login in browser and provide fresh cookie |
| `fetch failed` | API request failed | Check network/proxy settings; session may have expired |
| `download failed` | Image URL inaccessible | Usually means PHPSESSID expired — re-authenticate |
| `no matching illustrations found` | R18 filter too restrictive | Try with `-r18 1` to include all |
