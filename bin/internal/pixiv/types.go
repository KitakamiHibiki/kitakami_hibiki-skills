package pixiv

// IllustUser represents the user info within an illustration.
type IllustUser struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// IllustTag represents a single tag on an illustration.
type IllustTag struct {
	Name string `json:"name"`
}

// IllustImageURLs contains URLs for different image sizes.
type IllustImageURLs struct {
	SquareMedium string `json:"square_medium"`
	Medium       string `json:"medium"`
	Large        string `json:"large"`
}

// UserResponse represents the response from /ajax/user/extra.
type UserResponse struct {
	Error   bool   `json:"error"`
	Message string `json:"message"`
	Body    struct {
		UserID   int    `json:"user_id"`
		Name     string `json:"name"`
		Account  string `json:"account"`
		Image    string `json:"image"`
	} `json:"body"`
}

// WebIllust is the web API format for an illustration (/ajax/illust/* responses).
type WebIllust struct {
	ID            string   `json:"id"`
	Title         string   `json:"title"`
	IllustType    int      `json:"illustType"`
	XRestrict     int      `json:"xRestrict"`
	Description   string   `json:"description"`
	Tags          []string `json:"tags"`
	UserID        string   `json:"userId"`
	UserName      string   `json:"userName"`
	Width         int      `json:"width"`
	Height        int      `json:"height"`
	PageCount     int      `json:"pageCount"`
	BookmarkCount int      `json:"bookmarkCount"`
	ViewCount     int      `json:"viewCount"`
	LikeCount     int      `json:"likeCount"`
	URLs          struct {
		Thumb   string `json:"thumb"`
		Small   string `json:"small"`
		Regular string `json:"regular"`
		Original string `json:"original"`
	} `json:"urls"`
	CreateDate string `json:"createDate"`
}

// WebTopIllustResponse represents the response from /ajax/top/illust.
type WebTopIllustResponse struct {
	Error   bool   `json:"error"`
	Message string `json:"message"`
	Body    struct {
		Thumbnails struct {
			Illust []WebIllust `json:"illust"`
		} `json:"thumbnails"`
	} `json:"body"`
}

// WebIllustPageURLs contains image URLs for a single page from the pages endpoint.
type WebIllustPageURLs struct {
	ThumbMini  string `json:"thumb_mini"`
	Small      string `json:"small"`
	Regular    string `json:"regular"`
	Original   string `json:"original"`
}

// WebIllustPagesResponse represents the response from /ajax/illust/{id}/pages.
type WebIllustPagesResponse struct {
	Error   bool   `json:"error"`
	Message string `json:"message"`
	Body    []struct {
		URLs   WebIllustPageURLs `json:"urls"`
		Width  int               `json:"width"`
		Height int               `json:"height"`
	} `json:"body"`
}

// Illust represents a single Pixiv illustration (canonical format for display).
type Illust struct {
	ID             int             `json:"id"`
	Title          string          `json:"title"`
	Type           string          `json:"type"`
	XRestrict      int             `json:"x_restrict"`
	Caption        string          `json:"caption"`
	Width          int             `json:"width"`
	Height         int             `json:"height"`
	PageCount      int             `json:"page_count"`
	TotalBookmarks int             `json:"total_bookmarks"`
	TotalView      int             `json:"total_view"`
	User           IllustUser      `json:"user"`
	Tags           []IllustTag     `json:"tags"`
	ImageURLs      IllustImageURLs `json:"image_urls"`
}

// IllustResponse represents the API response for recommended illustrations.
type IllustResponse struct {
	Illusts []Illust `json:"illusts"`
	NextURL *string  `json:"next_url"`
}
