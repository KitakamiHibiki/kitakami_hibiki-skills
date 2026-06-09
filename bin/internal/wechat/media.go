package wechat

import (
	"encoding/json"
	"fmt"
)

const (
	uploadImagePath = "/cgi-bin/media/uploadimg"      // 文章正文图片
	uploadThumbPath = "/cgi-bin/media/uploadthumb"     // 封面缩略图
)

// UploadImage 上传文章正文图片。返回图片 URL。
func (c *Client) UploadImage(filePath string) (string, error) {
	body, err := c.postForm(uploadImagePath+"?", "media", filePath)
	if err != nil {
		return "", err
	}

	var resp MediaUploadResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return "", fmt.Errorf("parse upload response: %w", err)
	}
	if resp.URL == "" {
		return "", fmt.Errorf("upload succeeded but no URL returned")
	}
	return resp.URL, nil
}

// UploadThumb 上传封面缩略图。返回 media_id。
func (c *Client) UploadThumb(filePath string) (string, error) {
	body, err := c.postForm(uploadThumbPath+"?", "media", filePath)
	if err != nil {
		return "", err
	}

	var resp MediaUploadResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return "", fmt.Errorf("parse upload response: %w", err)
	}
	if resp.URL == "" {
		return "", fmt.Errorf("upload succeeded but no URL returned")
	}
	return resp.URL, nil
}
