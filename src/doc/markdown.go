package doc

import (
	"html/template"

	"github.com/russross/blackfriday"
)

var (
	mdRendererFlags = blackfriday.HTML_SKIP_HTML | blackfriday.HTML_SKIP_STYLE | blackfriday.HTML_SKIP_IMAGES | blackfriday.HTML_NOREFERRER_LINKS | blackfriday.HTML_SAFELINK | blackfriday.HTML_USE_XHTML
	mdRenderer      = blackfriday.HtmlRenderer(mdRendererFlags, "", "")
	mdOptions       = blackfriday.Options{Extensions: blackfriday.EXTENSION_STRIKETHROUGH | blackfriday.EXTENSION_HARD_LINE_BREAK}
)

func parseMarkdown(input string) string {
	if input == "" {
		return ""
	}

	ret := blackfriday.MarkdownOptions([]byte(template.HTMLEscapeString(input)), mdRenderer, mdOptions)

	return string(ret)
}
