package wechat

import (
	"encoding/json"
	"fmt"
)

const (
	publishSubmitPath = "/cgi-bin/freepublish/submit"
	publishListPath   = "/cgi-bin/freepublish/batchget"
	publishGetPath    = "/cgi-bin/freepublish/get"
	publishDeletePath = "/cgi-bin/freepublish/delete"
)

// SubmitPublish 提交发布草稿。draftID 为草稿的 media_id。
// 返回 publish_id，可用于查询发布状态。
func (c *Client) SubmitPublish(draftID string) (string, error) {
	body, err := c.postJSON(publishSubmitPath+"?", PublishSubmitRequest{DraftID: draftID})
	if err != nil {
		return "", err
	}

	var resp PublishSubmitResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return "", fmt.Errorf("parse response: %w", err)
	}
	return resp.PublishID, nil
}

// ListPublished 列出已发布文章。offset: 偏移量，count: 获取数量（1~20）。
func (c *Client) ListPublished(offset, count int) (*PublishListResponse, error) {
	body, err := c.postJSON(publishListPath+"?", map[string]int{
		"offset": offset,
		"count":  count,
	})
	if err != nil {
		return nil, err
	}

	var resp PublishListResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}
	return &resp, nil
}

// GetPublished 获取已发布文章详情。
func (c *Client) GetPublished(articleID string) (*PublishGetResponse, error) {
	body, err := c.postJSON(publishGetPath+"?", map[string]string{
		"article_id": articleID,
	})
	if err != nil {
		return nil, err
	}

	var resp PublishGetResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}
	return &resp, nil
}

// DeletePublished 删除已发布文章。
func (c *Client) DeletePublished(articleID string) error {
	_, err := c.postJSON(publishDeletePath+"?", PublishDeleteRequest{ArticleID: articleID})
	return err
}
