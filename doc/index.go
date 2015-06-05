package doc

import (
	"io"
	"text/template"
)

func (v *Docgen) generateIndex(out io.Writer) {
	indexTemplate := template.New("index")
	indexTemplate.Parse(INDEX_TEMPLATE_STR)
	indexTemplate.Execute(out, v.output)
}

const INDEX_TEMPLATE_STR = `<!DOCTYPE>
<html>
    <head>
        <title>Doctype Index</title>
    </head>

    <body>
        <h1>Docgen Index</h1>
        {{range .}}<p>{{.Name}}</p>{{end}}
    </body>
</html>`
