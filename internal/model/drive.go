package model

import "time"

type DriveFile struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Type        string    `json:"type"`        // file/folder
	MimeType    string    `json:"mime_type"`
	Size        int64     `json:"size"`
	ParentID    *string   `json:"parent_id,omitempty"`
	Path        string    `json:"path"`
	ThumbPath   string    `json:"thumb_path,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	IsDeleted   bool      `json:"is_deleted"`
}

type DriveFileCreate struct {
	Name     string  `json:"name"`
	ParentID *string `json:"parent_id,omitempty"`
	Type     string  `json:"type"` // file/folder
}

type DriveFileUpdate struct {
	Name     *string  `json:"name,omitempty"`
	ParentID *string  `json:"parent_id,omitempty"`
}

type DriveListParams struct {
	ParentID *string `json:"parent_id,omitempty"`
	Type     string  `json:"type,omitempty"`    // file/folder/all
	Search   string  `json:"search,omitempty"`
	SortBy   string  `json:"sort_by,omitempty"`
	SortDesc bool    `json:"sort_desc,omitempty"`
}

type DriveListResult struct {
	Files []DriveFile `json:"files"`
	Total int         `json:"total"`
	Path  []DriveFile `json:"path"` // breadcrumb
}