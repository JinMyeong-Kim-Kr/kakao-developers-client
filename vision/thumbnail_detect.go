package vision

import (
	"bytes"
	"encoding/json"
	"fmt"
	"internal/common"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

// Thumbnail represents coordinates of the point starting the thumbnail image and its width, height
type Thumbnail struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

// ThumbnailResult represents a document of a detected thumbnail result.
type ThumbnailResult struct {
	Width     int       `json:"width"`
	Height    int       `json:"height"`
	Thumbnail Thumbnail `json:"thumbnail"`
}

// ThumbnailDetectResult represents a Thumbnail Detection result.
type ThumbnailDetectResult struct {
	Rid    string          `json:"rid"`
	Result ThumbnailResult `json:"result"`
}

// String implements fmt.Stringer.
func (tr ThumbnailDetectResult) String() string { return common.String(tr) }

// SaveAs saves tr to @filename
//
// The file extension must be .json.
func (tr ThumbnailDetectResult) SaveAs(filename string) error {
	return common.SaveAsJSON(tr, filename)
}

// ThumbnailDetectIniailizer is a lazy thumbnail detector.
type ThumbnailDetectInitializer struct {
	AuthKey  string
	Image    *os.File
	ImageUrl string
	Width    int
	Height   int
}

// ThumbnailDetect helps to create a thumbnail image by detecting the representative area out of the given @source.
// Refer to https://developers.kakao.com/docs/latest/ko/vision/dev-guide#extract-thumbnail for more details.
func ThumbnailDetect(source string) *ThumbnailDetectInitializer {
	url, file := CheckSourceType(source)
	return &ThumbnailDetectInitializer{
		AuthKey:  common.KeyPrefix,
		ImageUrl: url,
		Image:    file,
		Width:    0,
		Height:   0,
	}
}

// AuthorizeWith sets the authorization key to @key
func (ti *ThumbnailDetectInitializer) AuthorizeWith(key string) *ThumbnailDetectInitializer {
	ti.AuthKey = common.FormatKey(key)
	return ti
}

// WidthTo sets Image width to @ratio.
func (ti *ThumbnailDetectInitializer) WidthTo(ratio int) *ThumbnailDetectInitializer {
	ti.Width = ratio
	return ti
}

// HeightTo sets Image Height to @ratio.
func (ti *ThumbnailDetectInitializer) HeightTo(ratio int) *ThumbnailDetectInitializer {
	ti.Height = ratio
	return ti
}

// Collect returns the thumbnail detection result.
func (ti *ThumbnailDetectInitializer) Collect() (res ThumbnailDetectResult, err error) {
	client := new(http.Client)
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	if ti.Image != nil {
		part, err := writer.CreateFormFile("image", filepath.Base(ti.Image.Name()))
		if err != nil {
			return res, err
		}
		io.Copy(part, ti.Image)
	}
	defer writer.Close()

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/thumbnail/detect?image_url=%s&width=%d&height=%d",
		prefix, ti.ImageUrl, ti.Width, ti.Height), body)
	if err != nil {
		return res, err
	}
	req.Close = true
	req.Header.Set(common.Authorization, ti.AuthKey)

	if ti.Image != nil {
		req.Header.Set("Content-Type", writer.FormDataContentType())
	} else {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	defer ti.Image.Close()

	resp, err := client.Do(req)
	if err != nil {
		return res, err
	}

	defer resp.Body.Close()

	if err = json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return res, err
	}
	return
}
