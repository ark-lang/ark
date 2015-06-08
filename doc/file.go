package doc

import (
	"html/template"
	"os"
	"path/filepath"
	"strings"
)

var fileTemplate = template.Must(template.New("file").Parse(FILE_TEMPLATE_STR))

type File struct {
	Name          string
	RootLoc       string // path from this file to the root directory (the directory containing index.html)
	VariableDecls []*Decl
	StructDecls   []*Decl
	FunctionDecls []*Decl
}

type FileTempData struct {
	Name  string
	Decls []*Decl
}

func (v *File) dir() string {
	return "files/" + filepath.Dir(v.Name) + "/"
}

func (v *File) base() string {
	return filepath.Base(v.Name)
}

func (v *Docgen) generateFile(file *File) {
	fileDir := v.Dir + file.dir()
	err := os.MkdirAll(fileDir, os.ModeDir|0777)
	if err != nil {
		panic(err)
	}

	out, err := os.OpenFile(fileDir+file.base()+".html", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		panic(err)
	}
	defer out.Close()

	for i := 0; i < 1+strings.Count(filepath.Clean(file.Name), "/"); i++ {
		file.RootLoc += "../"
	}

	err = fileTemplate.Execute(out, file)
	if err != nil {
		panic(err)
	}
}

const FILE_TEMPLATE_STR = `<!DOCTYPE>
<html>
    <head>
        <title>Doctype Index</title>
		<meta charset="UTF-8" />
		<link rel="stylesheet" type="text/css" href="{{.RootLoc}}style.css" />
    </head>

    <body>
		<a href="{{.RootLoc}}index.html">Index</a>
        <h1>File {{.Name}}</h1>

		<h2>Overview</h2>
		<dl>
			{{range .VariableDecls}}<dd><a href="#{{.Ident}}">{{.Snippet}}</a></dd>{{end}}
			{{range .StructDecls}}<dd><a href="#{{.Ident}}">{{.Snippet}}</a></dd>{{end}}
			{{range .FunctionDecls}}<dd><a href="#{{.Ident}}">{{.Snippet}}</a></dd>{{end}}
		</dl>

		<h2>Variables</h2>
		{{range .VariableDecls}}
		<h3 id="{{.Ident}}">{{.Ident}}</h3>
		<pre class="snippet"><code>{{.Snippet}}</code></pre>
		<div class="doccomment">{{.ParsedDocs}}</div>
		{{end}}

		<h2>Structs</h2>
		{{range .StructDecls}}
		<h3 id="{{.Ident}}">{{.Ident}}</h3>
		<pre class="snippet"><code>{{.Snippet}}</code></pre>
		<div class="doccomment">{{.ParsedDocs}}</div>
		{{end}}

		<h2>Functions</h2>
		{{range .FunctionDecls}}
		<h3 id="{{.Ident}}">{{.Ident}}</h3>
		<pre class="snippet"><code>{{.Snippet}}</code></pre>
		<div class="doccomment">{{.ParsedDocs}}</div>
		{{end}}
	</body>
</html>`
