package model

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
)

type VideoData struct {
	CategoryTags    string `json:"categoryTags"`
	CommentCounter  int64  `json:"commentCounter"`
	ContentID       string `json:"contentId"`
	Description     string `json:"description"`
	DescriptionFile string `json:"descriptionFile"`
	LengthSeconds   int64  `json:"lengthSeconds"`
	MylistCounter   int64  `json:"mylistCounter"`
	StartTime       string `json:"startTime"`
	Tags            string `json:"tags"`
	ThumbnailURL    string `json:"thumbnailUrl"`
	Title           string `json:"title"`
	ViewCounter     int64  `json:"viewCounter"`
	UpdatedAt       string `json:"updatedAt"`
}

func (vd *VideoData) SplitDescription(dir string) error {
	hash := sha256.Sum256([]byte(vd.Description))
	hashStr := hex.EncodeToString(hash[:])
	hashStr = hashStr[:20]

	vd.DescriptionFile = hashStr + ".txt"

	f, err := os.Create(filepath.Join(dir, vd.DescriptionFile))
	if err != nil {
		return err
	}

	_, err = f.Write([]byte(vd.Description))
	if err != nil {
		return err
	}

	vd.Description = ""

	return nil
}
