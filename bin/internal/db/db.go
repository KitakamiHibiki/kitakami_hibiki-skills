package db

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"sync"

	"skills/bin/internal/config"

	_ "modernc.org/sqlite"
)

var (
	pool *sql.DB
	once sync.Once
)

// Open 打开（或创建）SQLite 数据库，返回连接池。
// 数据库文件位于 ~/.kitakami_hibiki/data.db。
func Open() (*sql.DB, error) {
	var err error
	once.Do(func() {
		dir, e := config.AppDir()
		if e != nil {
			err = fmt.Errorf("get app dir: %w", e)
			return
		}
		dbPath := filepath.Join(dir, "data.db")

		db, e := sql.Open("sqlite", dbPath)
		if e != nil {
			err = fmt.Errorf("open db: %w", e)
			return
		}

		// Apply performance / safety pragmas.
		for _, pragma := range []string{
			"PRAGMA journal_mode=WAL",
			"PRAGMA busy_timeout=5000",
			"PRAGMA foreign_keys=ON",
		} {
			if _, e := db.Exec(pragma); e != nil {
				db.Close()
				err = fmt.Errorf("set pragma %q: %w", pragma, e)
				return
			}
		}

		db.SetMaxOpenConns(1) // SQLite only supports one writer at a time
		pool = db

		if e := migrate(db); e != nil {
			db.Close()
			pool = nil
			err = fmt.Errorf("migrate: %w", e)
			return
		}
	})
	return pool, err
}

// Close 关闭数据库连接池。
func Close() error {
	if pool != nil {
		return pool.Close()
	}
	return nil
}

// migrate 执行数据库 schema 迁移。
func migrate(db *sql.DB) error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS schema_version (
			version INTEGER PRIMARY KEY
		)`,

		`CREATE TABLE IF NOT EXISTS config (
			key   TEXT PRIMARY KEY,
			value TEXT NOT NULL
		)`,

		`CREATE TABLE IF NOT EXISTS pixiv_downloads (
			id              INTEGER PRIMARY KEY,
			title           TEXT    NOT NULL,
			type            TEXT    NOT NULL DEFAULT 'illust',
			x_restrict      INTEGER NOT NULL DEFAULT 0,
			caption         TEXT    NOT NULL DEFAULT '',
			width           INTEGER NOT NULL DEFAULT 0,
			height          INTEGER NOT NULL DEFAULT 0,
			page_count      INTEGER NOT NULL DEFAULT 1,
			total_bookmarks INTEGER NOT NULL DEFAULT 0,
			total_view      INTEGER NOT NULL DEFAULT 0,
			artist_id       INTEGER NOT NULL DEFAULT 0,
			artist_name     TEXT    NOT NULL DEFAULT '',
			tags            TEXT    NOT NULL DEFAULT '',
			downloaded_at   TEXT    NOT NULL DEFAULT (datetime('now'))
		)`,

		`CREATE TABLE IF NOT EXISTS wechat_drafts (
			media_id    TEXT PRIMARY KEY,
			title       TEXT NOT NULL,
			author      TEXT NOT NULL DEFAULT '',
			digest      TEXT NOT NULL DEFAULT '',
			create_time INTEGER NOT NULL DEFAULT 0,
			update_time INTEGER NOT NULL DEFAULT 0,
			state       TEXT NOT NULL DEFAULT 'draft'
		)`,
	}

	for _, stmt := range statements {
		if _, err := db.Exec(stmt); err != nil {
			return fmt.Errorf("exec %q: %w", stmt[:40], err)
		}
	}

	// Seed initial schema version if empty.
	var count int
	db.QueryRow("SELECT COUNT(*) FROM schema_version").Scan(&count)
	if count == 0 {
		db.Exec("INSERT INTO schema_version (version) VALUES (1)")
	}

	return nil
}

// --- Config key-value helpers ---

// ConfigGet 读取配置项。key 不存在时返回 ("", nil)。
func ConfigGet(key string) (string, error) {
	if pool == nil {
		return "", fmt.Errorf("db not opened")
	}
	var value string
	err := pool.QueryRow("SELECT value FROM config WHERE key = ?", key).Scan(&value)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return value, err
}

// ConfigSet 写入配置项。
func ConfigSet(key, value string) error {
	if pool == nil {
		return fmt.Errorf("db not opened")
	}
	_, err := pool.Exec(
		"INSERT INTO config (key, value) VALUES (?, ?) ON CONFLICT(key) DO UPDATE SET value = excluded.value",
		key, value,
	)
	return err
}

// ConfigDelete 删除配置项。
func ConfigDelete(key string) error {
	if pool == nil {
		return fmt.Errorf("db not opened")
	}
	_, err := pool.Exec("DELETE FROM config WHERE key = ?", key)
	return err
}

// --- Pixiv download history ---

// PixivDownload 表示一条下载记录。
type PixivDownload struct {
	ID             int
	Title          string
	Type           string
	XRestrict      int
	Caption        string
	Width          int
	Height         int
	PageCount      int
	TotalBookmarks int
	TotalView      int
	ArtistID       int
	ArtistName     string
	Tags           string
}

// InsertDownload 插入一条 pixiv 作品下载记录（upsert by id）。
func InsertDownload(d PixivDownload) error {
	if pool == nil {
		return fmt.Errorf("db not opened")
	}
	_, err := pool.Exec(`
		INSERT INTO pixiv_downloads
			(id, title, type, x_restrict, caption, width, height,
			 page_count, total_bookmarks, total_view,
			 artist_id, artist_name, tags)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			downloaded_at = datetime('now')
	`, d.ID, d.Title, d.Type, d.XRestrict, d.Caption, d.Width, d.Height,
		d.PageCount, d.TotalBookmarks, d.TotalView,
		d.ArtistID, d.ArtistName, d.Tags)
	return err
}
