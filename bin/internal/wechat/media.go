package wechat

import (
	"encoding/json"
	"fmt"
)

const (
	uploadImagePath   = "/cgi-bin/media/uploadimg"             // 文章正文图片
	addMaterialPath   = "/cgi-bin/material/add_material?type=" // 永久素材（type=thumb）
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

// UploadThumb 上传封面缩略图（永久素材）。返回 media_id，用于草稿的 thumb_media_id。
func (c *Client) UploadThumb(filePath string) (string, error) {
	body, err := c.postForm(addMaterialPath+"thumb", "media", filePath)
	if err != nil {
		return "", err
	}

	var resp MaterialAddResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return "", fmt.Errorf("parse upload response: %w", err)
	}
	if resp.MediaID == "" {
		return "", fmt.Errorf("upload succeeded but no media_id returned")
	}
	return resp.MediaID, nil
}
