package model

type VideoData struct {
	CategoryTags   string `json:"categoryTags"`
	CommentCounter int64  `json:"commentCounter"`
	ContentID      string `json:"contentId"`
	Description    string `json:"description"`
	LengthSeconds  int64  `json:"lengthSeconds"`
	MylistCounter  int64  `json:"mylistCounter"`
	StartTime      string `json:"startTime"`
	Tags           string `json:"tags"`
	ThumbnailURL   string `json:"thumbnailUrl"`
	Title          string `json:"title"`
	ViewCounter    int64  `json:"viewCounter"`
	UpdatedAt      string `json:"updatedAt"`
}
