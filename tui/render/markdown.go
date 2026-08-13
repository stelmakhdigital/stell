package render

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/budaev/agent/tui/theme"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/text"
)

const maxLines = 200
const maxSrc = 64 * 1024

// Markdown renders CommonMark to a terminal string using theme colors.
func Markdown(src string, width int, th theme.Theme) string {
	if width < 20 {
		width = 20
	}
	if len(src) > maxSrc {
		src = src[:maxSrc] + "\n"
	}
	md := goldmark.New(goldmark.WithExtensions(extension.GFM))
	reader := text.NewReader([]byte(src))
	doc := md.Parser().Parse(reader)
	w := &mdWriter{th: th, width: width, src: reader.Source(), dark: th.Dark()}
	walk(doc, w)
	out := strings.TrimRight(w.b.String(), "\n")
	if w.truncated {
		lines := strings.Split(out, "\n")
		if len(lines) > maxLines-1 {
			lines = lines[:maxLines-1]
		}
		return strings.Join(append(lines, "… [truncated]"), "\n")
	}
	return out
}

type mdWriter struct {
	b         strings.Builder
	lines     int
	truncated bool
	th        theme.Theme
	width     int
	src       []byte
	dark      bool
}

func (w *mdWriter) write(s string) {
	if w.lines >= maxLines {
		w.truncated = true
		return
	}
	w.b.WriteString(s)
	w.lines += strings.Count(s, "\n")
	if w.lines >= maxLines {
		w.truncated = true
	}
}

func (w *mdWriter) full() bool { return w.lines >= maxLines }

func walk(n ast.Node, w *mdWriter) {
	src := w.src
	ast.Walk(n, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		if w.full() {
			return ast.WalkStop, nil
		}
		if !entering {
			switch node.(type) {
			case *ast.Heading, *ast.Paragraph, *ast.ListItem, *ast.FencedCodeBlock, *ast.CodeBlock, *ast.Blockquote:
				if !strings.HasSuffix(w.b.String(), "\n") {
					w.write("\n")
				}
			}
			return ast.WalkContinue, nil
		}
		switch t := node.(type) {
		case *ast.Heading:
			txt := strings.TrimSpace(textOf(t, src))
			style := lipgloss.NewStyle().Bold(true).Foreground(w.th.Primary())
			w.write(style.Render(strings.Repeat("#", t.Level)+" "+txt) + "\n")
			return ast.WalkSkipChildren, nil
		case *ast.FencedCodeBlock:
			lang := strings.TrimSpace(string(t.Language(src)))
			code := codeText(t, src)
			w.write(lipgloss.NewStyle().Foreground(w.th.Muted()).Render("```"+lang) + "\n")
			w.write(highlight(code, lang, w.dark) + "\n")
			w.write(lipgloss.NewStyle().Foreground(w.th.Muted()).Render("```") + "\n")
			return ast.WalkSkipChildren, nil
		case *ast.CodeBlock:
			w.write(highlight(codeText(t, src), "", w.dark) + "\n")
			return ast.WalkSkipChildren, nil
		case *ast.ListItem:
			w.write(lipgloss.NewStyle().Foreground(w.th.Accent()).Render("• "))
		case *ast.ThematicBreak:
			w.write(lipgloss.NewStyle().Foreground(w.th.Muted()).Render(strings.Repeat("─", min(w.width, 40))) + "\n")
		case *ast.Text:
			s := string(t.Segment.Value(src))
			if t.SoftLineBreak() {
				s += " "
			}
			if t.HardLineBreak() {
				s += "\n"
			}
			w.write(lipgloss.NewStyle().Foreground(w.th.Foreground()).Render(s))
			return ast.WalkContinue, nil
		case *ast.CodeSpan:
			raw := textOf(t, src)
			w.write(lipgloss.NewStyle().Foreground(w.th.Secondary()).Render(raw))
			return ast.WalkSkipChildren, nil
		}
		return ast.WalkContinue, nil
	})
}

func textOf(n ast.Node, src []byte) string {
	var b strings.Builder
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		if t, ok := c.(*ast.Text); ok {
			b.Write(t.Segment.Value(src))
		} else {
			b.WriteString(textOf(c, src))
		}
	}
	return b.String()
}

func codeText(n ast.Node, src []byte) string {
	var b strings.Builder
	for i := 0; i < n.Lines().Len(); i++ {
		line := n.Lines().At(i)
		b.Write(line.Value(src))
	}
	return strings.TrimRight(b.String(), "\n")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
