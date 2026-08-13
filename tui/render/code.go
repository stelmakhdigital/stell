package render

import (
	"bytes"
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
)

// HighlightCode runs Chroma for lang. Unknown languages return plain text.
func HighlightCode(code, lang string) string {
	return highlight(code, lang, true)
}

func highlight(code, lang string, dark bool) string {
	code = strings.TrimRight(code, "\n")
	lexer := lexerFor(lang)
	if lexer == nil {
		return code
	}
	iter, err := lexer.Tokenise(nil, code)
	if err != nil {
		return code
	}
	styleName := "catppuccin-mocha"
	if !dark {
		styleName = "catppuccin-latte"
	}
	style := styles.Get(styleName)
	if style == nil {
		style = styles.Fallback
	}
	formatter := formatters.Get("terminal256")
	if formatter == nil {
		formatter = formatters.TTY256
	}
	var buf bytes.Buffer
	if err := formatter.Format(&buf, style, iter); err != nil {
		return code
	}
	out := buf.String()
	if out == "" {
		return code
	}
	return strings.TrimRight(out, "\n")
}

func lexerFor(lang string) chroma.Lexer {
	lang = strings.TrimSpace(strings.ToLower(lang))
	if lang == "" {
		return nil
	}
	if l := lexers.Get(lang); l != nil {
		return chroma.Coalesce(l)
	}
	return nil
}
