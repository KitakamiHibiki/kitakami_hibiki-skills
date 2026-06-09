package pixiv

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"skills/bin/internal/proxy"
)

const (
	apiBase = "https://www.pixiv.net"
)

// Client is a Pixiv web API client authenticated via PHPSESSID cookie.
type Client struct {
	http *http.Client
}

// NewClient creates a new Pixiv web API client with proxy support.
func NewClient() *Client {
	jar, _ := cookiejar.New(nil)
	transport := &http.Transport{
		Proxy: proxy.ProxyFromEnvironment(),
	}
	return &Client{
		http: &http.Client{Transport: transport, Jar: jar},
	}
}

// SetPHPSESSID sets the PHPSESSID cookie on the client for www.pixiv.net.
func (c *Client) SetPHPSESSID(cookieValue string) {
	u := &url.URL{Scheme: "https", Host: "www.pixiv.net"}
	cookie := &http.Cookie{
		Name:    "PHPSESSID",
		Value:   cookieValue,
		Path:    "/",
		Domain:  ".pixiv.net",
		Expires: time.Now().Add(365 * 24 * time.Hour),
	}
	c.http.Jar.SetCookies(u, []*http.Cookie{cookie})
}

// DownloadImage downloads an image from the given URL and saves it to destPath.
// Pixiv requires the Referer header for hotlink protection.
func (c *Client) DownloadImage(imageURL, destPath string) error {
	req, err := http.NewRequest("GET", imageURL, nil)
	if err != nil {
		return fmt.Errorf("create download request: %w", err)
	}
	req.Header.Set("Referer", "https://www.pixiv.net/")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Accept", "image/avif,image/webp,image/apng,*/*")

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("download request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("download failed (HTTP %d)", resp.StatusCode)
	}

	if err := os.MkdirAll(filepath.Dir(destPath), 0700); err != nil {
		return fmt.Errorf("create directory: %w", err)
	}

	f, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer f.Close()

	written, err := io.Copy(f, resp.Body)
	if err != nil {
		return fmt.Errorf("save file: %w", err)
	}

	if verbose {
		fmt.Printf("[debug] downloaded %s (%d bytes)\n", filepath.Base(destPath), written)
	}

	return nil
}

// FetchIllustPages fetches the actual page image URLs for an illustration.
// Returns the original URL for each page, keyed by page index.
func (c *Client) FetchIllustPages(illustID int) ([]string, error) {
	path := fmt.Sprintf("/ajax/illust/%d/pages", illustID)
	body, err := c.get(path)
	if err != nil {
		return nil, fmt.Errorf("fetch pages: %w", err)
	}

	var resp WebIllustPagesResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parse pages response: %w", err)
	}

	if resp.Error {
		return nil, fmt.Errorf("API error: %s", resp.Message)
	}

	if len(resp.Body) == 0 {
		return nil, fmt.Errorf("no pages found")
	}

	urls := make([]string, len(resp.Body))
	for i, p := range resp.Body {
		urls[i] = p.URLs.Original
		if urls[i] == "" {
			urls[i] = p.URLs.Regular
		}
	}
	return urls, nil
}

// DownloadIllust downloads all pages for an illust to the given directory.
// Files are named {id}_{page}.{ext} (page starts at 0).
// Skips pages whose files already exist on disk.
// Returns the list of saved (or already-present) file paths.
func (c *Client) DownloadIllust(illust Illust, dir string) ([]string, error) {
	if verbose {
		fmt.Printf("[debug] fetching pages for illust #%d\n", illust.ID)
	}
	pageURLs, err := c.FetchIllustPages(illust.ID)
	if err != nil {
		return nil, fmt.Errorf("get image URLs: %w", err)
	}

	ext := extFromURL(pageURLs[0])
	saved := make([]string, 0, len(pageURLs))
	for pageIdx, imageURL := range pageURLs {
		filename := fmt.Sprintf("%d_%03d%s", illust.ID, pageIdx, ext)
		destPath := filepath.Join(dir, filename)

		if _, err := os.Stat(destPath); err == nil {
			if verbose {
				fmt.Printf("[debug] skip (already exists): %s\n", filename)
			}
			saved = append(saved, destPath)
			continue
		}

		if err := c.DownloadImage(imageURL, destPath); err != nil {
			return saved, fmt.Errorf("download page %d: %w", pageIdx, err)
		}
		saved = append(saved, destPath)
	}
	return saved, nil
}

// extFromURL extracts the file extension from a URL path.
func extFromURL(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return ".jpg"
	}
	ext := filepath.Ext(u.Path)
	if ext == "" {
		return ".jpg"
	}
	return ext
}

// VerifySession calls /ajax/user/extra to check if the PHPSESSID is valid.
// Returns the user ID and name on success.
func (c *Client) VerifySession() (userID int, userName string, err error) {
	body, err := c.get("/ajax/user/extra")
	if err != nil {
		return 0, "", fmt.Errorf("verify session: %w", err)
	}

	var resp UserResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return 0, "", fmt.Errorf("parse user response: %w", err)
	}

	if resp.Error {
		return 0, "", fmt.Errorf("session verification failed: %s", resp.Message)
	}

	return resp.Body.UserID, resp.Body.Name, nil
}

// FetchRanking fetches the ranking of illustrations by mode.
// Supported modes: daily, weekly, monthly, random.
func (c *Client) FetchRanking(mode string) (*IllustResponse, error) {
	switch mode {
	case "daily", "weekly", "monthly":
		return c.fetchRankingFromPHP(mode)
	case "random":
		return c.fetchRecommended()
	default:
		return nil, fmt.Errorf("unsupported mode: %q (use: daily, weekly, monthly, random)", mode)
	}
}

// fetchRankingFromPHP fetches ranking data from the Pixiv ranking page.
func (c *Client) fetchRankingFromPHP(mode string) (*IllustResponse, error) {
	// Try the Ajax ranking endpoint (current Pixiv API).
	path := fmt.Sprintf("/ajax/ranking/illust?mode=%s&content=illust", mode)
	body, err := c.get(path)
	if err == nil {
		if verbose {
			preview := string(body)
			if len(preview) > 1000 {
				preview = preview[:1000]
			}
			fmt.Printf("[debug] ranking JSON body:\n%s\n", preview)
		}
		var resp WebRankingResponse
		if err := json.Unmarshal(body, &resp); err == nil && !resp.Error && len(resp.Body.Ranking) > 0 {
			illusts := make([]Illust, len(resp.Body.Ranking))
			for i, r := range resp.Body.Ranking {
				illusts[i] = webRankingIllustToIllust(r)
			}
			return &IllustResponse{Illusts: illusts}, nil
		}
	}

	// Fallback: parse the HTML page.
	return c.fetchRankingFromHTML(mode)
}

// fetchRankingFromHTML parses the ranking page HTML for embedded illustration data.
func (c *Client) fetchRankingFromHTML(mode string) (*IllustResponse, error) {
	path := fmt.Sprintf("/ranking.php?mode=%s&content=illust", mode)
	req, err := http.NewRequest("GET", apiBase+path, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Referer", "https://www.pixiv.net/")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("rank request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("rank request failed (HTTP %d)", resp.StatusCode)
	}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	return parseRankingHTML(raw)
}

// parseRankingHTML scans HTML for embedded ranking data.
func parseRankingHTML(raw []byte) (*IllustResponse, error) {
	html := string(raw)

	// Try __NEXT_DATA__ (Next.js apps embed initial state here).
	re := regexp.MustCompile(`<script id="__NEXT_DATA__"[^>]*>(.*?)</script>`)
	if m := re.FindStringSubmatch(html); len(m) > 1 {
		// Current format: pageProps.assign.contents
		var nextData struct {
			Props struct {
				PageProps struct {
					Assign struct {
						Contents []WebRankingContent `json:"contents"`
					} `json:"assign"`
				} `json:"pageProps"`
			} `json:"props"`
		}
		if err := json.Unmarshal([]byte(m[1]), &nextData); err == nil && len(nextData.Props.PageProps.Assign.Contents) > 0 {
			illusts := make([]Illust, len(nextData.Props.PageProps.Assign.Contents))
			for i, r := range nextData.Props.PageProps.Assign.Contents {
				illusts[i] = webRankingContentToIllust(r)
			}
			return &IllustResponse{Illusts: illusts}, nil
		}

		// Older format (via rankingItems).
		var nextDataOld struct {
			Props struct {
				PageProps struct {
					RankingItems []WebRankingIllust `json:"rankingItems"`
				} `json:"pageProps"`
			} `json:"props"`
		}
		if err := json.Unmarshal([]byte(m[1]), &nextDataOld); err == nil && len(nextDataOld.Props.PageProps.RankingItems) > 0 {
			illusts := make([]Illust, len(nextDataOld.Props.PageProps.RankingItems))
			for i, r := range nextDataOld.Props.PageProps.RankingItems {
				illusts[i] = webRankingIllustToIllust(r)
			}
			return &IllustResponse{Illusts: illusts}, nil
		}
	}

	// Try pixiv.ranking.data (older Pixiv page format).
	re = regexp.MustCompile(`pixiv\.ranking\.data\s*=\s*({.*?});`)
	if m := re.FindStringSubmatch(html); len(m) > 1 {
		var rankData struct {
			Illusts []WebRankingIllust `json:"illusts"`
		}
		if err := json.Unmarshal([]byte(m[1]), &rankData); err == nil && len(rankData.Illusts) > 0 {
			illusts := make([]Illust, len(rankData.Illusts))
			for i, r := range rankData.Illusts {
				illusts[i] = webRankingIllustToIllust(r)
			}
			return &IllustResponse{Illusts: illusts}, nil
		}
	}

	return nil, fmt.Errorf("could not find ranking data in the page")
}

// webRankingContentToIllust converts a WebRankingContent to the canonical Illust type.
func webRankingContentToIllust(w WebRankingContent) Illust {
	tags := make([]IllustTag, len(w.Tags))
	for i, t := range w.Tags {
		tags[i] = IllustTag{Name: t}
	}

	illustType := "illust"
	switch w.IllustType {
	case "1":
		illustType = "manga"
	case "2":
		illustType = "ugoira"
	}

	pageCount, _ := strconv.Atoi(w.IllustPageCount)

	return Illust{
		ID:             w.IllustID,
		Title:          w.Title,
		Type:           illustType,
		XRestrict:      w.IllustContentType.Sexual,
		Width:          w.Width,
		Height:         w.Height,
		PageCount:      pageCount,
		TotalBookmarks: w.RatingCount,
		TotalView:      w.ViewCount,
		User: IllustUser{
			ID:   w.UserID,
			Name: w.UserName,
		},
		Tags: tags,
		ImageURLs: IllustImageURLs{
			SquareMedium: w.URL,
			Medium:       w.URL,
			Large:        w.URL,
		},
	}
}

// fetchRecommended fetches recommended illustrations using the Pixiv web API.
func (c *Client) fetchRecommended() (*IllustResponse, error) {
	body, err := c.get("/ajax/top/illust?mode=all")
	if err != nil {
		return nil, fmt.Errorf("recommend request: %w", err)
	}

	var resp WebTopIllustResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parse recommend response: %w", err)
	}

	if resp.Error {
		return nil, fmt.Errorf("API error: %s", resp.Message)
	}

	illusts := make([]Illust, len(resp.Body.Thumbnails.Illust))
	for i, w := range resp.Body.Thumbnails.Illust {
		illusts[i] = webIllustToIllust(w)
	}

	return &IllustResponse{
		Illusts: illusts,
	}, nil
}

// webIllustToIllust converts a web API illust to the canonical Illust type.
func webIllustToIllust(w WebIllust) Illust {
	tags := make([]IllustTag, len(w.Tags))
	for i, t := range w.Tags {
		tags[i] = IllustTag{Name: t}
	}

	illustType := "illust"
	switch w.IllustType {
	case 1:
		illustType = "manga"
	case 2:
		illustType = "ugoira"
	}

	id, _ := strconv.Atoi(w.ID)
	userID, _ := strconv.Atoi(w.UserID)

	return Illust{
		ID:             id,
		Title:          w.Title,
		Type:           illustType,
		XRestrict:      w.XRestrict,
		Caption:        w.Description,
		Width:          w.Width,
		Height:         w.Height,
		PageCount:      w.PageCount,
		TotalBookmarks: w.BookmarkCount,
		TotalView:      w.ViewCount,
		User: IllustUser{
			ID:   userID,
			Name: w.UserName,
		},
		Tags: tags,
		ImageURLs: IllustImageURLs{
			SquareMedium: w.URLs.Small,
			Medium:       w.URLs.Regular,
			Large:        w.URLs.Original,
		},
	}
}

// webRankingIllustToIllust converts a ranking API illust to the canonical Illust type.
func webRankingIllustToIllust(w WebRankingIllust) Illust {
	tags := make([]IllustTag, len(w.Tags))
	for i, t := range w.Tags {
		tags[i] = IllustTag{Name: t}
	}

	illustType := "illust"
	switch w.IllustType {
	case 1:
		illustType = "manga"
	case 2:
		illustType = "ugoira"
	}

	id, _ := strconv.Atoi(w.IllustID)
	userID, _ := strconv.Atoi(w.UserID)

	return Illust{
		ID:             id,
		Title:          w.Title,
		Type:           illustType,
		XRestrict:      w.XRestrict,
		Caption:        w.Description,
		Width:          w.Width,
		Height:         w.Height,
		PageCount:      w.PageCount,
		TotalBookmarks: w.BookmarkCount,
		TotalView:      w.ViewCount,
		User: IllustUser{
			ID:   userID,
			Name: w.UserName,
		},
		Tags: tags,
		ImageURLs: IllustImageURLs{
			SquareMedium: w.URLs.Small,
			Medium:       w.URLs.Regular,
			Large:        w.URLs.Original,
		},
	}
}

// get performs an authenticated GET request and returns the raw response body.
func (c *Client) get(path string) ([]byte, error) {
	req, err := http.NewRequest("GET", apiBase+path, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Referer", "https://www.pixiv.net/")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if verbose {
		fmt.Printf("[debug] GET %s → %d\n", path, resp.StatusCode)
		if resp.StatusCode != 200 {
			fmt.Printf("[debug] response body: %s\n", string(raw))
		}
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("request failed (HTTP %d): %s", resp.StatusCode, string(raw))
	}

	return raw, nil
}

// FormatIllust formats a single illust for display.
func FormatIllust(i Illust) string {
	var tags []string
	for _, t := range i.Tags {
		tags = append(tags, t.Name)
	}

	r18 := ""
	if i.XRestrict > 0 {
		r18 = " [R18]"
	}

	return fmt.Sprintf("[%s] #%d%s\n"+
		"  Title:     %s\n"+
		"  Artist:    %s\n"+
		"  Size:      %dx%d\n"+
		"  Pages:     %d\n"+
		"  Views:     %s\n"+
		"  Bookmarks: %s\n"+
		"  Tags:      %s\n"+
		"  URL:       https://www.pixiv.net/artworks/%d\n",
		i.Type, i.ID, r18,
		i.Title,
		i.User.Name,
		i.Width, i.Height,
		i.PageCount,
		formatCount(i.TotalView),
		formatCount(i.TotalBookmarks),
		strings.Join(tags, ", "),
		i.ID,
	)
}

func formatCount(n int) string {
	if n >= 10000 {
		return fmt.Sprintf("%.1f万", float64(n)/10000)
	}
	return strconv.Itoa(n)
}

var verbose bool

// SetVerbose controls debug output from this package.
func SetVerbose(v bool) {
	verbose = v
}
