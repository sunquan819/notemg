package search

import (
	"fmt"
	"os"

	"github.com/blevesearch/bleve/v2"
)

type SearchResult struct {
	ID    string  `json:"id"`
	Title string  `json:"title"`
	Score float64 `json:"score"`
}

type Searcher struct {
	index bleve.Index
	path  string
}

type NoteDocument struct {
	Title   string
	Content string
}

func NewSearcher(indexPath string) (*Searcher, error) {
	var idx bleve.Index
	var err error

	mapping := bleve.NewIndexMapping()
	docMapping := bleve.NewDocumentMapping()

	titleField := bleve.NewTextFieldMapping()
	titleField.Analyzer = "standard"
	docMapping.AddFieldMappingsAt("Title", titleField)

	contentField := bleve.NewTextFieldMapping()
	contentField.Analyzer = "standard"
	docMapping.AddFieldMappingsAt("Content", contentField)

	mapping.AddDocumentMapping("NoteDocument", docMapping)

	if _, statErr := os.Stat(indexPath); os.IsNotExist(statErr) {
		idx, err = bleve.New(indexPath, mapping)
	} else {
		idx, err = bleve.Open(indexPath)
	}

	if err != nil {
		return nil, fmt.Errorf("open search index: %w", err)
	}

	return &Searcher{index: idx, path: indexPath}, nil
}

func (s *Searcher) Index(id, title, content string) error {
	doc := NoteDocument{Title: title, Content: content}
	return s.index.Index(id, doc)
}

func (s *Searcher) Delete(id string) error {
	return s.index.Delete(id)
}

func (s *Searcher) Search(query string) ([]SearchResult, error) {
	q := bleve.NewQueryStringQuery(query)
	req := bleve.NewSearchRequest(q)
	req.Size = 50
	req.Fields = []string{"Title"}

	res, err := s.index.Search(req)
	if err != nil {
		return nil, err
	}

	results := make([]SearchResult, 0, len(res.Hits))
	for _, hit := range res.Hits {
		title := ""
		if t, ok := hit.Fields["Title"].(string); ok {
			title = t
		}
		results = append(results, SearchResult{
			ID:    hit.ID,
			Title: title,
			Score: hit.Score,
		})
	}

	return results, nil
}

func (s *Searcher) Close() error {
	if s.index != nil {
		return s.index.Close()
	}
	return nil
}
