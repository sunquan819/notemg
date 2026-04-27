package model

import "time"

type Attachment struct {
	ID        string    `json:"id"`
	NoteID    *string   `json:"note_id,omitempty"`
	Filename  string    `json:"filename"`
	FilePath  string    `json:"file_path"`
	FileSize  int64     `json:"file_size"`
	MimeType  string    `json:"mime_type"`
	CreatedAt time.Time `json:"created_at"`
}
