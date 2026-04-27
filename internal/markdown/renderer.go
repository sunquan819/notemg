package markdown

import (
	"bytes"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer/html"
)

type Engine struct {
	gm goldmark.Markdown
}

func NewEngine() *Engine {
	gm := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			extension.DefinitionList,
			extension.Footnote,
			extension.NewCJK(),
		),
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
			html.WithXHTML(),
		),
	)

	return &Engine{gm: gm}
}

func (e *Engine) Render(src string) (string, error) {
	var buf bytes.Buffer
	if err := e.gm.Convert([]byte(src), &buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}
