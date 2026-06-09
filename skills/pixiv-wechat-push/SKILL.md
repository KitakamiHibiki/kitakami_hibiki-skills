# pixiv-wechat-push

Fetch Pixiv daily recommendations and publish them as a WeChat Official Account article.

## Overview

This skill automates the full workflow: fetch Pixiv trending illustrations → generate a WeChat article with images → upload draft → publish.

## Prerequisites

Both CLI tools must be configured before running this skill:

### 1. Pixiv authentication (PHPSESSID)

```bash
pixiv login
```

### 2. WeChat Official Account credentials

```bash
wechat login
```

## Workflow

The skill orchestrates the following steps:

### Step 1 — Fetch Pixiv Daily Recommendations

```bash
pixiv recommand -mode daily -limit ${LIMIT} -r18 0
```

Default `LIMIT=10`. Images are downloaded to `~/.kitakami_hibiki/pixiv/recommand/` as `{illustId}_000.jpg`.

### Step 2 — Parse Results

From the CLI output, extract for each illustration:
- **ID** (used for filename lookup and artwork URL)
- **Title** (article body heading)
- **Artist name**
- **Tags** (comma-separated)
- **Views / Bookmarks** (engagement metrics)
- **Pixiv URL** (`https://www.pixiv.net/artworks/{id}`)

### Step 3 — Upload Images to WeChat

For each illustration's first page image, upload to WeChat CDN:

```bash
wechat media upload "~/.kitakami_hibiki/pixiv/recommand/{id}_000.jpg"
```

Capture the returned URL — it will be used in the article HTML.

> If the image file doesn't exist (e.g. download failed), skip that illustration.

### Step 4 — Generate Article HTML

Build a WeChat-friendly HTML article with the following structure:

- **Title**: `Pixiv 今日推荐 — YYYY-MM-DD`
- **Author**: `Pixiv Daily`
- **Body**: For each illustration (up to `LIMIT`), render a section with:
  - The uploaded WeChat image (full-width, using `<img src="...">`)
  - Illustration title as heading
  - Artist name, tags, view/bookmark counts
  - A "查看原图" link pointing to the Pixiv artwork URL
- Wrap `<img>` tags with `<span style="width:100%;display:block;">` for responsive layout
- Use `<p>` for text content

### Step 5 — Create Draft JSON

Write a JSON file (e.g. `~/.kitakami_hibiki/temp/pixiv_daily.json`):

```json
{
  "articles": [{
    "title": "Pixiv 今日推荐 — YYYY-MM-DD",
    "author": "Pixiv Daily",
    "content": "<html content from step 4>",
    "need_open_comment": 1,
    "only_fans_can_comment": 0,
    "show_cover_pic": 1
  }]
}
```

> Note: `thumb_media_id` is omitted — WeChat will auto-generate a cover from the content.

### Step 6 — Upload Draft

```bash
wechat draft create ~/.kitakami_hibiki/temp/pixiv_daily.json
```

Capture the returned `media_id`.

### Step 7 — Publish

```bash
wechat publish submit ${MEDIA_ID}
```

### Step 8 — Cleanup

Remove the temporary JSON file.

## Output

- Images downloaded to `~/.kitakami_hibiki/pixiv/recommand/`
- A WeChat article published to the configured Official Account
- Console shows the publish_id for tracking

## Error Handling

| Step | Error | Likely Cause | Action |
|------|-------|-------------|--------|
| 1 | `no credentials found` | Pixiv PHPSESSID not set | Run `pixiv login` |
| 1 | `fetch failed` | Network / proxy issue | Check network; verify proxy config |
| 3 | `file not found` | Image not downloaded | Image may have failed to download; skip |
| 3 | `upload failed` | WeChat credentials invalid | Run `wechat login` |
| 3 | `upload failed` | Invalid file type | WeChat requires JPG/PNG; check file |
| 6 | `draft create failed` | Article content too long | Reduce LIMIT or shorten article |
| 7 | `publish failed` | Draft ID invalid | Verify the media_id was captured correctly |

## Implementation Notes

- Image files larger than 10MB may fail WeChat upload — the pixiv skill downloads originals which could be large
- WeChat article content has a character limit (~20,000 bytes for the HTML body); keep sections concise
- The uploaded WeChat image URLs are temporary; they should be used in the article immediately within the same session
- Tags are displayed as plain text in the article (WeChat doesn't support clickable hashtags)
