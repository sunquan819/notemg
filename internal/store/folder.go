package store

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/notemg/notemg/internal/model"
)

type FolderStore struct {
	db *sql.DB
}

func NewFolderStore(db *sql.DB) *FolderStore {
	return &FolderStore{db: db}
}

func (s *FolderStore) Tree() ([]model.Folder, error) {
	rows, err := s.db.Query(
		"SELECT id, name, parent_id, sort_order, created_at, updated_at, is_deleted FROM folders WHERE is_deleted = 0 ORDER BY sort_order ASC, name ASC",
	)
	if err != nil {
		return nil, fmt.Errorf("query folders: %w", err)
	}
	defer rows.Close()

	all := make([]model.Folder, 0)
	for rows.Next() {
		var f model.Folder
		var parentID sql.NullString
		if err := rows.Scan(&f.ID, &f.Name, &parentID, &f.SortOrder, &f.CreatedAt, &f.UpdatedAt, &f.IsDeleted); err != nil {
			return nil, fmt.Errorf("scan folder: %w", err)
		}
		if parentID.Valid {
			f.ParentID = &parentID.String
		}
		all = append(all, f)
	}

	return s.buildTree(all), nil
}

func (s *FolderStore) GetByID(id string) (*model.Folder, error) {
	var f model.Folder
	var parentID sql.NullString
	err := s.db.QueryRow(
		"SELECT id, name, parent_id, sort_order, created_at, updated_at, is_deleted FROM folders WHERE id = $1",
		id,
	).Scan(&f.ID, &f.Name, &parentID, &f.SortOrder, &f.CreatedAt, &f.UpdatedAt, &f.IsDeleted)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get folder: %w", err)
	}
	if parentID.Valid {
		f.ParentID = &parentID.String
	}
	return &f, nil
}

func (s *FolderStore) Create(input model.FolderCreate) (*model.Folder, error) {
	id := uuid.New().String()
	now := time.Now()

	var parentID interface{}
	if input.ParentID != nil {
		parentID = *input.ParentID
	}

	var maxOrder sql.NullInt64
	s.db.QueryRow("SELECT MAX(sort_order) FROM folders WHERE parent_id IS NOT NULL AND parent_id = $1", parentID).Scan(&maxOrder)
	sortOrder := 0
	if maxOrder.Valid {
		sortOrder = int(maxOrder.Int64) + 1
	}

	_, err := s.db.Exec(
		"INSERT INTO folders (id, name, parent_id, sort_order, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)",
		id, input.Name, parentID, sortOrder, now, now,
	)
	if err != nil {
		return nil, fmt.Errorf("insert folder: %w", err)
	}

	return s.GetByID(id)
}

func (s *FolderStore) Update(id string, input model.FolderUpdate) (*model.Folder, error) {
	sets := []string{}
	args := []interface{}{}
	argIdx := 1

	if input.Name != nil {
		sets = append(sets, fmt.Sprintf("name = $%d", argIdx))
		args = append(args, *input.Name)
		argIdx++
	}
	if input.ParentID != nil {
		sets = append(sets, fmt.Sprintf("parent_id = $%d", argIdx))
		args = append(args, *input.ParentID)
		argIdx++
	}
	if input.SortOrder != nil {
		sets = append(sets, fmt.Sprintf("sort_order = $%d", argIdx))
		args = append(args, *input.SortOrder)
		argIdx++
	}

	if len(sets) > 0 {
		sets = append(sets, fmt.Sprintf("updated_at = $%d", argIdx))
		args = append(args, time.Now())
		argIdx++
		args = append(args, id)
		query := "UPDATE folders SET " + joinSets(sets) + fmt.Sprintf(" WHERE id = $%d", argIdx)
		if _, err := s.db.Exec(query, args...); err != nil {
			return nil, fmt.Errorf("update folder: %w", err)
		}
	}

	return s.GetByID(id)
}

func (s *FolderStore) Delete(id string) error {
	_, err := s.db.Exec("UPDATE folders SET is_deleted = 1, updated_at = $1 WHERE id = $2", time.Now(), id)
	return err
}

func (s *FolderStore) buildTree(folders []model.Folder) []model.Folder {
	childrenMap := make(map[string][]model.Folder)
	var roots []model.Folder

	for i := range folders {
		if folders[i].ParentID == nil {
			roots = append(roots, folders[i])
		} else {
			pid := *folders[i].ParentID
			childrenMap[pid] = append(childrenMap[pid], folders[i])
		}
	}

	var attachChildren func(parents []model.Folder)
	attachChildren = func(parents []model.Folder) {
		for i := range parents {
			if children, ok := childrenMap[parents[i].ID]; ok {
				parents[i].Children = children
				attachChildren(parents[i].Children)
			}
		}
	}
	attachChildren(roots)

	return roots
}

func joinSets(sets []string) string {
	result := sets[0]
	for i := 1; i < len(sets); i++ {
		result += ", " + sets[i]
	}
	return result
}
