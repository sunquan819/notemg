package model

type Tag struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type TagCreate struct {
	Name string `json:"name"`
}

type TagWithCount struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	NoteCount int    `json:"note_count"`
}
