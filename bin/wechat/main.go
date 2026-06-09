package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"skills/bin/internal/db"
	"skills/bin/internal/version"
	"skills/bin/internal/wechat"
)

func main() {
	if _, err := db.Open(); err != nil {
		fmt.Fprintf(os.Stderr, "error initializing database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	verbose := false
	for _, a := range os.Args {
		if a == "-v" || a == "--verbose" {
			verbose = true
		}
		if a == "--version" {
			fmt.Println(version.Info())
			return
		}
	}

	if len(os.Args) < 2 {
		printUsage()
		return
	}

	subcommand := os.Args[1]
	subArgs := os.Args[2:]

	// Filter out -v/--verbose for subcommand parsing.
	filtered := make([]string, 0, len(subArgs))
	for _, a := range subArgs {
		if a != "-v" && a != "--verbose" {
			filtered = append(filtered, a)
		}
	}

	switch subcommand {
	case "login":
		runLogin(verbose, filtered)
	case "draft":
		runDraft(verbose, filtered)
	case "publish":
		runPublish(verbose, filtered)
	case "media":
		runMedia(verbose, filtered)
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "unknown subcommand: %q\n", subcommand)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage: wechat <subcommand> [flags]")
	fmt.Println()
	fmt.Println("Subcommands:")
	fmt.Println("  login                        configure AppID & AppSecret")
	fmt.Println("  draft list                   list drafts")
	fmt.Println("  draft create <file.json>     create draft from JSON")
	fmt.Println("  draft show <media_id>        show draft details")
	fmt.Println("  draft update <id> <file>     update draft")
	fmt.Println("  draft delete <media_id>      delete draft")
	fmt.Println("  publish submit <media_id>    submit draft for publishing")
	fmt.Println("  publish list                 list published articles")
	fmt.Println("  media upload <file>          upload image, returns URL")
	fmt.Println()
	fmt.Println("Global flags:")
	fmt.Println("  -v, --verbose                verbose output")
	fmt.Println("  --version                    show version")
	fmt.Println()
	fmt.Println("Config:")
	fmt.Println("  Credentials are stored in SQLite (~/.kitakami_hibiki/data.db)")
	fmt.Println("  Run `wechat login` to set up AppID and AppSecret")
}

// --- login ---

func runLogin(verbose bool, args []string) {
	fs := flag.NewFlagSet("login", flag.ContinueOnError)
	fs.Parse(args)

	fmt.Print("AppID: ")
	var appID string
	fmt.Scanln(&appID)
	appID = strings.TrimSpace(appID)

	fmt.Print("AppSecret: ")
	var appSecret string
	fmt.Scanln(&appSecret)
	appSecret = strings.TrimSpace(appSecret)

	if appID == "" || appSecret == "" {
		fmt.Fprintln(os.Stderr, "error: AppID and AppSecret are required")
		os.Exit(1)
	}

	if err := db.ConfigSet("wechat_appid", appID); err != nil {
		fmt.Fprintf(os.Stderr, "save appid failed: %v\n", err)
		os.Exit(1)
	}
	if err := db.ConfigSet("wechat_appsecret", appSecret); err != nil {
		fmt.Fprintf(os.Stderr, "save appsecret failed: %v\n", err)
		os.Exit(1)
	}

	// Verify by fetching token.
	if verbose {
		fmt.Println("verifying credentials...")
	}
	client := wechat.NewClient()
	_, err := client.ListDrafts(0, 1)
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: credential verification failed: %v\n", err)
		fmt.Println("credentials saved but may be invalid")
		return
	}

	fmt.Println("login successful, credentials verified")
}

// --- draft ---

func runDraft(verbose bool, args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "usage: wechat draft <list|create|show|update|delete> [args]")
		os.Exit(1)
	}

	client := wechat.NewClient()

	switch args[0] {
	case "list":
		runDraftList(verbose, client, args[1:])
	case "create":
		runDraftCreate(verbose, client, args[1:])
	case "show":
		runDraftShow(verbose, client, args[1:])
	case "update":
		runDraftUpdate(verbose, client, args[1:])
	case "delete":
		runDraftDelete(verbose, client, args[1:])
	default:
		fmt.Fprintf(os.Stderr, "unknown draft subcommand: %q\n", args[0])
		os.Exit(1)
	}
}

func runDraftList(verbose bool, client *wechat.Client, args []string) {
	fs := flag.NewFlagSet("draft list", flag.ContinueOnError)
	offset := fs.Int("offset", 0, "pagination offset")
	count := fs.Int("count", 10, "number of drafts to fetch (max 20)")
	fs.Parse(args)

	resp, err := client.ListDrafts(*offset, *count)
	if err != nil {
		fmt.Fprintf(os.Stderr, "list drafts failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Total drafts: %d\n\n", resp.TotalCount)
	for _, item := range resp.Items {
		article := item.Content.NewsItem[0]
		t := time.Unix(item.UpdateTime, 0).Format("2006-01-02 15:04")
		fmt.Printf("  %s\n", item.MediaID)
		fmt.Printf("    Title:  %s\n", article.Title)
		fmt.Printf("    Author: %s\n", article.Author)
		fmt.Printf("    Digest: %s\n", article.Digest)
		fmt.Printf("    Updated: %s\n", t)
		fmt.Println()
	}
}

func runDraftCreate(verbose bool, client *wechat.Client, args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "usage: wechat draft create <file.json>")
		os.Exit(1)
	}

	articles, err := loadDraftFile(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "load draft file: %v\n", err)
		os.Exit(1)
	}

	if verbose {
		fmt.Printf("creating draft with %d article(s)...\n", len(articles))
	}

	mediaID, err := client.CreateDraft(articles)
	if err != nil {
		fmt.Fprintf(os.Stderr, "create draft failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("draft created: %s\n", mediaID)
}

func runDraftShow(verbose bool, client *wechat.Client, args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "usage: wechat draft show <media_id>")
		os.Exit(1)
	}

	mediaID := args[0]
	resp, err := client.GetDraft(mediaID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "get draft failed: %v\n", err)
		os.Exit(1)
	}

	for i, article := range resp.NewsItem {
		if i > 0 {
			fmt.Println("---")
		}
		fmt.Printf("Article #%d:\n", i+1)
		fmt.Printf("  Title:              %s\n", article.Title)
		fmt.Printf("  Author:             %s\n", article.Author)
		fmt.Printf("  Digest:             %s\n", article.Digest)
		fmt.Printf("  Content URL:        %s\n", article.ContentSourceURL)
		fmt.Printf("  Thumb Media ID:     %s\n", article.ThumbMediaID)
		fmt.Printf("  Show Cover:         %d\n", article.ShowCover)
		fmt.Printf("  Open Comment:       %d\n", article.NeedOpenComment)
		fmt.Printf("  Fans Can Comment:   %d\n", article.OnlyFansCanComment)
		fmt.Printf("  URL:                %s\n", article.URL)
		fmt.Printf("  Content length:     %d chars\n", len(article.Content))
	}
}

