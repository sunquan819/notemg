package model

import "time"

type Folder struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	ParentID  *string   `json:"parent_id,omitempty"`
	SortOrder int       `json:"sort_order"`
	Children  []Folder  `json:"children,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	IsDeleted bool      `json:"is_deleted"`
}

type FolderCreate struct {
	Name     string  `json:"name"`
	ParentID *string `json:"parent_id,omitempty"`
}

type FolderUpdate struct {
	Name      *string `json:"name,omitempty"`
	ParentID  *string `json:"parent_id,omitempty"`
	SortOrder *int    `json:"sort_order,omitempty"`
}

type FolderTree struct {
	Folders []Folder `json:"folders"`
}
