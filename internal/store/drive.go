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

type DriveStore struct {
	db  *sql.DB
	cfg *config.Config
}

func NewDriveStore(db *sql.DB, cfg *config.Config) *DriveStore {
	return &DriveStore{db: db, cfg: cfg}
}

func (s *DriveStore) Init() error {
	schema := `
CREATE TABLE IF NOT EXISTS drive_files (
    id          TEXT PRIMARY KEY,
    name        TEXT NOT NULL,
    type        TEXT NOT NULL,
    mime_type   TEXT DEFAULT '',
    size        INTEGER DEFAULT 0,
    parent_id   TEXT,
    path        TEXT NOT NULL,
    thumb_path  TEXT DEFAULT '',
    created_at  DATETIME NOT NULL DEFAULT (datetime('now')),
    updated_at  DATETIME NOT NULL DEFAULT (datetime('now')),
    is_deleted  INTEGER DEFAULT 0,
    FOREIGN KEY (parent_id) REFERENCES drive_files(id)
);
CREATE INDEX IF NOT EXISTS idx_drive_parent_id ON drive_files(parent_id);
CREATE INDEX IF NOT EXISTS idx_drive_type ON drive_files(type);
CREATE INDEX IF NOT EXISTS idx_drive_is_deleted ON drive_files(is_deleted);
`
	_, err := s.db.Exec(schema)
	return err
}

func (s *DriveStore) List(params model.DriveListParams) (*model.DriveListResult, error) {
	query := "SELECT id, name, type, mime_type, size, parent_id, path, thumb_path, created_at, updated_at, is_deleted FROM drive_files WHERE is_deleted = 0"
	args := []interface{}{}
	argIdx := 1

	if params.ParentID != nil {
		query += fmt.Sprintf(" AND parent_id = $%d", argIdx)
		args = append(args, *params.ParentID)
		argIdx++
	} else {
		query += " AND parent_id IS NULL"
	}

	if params.Type != "" && params.Type != "all" {
		query += fmt.Sprintf(" AND type = $%d", argIdx)
		args = append(args, params.Type)
		argIdx++
	}

	if params.Search != "" {
		query += fmt.Sprintf(" AND name LIKE $%d", argIdx)
		args = append(args, "%"+params.Search+"%")
		argIdx++
	}

	sortCol := "name"
	switch params.SortBy {
	case "size":
		sortCol = "size"
	case "created_at":
		sortCol = "created_at"
	case "updated_at":
		sortCol = "updated_at"
	}
	sortDir := "ASC"
	if params.SortDesc {
		sortDir = "DESC"
	}
	if sortCol == "name" {
		query += fmt.Sprintf(" ORDER BY type DESC, %s %s", sortCol, sortDir)
	} else {
		query += fmt.Sprintf(" ORDER BY %s %s", sortCol, sortDir)
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("query drive files: %w", err)
	}
	defer rows.Close()

	files := make([]model.DriveFile, 0)
	for rows.Next() {
		var f model.DriveFile
		var parentID sql.NullString
		var thumbPath sql.NullString
		if err := rows.Scan(&f.ID, &f.Name, &f.Type, &f.MimeType, &f.Size, &parentID, &f.Path, &thumbPath, &f.CreatedAt, &f.UpdatedAt, &f.IsDeleted); err != nil {
			return nil, fmt.Errorf("scan drive file: %w", err)
		}
		if parentID.Valid {
			f.ParentID = &parentID.String
		}
		if thumbPath.Valid {
			f.ThumbPath = thumbPath.String
		}
		files = append(files, f)
	}

	breadcrumb := make([]model.DriveFile, 0)
	if params.ParentID != nil {
		breadcrumb = s.getBreadcrumb(*params.ParentID)
	}

	return &model.DriveListResult{Files: files, Total: len(files), Path: breadcrumb}, nil
}

func (s *DriveStore) getBreadcrumb(id string) []model.DriveFile {
	result := make([]model.DriveFile, 0)
	current := id
	for current != "" {
		var f model.DriveFile
		var parentID sql.NullString
		err := s.db.QueryRow("SELECT id, name, type, parent_id FROM drive_files WHERE id = $1", current).Scan(&f.ID, &f.Name, &f.Type, &parentID)
		if err != nil {
			break
		}
		result = append([]model.DriveFile{f}, result...)
		if parentID.Valid {
			current = parentID.String
		} else {
			current = ""
		}
	}
	return result
}

func (s *DriveStore) GetByID(id string) (*model.DriveFile, error) {
	var f model.DriveFile
	var parentID sql.NullString
	var thumbPath sql.NullString
	err := s.db.QueryRow(
		"SELECT id, name, type, mime_type, size, parent_id, path, thumb_path, created_at, updated_at, is_deleted FROM drive_files WHERE id = $1",
		id,
	).Scan(&f.ID, &f.Name, &f.Type, &f.MimeType, &f.Size, &parentID, &f.Path, &thumbPath, &f.CreatedAt, &f.UpdatedAt, &f.IsDeleted)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get drive file: %w", err)
	}
	if parentID.Valid {
		f.ParentID = &parentID.String
	}
	if thumbPath.Valid {
		f.ThumbPath = thumbPath.String
	}
	return &f, nil
}

