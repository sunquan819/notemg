package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/notemg/notemg/internal/config"
	"github.com/notemg/notemg/internal/model"
)

type NoteStore struct {
	db  *sql.DB
	cfg *config.Config
}

func NewNoteStore(db *sql.DB, cfg *config.Config) *NoteStore {
	return &NoteStore{db: db, cfg: cfg}
}

func (s *NoteStore) List(params model.NoteListParams) (*model.NoteListResult, error) {
	query := "SELECT id, title, folder_id, content_path, word_count, created_at, updated_at, is_deleted FROM notes WHERE 1=1"
	args := []interface{}{}
	argIdx := 1

	if params.IsDeleted {
		query += fmt.Sprintf(" AND is_deleted = 1")
	} else {
		query += fmt.Sprintf(" AND is_deleted = 0")
	}

	if params.FolderID != nil {
		query += fmt.Sprintf(" AND folder_id = $%d", argIdx)
		args = append(args, *params.FolderID)
		argIdx++
	}

	if params.TagID != nil {
		query += fmt.Sprintf(" AND id IN (SELECT note_id FROM note_tags WHERE tag_id = $%d)", argIdx)
		args = append(args, *params.TagID)
		argIdx++
	}

	if params.Search != "" {
		query += fmt.Sprintf(" AND title LIKE $%d", argIdx)
		args = append(args, "%"+params.Search+"%")
		argIdx++
	}

	countQuery := strings.Replace(query, "SELECT id, title, folder_id, content_path, word_count, created_at, updated_at, is_deleted", "SELECT COUNT(*)", 1)
	var total int
	if err := s.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("count notes: %w", err)
	}

	sortCol := "updated_at"
	switch params.SortBy {
	case "title":
		sortCol = "title"
	case "created_at":
		sortCol = "created_at"
	}
	sortDir := "DESC"
	if !params.SortDesc {
		sortDir = "ASC"
	}
	query += fmt.Sprintf(" ORDER BY %s %s", sortCol, sortDir)

	if params.Page < 1 {
		params.Page = 1
	}
	if params.PageSize < 1 {
		params.PageSize = 20
	}
	offset := (params.Page - 1) * params.PageSize
	query += fmt.Sprintf(" LIMIT %d OFFSET %d", params.PageSize, offset)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("query notes: %w", err)
	}
	defer rows.Close()

	notes := make([]model.Note, 0)
	for rows.Next() {
		var n model.Note
		var folderID sql.NullString
		if err := rows.Scan(&n.ID, &n.Title, &folderID, &n.ContentPath, &n.WordCount, &n.CreatedAt, &n.UpdatedAt, &n.IsDeleted); err != nil {
			return nil, fmt.Errorf("scan note: %w", err)
		}
		if folderID.Valid {
			n.FolderID = &folderID.String
		}
		notes = append(notes, n)
	}

	for i := range notes {
		s.loadTags(&notes[i])
	}

	return &model.NoteListResult{Notes: notes, Total: total}, nil
}

func (s *NoteStore) GetByID(id string) (*model.Note, error) {
	var n model.Note
	var folderID sql.NullString
	err := s.db.QueryRow(
		"SELECT id, title, folder_id, content_path, word_count, created_at, updated_at, is_deleted FROM notes WHERE id = $1",
		id,
	).Scan(&n.ID, &n.Title, &folderID, &n.ContentPath, &n.WordCount, &n.CreatedAt, &n.UpdatedAt, &n.IsDeleted)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get note: %w", err)
	}
	if folderID.Valid {
		n.FolderID = &folderID.String
	}

	content, err := s.readContent(n.ContentPath)
	if err != nil {
		return nil, fmt.Errorf("read note content: %w", err)
	}
	n.Content = content

	s.loadTags(&n)
	return &n, nil
}

func (s *NoteStore) Create(input model.NoteCreate) (*model.Note, error) {
	id := uuid.New().String()
	now := time.Now()
	contentPath := filepath.Join(s.cfg.Data.NotesDir, id+".md")

	if err := s.writeContent(contentPath, input.Content); err != nil {
		return nil, fmt.Errorf("write note content: %w", err)
	}

	wordCount := len(strings.Fields(input.Content))
	title := input.Title
	if title == "" {
		title = s.extractTitle(input.Content)
	}

	var folderID interface{}
	if input.FolderID != nil {
		folderID = *input.FolderID
	} else {
		folderID = nil
	}

	_, err := s.db.Exec(
		"INSERT INTO notes (id, title, folder_id, content_path, word_count, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7)",
		id, title, folderID, contentPath, wordCount, now, now,
	)
	if err != nil {
		return nil, fmt.Errorf("insert note: %w", err)
	}

	if len(input.Tags) > 0 {
		s.syncTags(id, input.Tags)
	}

	return s.GetByID(id)
}

