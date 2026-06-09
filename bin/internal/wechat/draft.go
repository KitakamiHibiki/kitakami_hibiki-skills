package wechat

import (
	"encoding/json"
	"fmt"
)

const (
	draftCreatePath  = "/cgi-bin/draft/create"
	draftUpdatePath  = "/cgi-bin/draft/update"
	draftDeletePath  = "/cgi-bin/draft/delete"
	draftGetPath     = "/cgi-bin/draft/get"
	draftListPath    = "/cgi-bin/draft/batchget"
)

// CreateDraft 创建草稿。支持多图文（多个 DraftArticle）。
func (c *Client) CreateDraft(articles []DraftArticle) (string, error) {
	body, err := c.postJSON(draftCreatePath+"?", DraftCreateRequest{Articles: articles})
	if err != nil {
		return "", err
	}

	var resp DraftCreateResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return "", fmt.Errorf("parse response: %w", err)
	}
	return resp.MediaID, nil
}

// UpdateDraft 更新草稿。index 为多图文中的文章序号（从 0 开始），-1 表示替换全部。
func (c *Client) UpdateDraft(mediaID string, index int, articles []DraftArticle) error {
	_, err := c.postJSON(draftUpdatePath+"?", DraftUpdateRequest{
		MediaID:  mediaID,
		Index:    index,
		Articles: articles,
	})
	return err
}

// DeleteDraft 删除草稿。
func (c *Client) DeleteDraft(mediaID string) error {
	_, err := c.postJSON(draftDeletePath+"?", DraftDeleteRequest{MediaID: mediaID})
	return err
}

// GetDraft 获取草稿详情。
func (c *Client) GetDraft(mediaID string) (*DraftGetResponse, error) {
	body, err := c.postJSON(draftGetPath+"?", DraftDeleteRequest{MediaID: mediaID})
	if err != nil {
		return nil, err
	}

	var resp DraftGetResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}
	return &resp, nil
}

// ListDrafts 列出草稿列表。offset: 偏移量，count: 获取数量（1~20）。
func (c *Client) ListDrafts(offset, count int) (*DraftListResponse, error) {
	body, err := c.postJSON(draftListPath+"?", map[string]int{
		"offset": offset,
		"count":  count,
	})
	if err != nil {
		return nil, err
	}

	var resp DraftListResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}
	return &resp, nil
}
