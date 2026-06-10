package wechat

// --- 通用 ---

// ErrorResponse 微信公众平台 API 通用错误响应。
type ErrorResponse struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

// AccessTokenResponse 获取 access_token 的响应。
type AccessTokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

// --- 草稿 ---

// DraftArticle 单篇图文。
type DraftArticle struct {
	Title              string `json:"title"`
	Author             string `json:"author,omitempty"`
	Digest             string `json:"digest,omitempty"`
	Content            string `json:"content"`
	ContentSourceURL   string `json:"content_source_url,omitempty"`
	ThumbMediaID       string `json:"thumb_media_id"`
	NeedOpenComment    int    `json:"need_open_comment,omitempty"`    // 0 不打开, 1 打开
	OnlyFansCanComment int    `json:"only_fans_can_comment,omitempty"` // 0 所有人, 1 粉丝
	URL                string `json:"url,omitempty"`
	ShowCover          int    `json:"show_cover_pic,omitempty"`
}

// DraftCreateRequest 创建草稿请求体。
type DraftCreateRequest struct {
	Articles []DraftArticle `json:"articles"`
}

// DraftCreateResponse 创建草稿响应。
type DraftCreateResponse struct {
	ErrorResponse
	MediaID string `json:"media_id"`
}

// DraftUpdateRequest 更新草稿请求体。
type DraftUpdateRequest struct {
	MediaID  string         `json:"media_id"`
	Index    int            `json:"index"` // 多图文时更新的文章序号，从 0 开始
	Articles []DraftArticle `json:"articles"`
}

// DraftGetResponse 获取草稿响应。
type DraftGetResponse struct {
	ErrorResponse
	NewsItem []DraftArticle `json:"news_item"`
}

// DraftListItem 草稿列表中的单条记录。
type DraftListItem struct {
	MediaID    string         `json:"media_id"`
	Content    DraftGetResponse `json:"content"`
	UpdateTime int64          `json:"update_time"`
}

// DraftListResponse 草稿列表响应。
type DraftListResponse struct {
	ErrorResponse
	TotalCount int             `json:"total_count"`
	ItemCount  int             `json:"item_count"`
	Items      []DraftListItem `json:"item"`
}

// DraftDeleteRequest 删除草稿请求体。
type DraftDeleteRequest struct {
	MediaID string `json:"media_id"`
}

// --- 发布 ---

// PublishSubmitRequest 发布草稿请求体。
type PublishSubmitRequest struct {
	DraftID string `json:"draft_id"`
}

// PublishSubmitResponse 发布草稿响应。
type PublishSubmitResponse struct {
	ErrorResponse
	PublishID string `json:"publish_id"`
}

// PublishedArticle 已发布文章信息。
type PublishedArticle struct {
	Title  string `json:"title"`
	URL    string `json:"url"`
	Digest string `json:"digest"`
}

// PublishedItem 发布列表中的单条记录。
type PublishedItem struct {
	ArticleID  string             `json:"article_id"`
	Article    PublishedArticle   `json:"article"`
	UpdateTime int64              `json:"update_time"`
}

// PublishListResponse 发布列表响应。
type PublishListResponse struct {
	ErrorResponse
	TotalCount int             `json:"total_count"`
	ItemCount  int             `json:"item_count"`
	Items      []PublishedItem `json:"item"`
}

// PublishGetResponse 获取发布文章响应。
type PublishGetResponse struct {
	ErrorResponse
	NewsItem []PublishedArticle `json:"news_item"`
}

// PublishDeleteRequest 删除已发布文章请求体。
type PublishDeleteRequest struct {
	ArticleID string `json:"article_id"`
}

// PublishDeleteResponse 删除已发布文章响应。
type PublishDeleteResponse struct {
	ErrorResponse
}

// --- 素材 ---

// MediaUploadResponse 上传图片响应。
type MediaUploadResponse struct {
	ErrorResponse
	URL      string `json:"url"`
	MediaID  string `json:"media_id"`  // 上传永久素材时返回
	Item     string `json:"item"`      // uploadimg 返回的 url
	ThumbURL string `json:"thumb_url"`
}

// MaterialAddResponse 新增永久素材响应（用于上传缩略图）。
type MaterialAddResponse struct {
	ErrorResponse
	MediaID string `json:"media_id"`
	URL     string `json:"url"` // 新增图文素材时返回
}
