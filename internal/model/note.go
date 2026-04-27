package model

import "time"

type Note struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	FolderID    *string   `json:"folder_id,omitempty"`
	ContentPath string    `json:"content_path"`
	Content     string    `json:"content,omitempty"`
	WordCount   int       `json:"word_count"`
	Tags        []Tag     `json:"tags,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	IsDeleted   bool      `json:"is_deleted"`
}

type NoteCreate struct {
	Title    string  `json:"title"`
	FolderID *string `json:"folder_id,omitempty"`
	Content  string  `json:"content"`
	Tags     []string `json:"tags,omitempty"`
}

type NoteUpdate struct {
	Title    *string  `json:"title,omitempty"`
	FolderID *string  `json:"folder_id,omitempty"`
	Content  *string  `json:"content,omitempty"`
	Tags     []string `json:"tags,omitempty"`
}

type NoteListParams struct {
	FolderID *string `json:"folder_id,omitempty"`
	TagID    *string `json:"tag_id,omitempty"`
	Search   string  `json:"search,omitempty"`
	Page     int     `json:"page"`
	PageSize int     `json:"page_size"`
	SortBy   string  `json:"sort_by,omitempty"`
	SortDesc bool    `json:"sort_desc,omitempty"`
	IsDeleted bool   `json:"is_deleted,omitempty"`
}

type NoteListResult struct {
	Notes []Note `json:"notes"`
	Total int    `json:"total"`
}