func (s *DriveStore) CreateFolder(name string, parentID *string) (*model.DriveFile, error) {
	id := uuid.New().String()
	now := time.Now()
	path := s.cfg.Data.AttachmentsDir

	if parentID != nil {
		parent, err := s.GetByID(*parentID)
		if err != nil || parent == nil {
			return nil, fmt.Errorf("parent folder not found")
		}
		path = parent.Path
	}

	folderPath := filepath.Join(path, name)
	if err := os.MkdirAll(folderPath, 0755); err != nil {
		return nil, fmt.Errorf("create folder: %w", err)
	}

	_, err := s.db.Exec(
		"INSERT INTO drive_files (id, name, type, parent_id, path, created_at, updated_at) VALUES ($1, $2, 'folder', $3, $4, $5, $6)",
		id, name, parentID, folderPath, now, now,
	)
	if err != nil {
		return nil, fmt.Errorf("insert folder: %w", err)
	}

	return s.GetByID(id)
}

func (s *DriveStore) SaveFile(name string, mimeType string, size int64, parentID *string, content []byte) (*model.DriveFile, error) {
	id := uuid.New().String()
	now := time.Now()

	basePath := s.cfg.AttachmentsPath()
	if parentID != nil {
		parent, err := s.GetByID(*parentID)
		if err != nil || parent == nil {
			parentID = nil
		} else {
			basePath = parent.Path
		}
	}

	ext := filepath.Ext(name)
	storedName := id + ext
	filePath := filepath.Join(basePath, storedName)

	if err := os.WriteFile(filePath, content, 0644); err != nil {
		return nil, fmt.Errorf("write file: %w", err)
	}

	thumbPath := ""
	if strings.HasPrefix(mimeType, "image/") {
		thumbPath = s.generateThumbnail(filePath, id)
	}

	_, err := s.db.Exec(
		"INSERT INTO drive_files (id, name, type, mime_type, size, parent_id, path, thumb_path, created_at, updated_at) VALUES ($1, $2, 'file', $3, $4, $5, $6, $7, $8, $9)",
		id, name, mimeType, size, parentID, filePath, thumbPath, now, now,
	)
	if err != nil {
		os.Remove(filePath)
		return nil, fmt.Errorf("insert file: %w", err)
	}

	return s.GetByID(id)
}

func (s *DriveStore) generateThumbnail(filePath string, id string) string {
	return filePath
}

func (s *DriveStore) Delete(id string) error {
	f, err := s.GetByID(id)
	if err != nil || f == nil {
		return nil
	}

	if f.Type == "folder" {
		children, _ := s.List(model.DriveListParams{ParentID: &id})
		for _, child := range children.Files {
			s.Delete(child.ID)
		}
		os.Remove(f.Path)
	} else {
		os.Remove(f.Path)
		if f.ThumbPath != "" {
			os.Remove(f.ThumbPath)
		}
	}

	_, err = s.db.Exec("DELETE FROM drive_files WHERE id = $1", id)
	return err
}

func (s *DriveStore) Rename(id string, name string) (*model.DriveFile, error) {
	f, err := s.GetByID(id)
	if err != nil || f == nil {
		return nil, fmt.Errorf("file not found")
	}

	if f.Type == "folder" {
		newPath := filepath.Join(filepath.Dir(f.Path), name)
		if err := os.Rename(f.Path, newPath); err != nil {
			return nil, fmt.Errorf("rename folder: %w", err)
		}
		_, err = s.db.Exec("UPDATE drive_files SET name = $1, path = $2, updated_at = $3 WHERE id = $4", name, newPath, time.Now(), id)
	} else {
		_, err = s.db.Exec("UPDATE drive_files SET name = $1, updated_at = $2 WHERE id = $3", name, time.Now(), id)
	}

	if err != nil {
		return nil, err
	}
	return s.GetByID(id)
}

func (s *DriveStore) Move(id string, parentID *string) (*model.DriveFile, error) {
	f, err := s.GetByID(id)
	if err != nil || f == nil {
		return nil, fmt.Errorf("file not found")
	}

	if parentID != nil {
		parent, err := s.GetByID(*parentID)
		if err != nil || parent == nil || parent.Type != "folder" {
			return nil, fmt.Errorf("target folder not found")
		}
	} else {
		parentID = nil
	}

	if f.Type == "folder" {
		newPath := s.cfg.AttachmentsPath()
		if parentID != nil {
			parent, _ := s.GetByID(*parentID)
			if parent != nil {
				newPath = parent.Path
			}
		}
		newPath = filepath.Join(newPath, f.Name)
		if err := os.Rename(f.Path, newPath); err != nil {
			return nil, fmt.Errorf("move folder: %w", err)
		}
		_, err = s.db.Exec("UPDATE drive_files SET parent_id = $1, path = $2, updated_at = $3 WHERE id = $4", parentID, newPath, time.Now(), id)
	} else {
		_, err = s.db.Exec("UPDATE drive_files SET parent_id = $1, updated_at = $2 WHERE id = $3", parentID, time.Now(), id)
	}

	if err != nil {
		return nil, err
	}
	return s.GetByID(id)
}

func (s *DriveStore) Search(query string) ([]model.DriveFile, error) {
	rows, err := s.db.Query(
		"SELECT id, name, type, mime_type, size, parent_id, path, thumb_path, created_at, updated_at, is_deleted FROM drive_files WHERE is_deleted = 0 AND name LIKE $1 ORDER BY name ASC LIMIT 100",
		"%"+query+"%",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	files := make([]model.DriveFile, 0)
	for rows.Next() {
		var f model.DriveFile
		var parentID sql.NullString
		var thumbPath sql.NullString
		if err := rows.Scan(&f.ID, &f.Name, &f.Type, &f.MimeType, &f.Size, &parentID, &f.Path, &thumbPath, &f.CreatedAt, &f.UpdatedAt, &f.IsDeleted); err != nil {
			continue
		}
		if parentID.Valid {
			f.ParentID = &parentID.String
		}
		if thumbPath.Valid {
			f.ThumbPath = thumbPath.String
		}
		files = append(files, f)
	}
	return files, nil
}