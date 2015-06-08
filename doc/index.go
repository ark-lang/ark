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

const INDEX_TEMPLATE_STR = `<!DOCTYPE html>
<html lang="en">
    <head>
        <meta charset="UTF-8" />
        <title>Doctype Index</title>

        <link href='http://fonts.googleapis.com/css?family=Fira+Sans:300,400,500,700|Fira+Mono|Source+Serif+Pro:400,700' rel='stylesheet' type='text/css'>
        <link rel="stylesheet" type="text/css" href="style.css" />
    </head>

    <body>
        <div class="slab">
            <div class="wrapper">
                <h1 class="slab-title">Docgen Index</h1>
                <a href="index.html">Index</a>
            </div>
        </div>

        <div class="wrapper">
            <ul class="files">
                {{range .Files}}<li><a href="files/{{.Name}}.html">{{.Name}}</a></li>{{end}}
            </ul>
        </div>
    </body>
</html>`
