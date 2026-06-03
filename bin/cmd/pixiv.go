package cmd

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	"skills/bin/internal/config"
	"skills/bin/internal/pixiv"
)

func runPixiv(args []string) {
	fs := flag.NewFlagSet("pixiv", flag.ContinueOnError)
	fs.Parse(args)

	if fs.NArg() < 1 {
		printPixivUsage()
		return
	}

	switch fs.Arg(0) {
	case "login":
		runPixivLogin(fs.Args()[1:])
	case "recommand":
		runPixivRecommand(fs.Args()[1:])
	case "help", "-h", "--help":
		printPixivUsage()
	default:
		fmt.Printf("unknown pixiv subcommand: %q\n", fs.Arg(0))
		os.Exit(1)
	}
}

func runPixivLogin(args []string) {
	_ = args // reserved for future flags

	// Check if already logged in.
	cfg, err := pixiv.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	if cfg != nil && cfg.PHPSESSID != "" {
		fmt.Println("already logged in (PHPSESSID saved)")
		fmt.Printf("run `%s pixiv login` again to re-authenticate\n", appName)
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

	path, _ := pixiv.ConfigPath()
	fmt.Printf("login successful as %s (user ID: %d)\n", userName, userID)
	fmt.Printf("config saved to %s\n", path)
}

func runPixivRecommand(args []string) {
	fs := flag.NewFlagSet("pixiv recommand", flag.ContinueOnError)
	limit := fs.Int("limit", 10, " number of illustrations to show")
	r18 := fs.Int("r18", 0, " R18 filter: 0=exclude, 1=include, 2=only R18")
	fs.Parse(args)

	phpsessid := os.Getenv("PIXIV_PHPSESSID")

	// Fall back to config file if env var is not set.
	if phpsessid == "" {
		cfg, err := pixiv.LoadConfig()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		if cfg == nil || cfg.PHPSESSID == "" {
			fmt.Fprintln(os.Stderr, "error: no credentials found. Run `skills-cli pixiv login` first or set PIXIV_PHPSESSID")
			os.Exit(1)
		}
		phpsessid = cfg.PHPSESSID
	}

	client := pixiv.NewClient()
	client.SetPHPSESSID(phpsessid)

	if verbose {
		pixiv.SetVerbose(true)
		fmt.Println("fetching recommended illustrations...")
	}

	result, err := client.FetchRecommended()
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
			fmt.Fprintf(os.Stderr, "  ⚠ download failed: %v\n\n", err)
		} else {
			for _, p := range saved {
				fmt.Printf("  Saved: %s\n", p)
			}
			fmt.Println()
		}
	}
}

func printPixivUsage() {
	fmt.Printf("Usage: %s pixiv <subcommand> [flags]\n", appName)
	fmt.Println()
	fmt.Println("Subcommands:")
	fmt.Println("  login        save Pixiv PHPSESSID cookie")
	fmt.Println("  recommand    fetch recommended illustrations")
	fmt.Println("  help         show this help")
	fmt.Println()
	fmt.Println("Flags (recommand):")
	fmt.Println("  -limit       number of illustrations to show (default 10)")
	fmt.Println("  -r18         0=exclude R18, 1=include all, 2=only R18 (default 0)")
	fmt.Println()
	fmt.Println("Config:")
	fmt.Println("  PHPSESSID is saved to ~/.kitakami_hibiki/config/pixiv.json")
	fmt.Println("  Environment variable PIXIV_PHPSESSID takes priority over config")
	fmt.Println()
	fmt.Println("How to get PHPSESSID:")
	fmt.Println("  1. Log in to https://www.pixiv.net/ in your browser")
	fmt.Println("  2. Open DevTools (F12) → Application → Cookies → pixiv.net")
	fmt.Println("  3. Copy the value of the 'PHPSESSID' cookie")
}