func runDraftUpdate(verbose bool, client *wechat.Client, args []string) {
	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: wechat draft update <media_id> <file.json>")
		os.Exit(1)
	}

	mediaID := args[0]
	articles, err := loadDraftFile(args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "load draft file: %v\n", err)
		os.Exit(1)
	}

	if verbose {
		fmt.Printf("updating draft %s...\n", mediaID)
	}

	if err := client.UpdateDraft(mediaID, 0, articles); err != nil {
		fmt.Fprintf(os.Stderr, "update draft failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("draft updated")
}

func runDraftDelete(verbose bool, client *wechat.Client, args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "usage: wechat draft delete <media_id>")
		os.Exit(1)
	}

	mediaID := args[0]
	if verbose {
		fmt.Printf("deleting draft %s...\n", mediaID)
	}

	if err := client.DeleteDraft(mediaID); err != nil {
		fmt.Fprintf(os.Stderr, "delete draft failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("draft deleted")
}

// --- publish ---

func runPublish(verbose bool, args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "usage: wechat publish <submit|list> [args]")
		os.Exit(1)
	}

	client := wechat.NewClient()

	switch args[0] {
	case "submit":
		runPublishSubmit(verbose, client, args[1:])
	case "list":
		runPublishList(verbose, client, args[1:])
	default:
		fmt.Fprintf(os.Stderr, "unknown publish subcommand: %q\n", args[0])
		os.Exit(1)
	}
}

func runPublishSubmit(verbose bool, client *wechat.Client, args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "usage: wechat publish submit <media_id>")
		os.Exit(1)
	}

	mediaID := args[0]
	if verbose {
		fmt.Printf("submitting draft %s for publishing...\n", mediaID)
	}

	publishID, err := client.SubmitPublish(mediaID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "publish failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("publish submitted, publish_id: %s\n", publishID)
	fmt.Println("note: publishing is asynchronous, check wechat backend for status")
}

func runPublishList(verbose bool, client *wechat.Client, args []string) {
	fs := flag.NewFlagSet("publish list", flag.ContinueOnError)
	offset := fs.Int("offset", 0, "pagination offset")
	count := fs.Int("count", 10, "number of articles to fetch (max 20)")
	fs.Parse(args)

	resp, err := client.ListPublished(*offset, *count)
	if err != nil {
		fmt.Fprintf(os.Stderr, "list published failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Total published: %d\n\n", resp.TotalCount)
	for _, item := range resp.Items {
		t := time.Unix(item.UpdateTime, 0).Format("2006-01-02 15:04")
		fmt.Printf("  %s\n", item.ArticleID)
		fmt.Printf("    Title:  %s\n", item.Article.Title)
		fmt.Printf("    Digest: %s\n", item.Article.Digest)
		fmt.Printf("    URL:    %s\n", item.Article.URL)
		fmt.Printf("    Published: %s\n", t)
		fmt.Println()
	}
}

// --- media ---

func runMedia(verbose bool, args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "usage: wechat media upload <file>")
		os.Exit(1)
	}

	client := wechat.NewClient()

	switch args[0] {
	case "upload":
		runMediaUpload(verbose, client, args[1:])
	default:
		fmt.Fprintf(os.Stderr, "unknown media subcommand: %q\n", args[0])
		os.Exit(1)
	}
}

func runMediaUpload(verbose bool, client *wechat.Client, args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "usage: wechat media upload <file>")
		os.Exit(1)
	}

	filePath := args[0]
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "file not found: %s\n", filePath)
		os.Exit(1)
	}

	if verbose {
		fmt.Printf("uploading %s...\n", filePath)
	}

	url, err := client.UploadImage(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "upload failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("image URL: %s\n", url)
}

// --- helpers ---

// loadDraftFile 从 JSON 文件加载 DraftArticle 列表。
// 支持顶层 articles 数组或直接传数组。
func loadDraftFile(filePath string) ([]wechat.DraftArticle, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	// Try { "articles": [...] } format.
	var wrapper struct {
		Articles []wechat.DraftArticle `json:"articles"`
	}
	if err := json.Unmarshal(data, &wrapper); err == nil && len(wrapper.Articles) > 0 {
		return wrapper.Articles, nil
	}

	// Try direct array format.
	var articles []wechat.DraftArticle
	if err := json.Unmarshal(data, &articles); err != nil {
		return nil, fmt.Errorf("parse JSON: %w (expected array or {articles: [...]})", err)
	}
	return articles, nil
}