func (s *NoteStore) Update(id string, input model.NoteUpdate) (*model.Note, error) {
	n, err := s.GetByID(id)
	if err != nil {
		return nil, err
	}
	if n == nil {
		return nil, fmt.Errorf("note not found: %s", id)
	}

	sets := []string{}
	args := []interface{}{}
	argIdx := 1

	if input.Title != nil {
		sets = append(sets, fmt.Sprintf("title = $%d", argIdx))
		args = append(args, *input.Title)
		argIdx++
	}
	if input.FolderID != nil {
		sets = append(sets, fmt.Sprintf("folder_id = $%d", argIdx))
		args = append(args, *input.FolderID)
		argIdx++
	}
	if input.Content != nil {
		if err := s.writeContent(n.ContentPath, *input.Content); err != nil {
			return nil, fmt.Errorf("write note content: %w", err)
		}
		wordCount := len(strings.Fields(*input.Content))
		sets = append(sets, fmt.Sprintf("word_count = $%d", argIdx))
		args = append(args, wordCount)
		argIdx++

		if n.Title == "" || n.Title == s.extractTitle(n.Content) {
			newTitle := s.extractTitle(*input.Content)
			if newTitle != "" {
				sets = append(sets, fmt.Sprintf("title = $%d", argIdx))
				args = append(args, newTitle)
				argIdx++
			}
		}
	}

	if len(sets) > 0 {
		sets = append(sets, fmt.Sprintf("updated_at = $%d", argIdx))
		args = append(args, time.Now())
		argIdx++

		args = append(args, id)
		query := "UPDATE notes SET " + strings.Join(sets, ", ") + fmt.Sprintf(" WHERE id = $%d", argIdx)
		if _, err := s.db.Exec(query, args...); err != nil {
			return nil, fmt.Errorf("update note: %w", err)
		}
	}

	if input.Tags != nil {
		s.syncTags(id, input.Tags)
	}

	return s.GetByID(id)
}

func (s *NoteStore) Delete(id string) error {
	_, err := s.db.Exec("UPDATE notes SET is_deleted = 1, updated_at = $1 WHERE id = $2", time.Now(), id)
	return err
}

func (s *NoteStore) Restore(id string) error {
	_, err := s.db.Exec("UPDATE notes SET is_deleted = 0, updated_at = $1 WHERE id = $2", time.Now(), id)
	return err
}

func (s *NoteStore) PermanentDelete(id string) error {
	n, err := s.GetByID(id)
	if err != nil {
		return err
	}
	if n == nil {
		return nil
	}

	fullPath := s.cfg.DataDir(n.ContentPath)
	os.Remove(fullPath)

	_, err = s.db.Exec("DELETE FROM notes WHERE id = $1", id)
	return err
}

func (s *NoteStore) Move(id string, folderID *string) error {
	var fid interface{}
	if folderID != nil {
		fid = *folderID
	}
	_, err := s.db.Exec("UPDATE notes SET folder_id = $1, updated_at = $2 WHERE id = $3", fid, time.Now(), id)
	return err
}

func (s *NoteStore) Duplicate(id string) (*model.Note, error) {
	n, err := s.GetByID(id)
	if err != nil {
		return nil, err
	}
	if n == nil {
		return nil, fmt.Errorf("note not found: %s", id)
	}

	tagNames := make([]string, len(n.Tags))
	for i, t := range n.Tags {
		tagNames[i] = t.Name
	}

	return s.Create(model.NoteCreate{
		Title:    n.Title + " (copy)",
		FolderID: n.FolderID,
		Content:  n.Content,
		Tags:     tagNames,
	})
}

func (s *NoteStore) loadTags(n *model.Note) {
	rows, err := s.db.Query(
		"SELECT t.id, t.name FROM tags t JOIN note_tags nt ON t.id = nt.tag_id WHERE nt.note_id = $1",
		n.ID,
	)
	if err != nil {
		return
	}
	defer rows.Close()

	n.Tags = make([]model.Tag, 0)
	for rows.Next() {
		var t model.Tag
		if err := rows.Scan(&t.ID, &t.Name); err == nil {
			n.Tags = append(n.Tags, t)
		}
	}
}

func (s *NoteStore) syncTags(noteID string, tagNames []string) {
	s.db.Exec("DELETE FROM note_tags WHERE note_id = $1", noteID)

	for _, name := range tagNames {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}

		var tagID string
		err := s.db.QueryRow("SELECT id FROM tags WHERE name = $1", name).Scan(&tagID)
		if err == sql.ErrNoRows {
			tagID = uuid.New().String()
			_, err = s.db.Exec("INSERT INTO tags (id, name) VALUES ($1, $2)", tagID, name)
			if err != nil {
				s.db.QueryRow("SELECT id FROM tags WHERE name = $1", name).Scan(&tagID)
			}
		}

		if tagID != "" {
			s.db.Exec("INSERT OR IGNORE INTO note_tags (note_id, tag_id) VALUES ($1, $2)", noteID, tagID)
		}
	}
}

func (s *NoteStore) readContent(contentPath string) (string, error) {
	fullPath := s.cfg.DataDir(contentPath)
	data, err := os.ReadFile(fullPath)
	if os.IsNotExist(err) {
		return "", nil
	}
	return string(data), err
}

func (s *NoteStore) writeContent(contentPath string, content string) error {
	fullPath := s.cfg.DataDir(contentPath)
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(fullPath, []byte(content), 0644)
}

func (s *NoteStore) extractTitle(content string) string {
	lines := strings.SplitN(content, "\n", 10)
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "# ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "# "))
		}
	}
	return ""
}
