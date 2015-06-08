package doc

import (
	"html/template"
	"os"
)

var indexTemplate = template.Must(template.New("index").Parse(INDEX_TEMPLATE_STR))

type IndexTempData struct {
	Files []*File
}

func (v *Docgen) generateIndex() {
	out, err := os.OpenFile(v.Dir+"index.html", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		panic(err)
	}
	defer out.Close()

	indexTempData := IndexTempData{
		Files: v.output,
	}

	err = indexTemplate.Execute(out, indexTempData)
	if err != nil {
		panic(err)
	}
}

const INDEX_TEMPLATE_STR = `<!DOCTYPE>
<html>
    <head>
        <title>Doctype Index</title>
		<meta charset="UTF-8" />
		<link rel="stylesheet" type="text/css" href="style.css" />
    </head>

    <body>
        <h1>Docgen Index</h1>

		<dl>
        	{{range .Files}}<dd><a href="files/{{.Name}}.html">{{.Name}}</a></dd>{{end}}
		</dl>
    </body>
</html>`
