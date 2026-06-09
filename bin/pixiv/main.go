package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	"skills/bin/internal/config"
	"skills/bin/internal/db"
	"skills/bin/internal/pixiv"
	"skills/bin/internal/version"
)

func main() {
	// 初始化 SQLite 数据库。
	if _, err := db.Open(); err != nil {
		fmt.Fprintf(os.Stderr, "error initializing database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	// Pre-scan: check for -v / --verbose / --version anywhere in args.
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

	switch os.Args[1] {
	case "login":
		runLogin(verbose, os.Args[2:])
	case "recommand":
		runRecommand(verbose, os.Args[2:])
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "unknown subcommand: %q\n", os.Args[1])
		os.Exit(1)
	}
}

func runLogin(verbose bool, args []string) {
	_ = args // reserved for future flags

	// Check if already logged in.
	cfg, err := pixiv.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	if cfg != nil && cfg.PHPSESSID != "" {
		fmt.Println("already logged in (PHPSESSID saved)")
		fmt.Println("run `pixiv login` again to re-authenticate")
		return
	}

	fmt.Println("To log in to Pixiv:")
	fmt.Println()
	fmt.Println("  1. Open https://www.pixiv.net/ in your browser and log in")
	fmt.Println()
	fmt.Println("  2. Open Developer Tools (F12) → Application (or 存储) → Cookies")
	fmt.Println("     → https://www.pixiv.net")
	fmt.Println()
	fmt.Println("  3. Find the cookie named 'PHPSESSID' and copy its value")
	fmt.Println()
	fmt.Println("Paste the PHPSESSID value below and press Enter:")

	reader := bufio.NewReader(os.Stdin)
	phpsessid, err := reader.ReadString('\n')
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading input: %v\n", err)
		os.Exit(1)
	}
	phpsessid = strings.TrimSpace(phpsessid)

	if phpsessid == "" {
		fmt.Fprintln(os.Stderr, "error: PHPSESSID is required")
		os.Exit(1)
	}

	if verbose {
		fmt.Println("verifying PHPSESSID with Pixiv...")
	}

	client := pixiv.NewClient()
	if verbose {
		pixiv.SetVerbose(true)
	}
	client.SetPHPSESSID(phpsessid)

	userID, userName, err := client.VerifySession()
	if err != nil {
		fmt.Fprintf(os.Stderr, "login failed: %v\n", err)
		fmt.Fprintln(os.Stderr, "make sure the PHPSESSID is correct and not expired")
		os.Exit(1)
	}

	if err := pixiv.SaveConfig(&pixiv.Config{PHPSESSID: phpsessid}); err != nil {
		fmt.Fprintf(os.Stderr, "save config failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("login successful as %s (user ID: %d)\n", userName, userID)
}

func runRecommand(verbose bool, args []string) {
	fs := flag.NewFlagSet("recommand", flag.ContinueOnError)
	limit := fs.Int("limit", 10, " number of illustrations to show")
	r18 := fs.Int("r18", 0, " R18 filter: 0=exclude, 1=include, 2=only R18")
	mode := fs.String("mode", "daily", " ranking mode: daily, weekly, monthly, random")

	// Filter out -v/--verbose so flag.Parse doesn't reject them.
	filtered := make([]string, 0, len(args))
	for _, a := range args {
		if a != "-v" && a != "--verbose" {
			filtered = append(filtered, a)
		}
	}
	fs.Parse(filtered)

	phpsessid := os.Getenv("PIXIV_PHPSESSID")

	// Fall back to config file if env var is not set.
	if phpsessid == "" {
		cfg, err := pixiv.LoadConfig()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		if cfg == nil || cfg.PHPSESSID == "" {
			fmt.Fprintln(os.Stderr, "error: no credentials found. Run `pixiv login` first or set PIXIV_PHPSESSID")
			os.Exit(1)
		}
		phpsessid = cfg.PHPSESSID
	}

	client := pixiv.NewClient()
	client.SetPHPSESSID(phpsessid)

	if verbose {
		pixiv.SetVerbose(true)
		fmt.Printf("fetching %s ranking...\n", *mode)
	}

	result, err := client.FetchRanking(*mode)
	if err != nil {
		fmt.Fprintf(os.Stderr, "fetch failed: %v\n", err)
		os.Exit(1)
	}

	if len(result.Illusts) == 0 {
		fmt.Println("no recommended illustrations found.")
		return
	}

	// Filter R18 content based on --r18 value.
	illusts := result.Illusts
	switch *r18 {
	case 0: // exclude R18
		filtered := make([]pixiv.Illust, 0, len(illusts))
		for _, ill := range illusts {
			if ill.XRestrict == 0 {
				filtered = append(filtered, ill)
			}
		}
		illusts = filtered
	case 2: // only R18
		filtered := make([]pixiv.Illust, 0, len(illusts))
		for _, ill := range illusts {
			if ill.XRestrict > 0 {
				filtered = append(filtered, ill)
			}
		}
		illusts = filtered
	}
	// case 1: include all, no filtering

	if len(illusts) == 0 {
		fmt.Println("no matching illustrations found.")
		return
	}

	count := *limit
	if count > len(illusts) {
		count = len(illusts)
	}

	// Prepare download directory.
	downloadDir, err := config.DownloadDir("recommand")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Recommended Illustrations (%d shown):\n\n", count)
	for i := 0; i < count; i++ {
		illust := illusts[i]
		fmt.Print(pixiv.FormatIllust(illust))

		// Download all pages.
		saved, err := client.DownloadIllust(illust, downloadDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  ! download failed: %v\n\n", err)
		} else {
			for _, p := range saved {
				fmt.Printf("  Saved: %s\n", p)
			}
			fmt.Println()
		}

		// Record download to database.
		tags := make([]string, len(illust.Tags))
		for j, t := range illust.Tags {
			tags[j] = t.Name
		}
		if recErr := db.InsertDownload(db.PixivDownload{
			ID:             illust.ID,
			Title:          illust.Title,
			Type:           illust.Type,
			XRestrict:      illust.XRestrict,
			Caption:        illust.Caption,
			Width:          illust.Width,
			Height:         illust.Height,
			PageCount:      illust.PageCount,
			TotalBookmarks: illust.TotalBookmarks,
			TotalView:      illust.TotalView,
			ArtistID:       illust.User.ID,
			ArtistName:     illust.User.Name,
			Tags:           strings.Join(tags, ","),
		}); recErr != nil && verbose {
			fmt.Printf("  ! failed to record download: %v\n", recErr)
		}
	}
}

func printUsage() {
	fmt.Println("Usage: pixiv <subcommand> [flags]")
	fmt.Println()
	fmt.Println("Subcommands:")
	fmt.Println("  login        save Pixiv PHPSESSID cookie")
	fmt.Println("  recommand    fetch recommended illustrations")
	fmt.Println("  help         show this help")
	fmt.Println()
	fmt.Println("Flags (recommand):")
	fmt.Println("  -mode        ranking mode: daily, weekly, monthly, random (default \"daily\")")
	fmt.Println("  -limit       number of illustrations to show (default 10)")
	fmt.Println("  -r18         0=exclude R18, 1=include all, 2=only R18 (default 0)")
	fmt.Println()
	fmt.Println("Global flags:")
	fmt.Println("  -v, --verbose              verbose output")
	fmt.Println("  --version                  show version")
	fmt.Println()
	fmt.Println("Config:")
	fmt.Println("  Credentials are saved to SQLite database (~/.kitakami_hibiki/data.db)")
	fmt.Println("  Environment variable PIXIV_PHPSESSID takes priority over config")
	fmt.Println()
	fmt.Println("How to get PHPSESSID:")
	fmt.Println("  1. Log in to https://www.pixiv.net/ in your browser")
	fmt.Println("  2. Open DevTools (F12) -> Application -> Cookies -> pixiv.net")
	fmt.Println("  3. Copy the value of the 'PHPSESSID' cookie")
}
