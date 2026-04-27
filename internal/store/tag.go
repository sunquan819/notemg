package store

import (
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/notemg/notemg/internal/model"
)

type TagStore struct {
	db *sql.DB
}

func NewTagStore(db *sql.DB) *TagStore {
	return &TagStore{db: db}
}

func (s *TagStore) List() ([]model.TagWithCount, error) {
	rows, err := s.db.Query(
		"SELECT t.id, t.name, COUNT(nt.note_id) as note_count FROM tags t LEFT JOIN note_tags nt ON t.id = nt.tag_id LEFT JOIN notes n ON nt.note_id = n.id AND n.is_deleted = 0 GROUP BY t.id, t.name ORDER BY t.name ASC",
	)
	if err != nil {
		return nil, fmt.Errorf("query tags: %w", err)
	}
	defer rows.Close()

	tags := make([]model.TagWithCount, 0)
	for rows.Next() {
		var t model.TagWithCount
		if err := rows.Scan(&t.ID, &t.Name, &t.NoteCount); err != nil {
			return nil, fmt.Errorf("scan tag: %w", err)
		}
		tags = append(tags, t)
	}
	return tags, nil
}

func (s *TagStore) GetByID(id string) (*model.Tag, error) {
	var t model.Tag
	err := s.db.QueryRow("SELECT id, name FROM tags WHERE id = $1", id).Scan(&t.ID, &t.Name)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get tag: %w", err)
	}
	return &t, nil
}

func (s *TagStore) Create(input model.TagCreate) (*model.Tag, error) {
	id := uuid.New().String()
	_, err := s.db.Exec("INSERT INTO tags (id, name) VALUES ($1, $2)", id, input.Name)
	if err != nil {
		existing, _ := s.GetByName(input.Name)
		if existing != nil {
			return existing, nil
		}
		return nil, fmt.Errorf("insert tag: %w", err)
	}
	return s.GetByID(id)
}

func (s *TagStore) Delete(id string) error {
	_, err := s.db.Exec("DELETE FROM note_tags WHERE tag_id = $1", id)
	if err != nil {
		return err
	}
	_, err = s.db.Exec("DELETE FROM tags WHERE id = $1", id)
	return err
}

func (s *TagStore) GetByName(name string) (*model.Tag, error) {
	var t model.Tag
	err := s.db.QueryRow("SELECT id, name FROM tags WHERE name = $1", name).Scan(&t.ID, &t.Name)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (s *TagStore) NotesByTag(tagID string) ([]string, error) {
	rows, err := s.db.Query("SELECT note_id FROM note_tags WHERE tag_id = $1", tagID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ids := make([]string, 0)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err == nil {
			ids = append(ids, id)
		}
	}
	return ids, nil
}
