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

const FILE_TEMPLATE_STR = `<!DOCTYPE html>
<html lang="en">
    <head>
		<meta charset="UTF-8" />
        <title>Doctype Index</title>

		<link href='http://fonts.googleapis.com/css?family=Fira+Sans:300,400,500,700|Fira+Mono|Source+Serif+Pro:400,700' rel='stylesheet' type='text/css'>
		<link rel="stylesheet" type="text/css" href="{{.RootLoc}}style.css" />
    </head>

    <body>
        <div class="slab">
        	<div class="wrapper">
		        <h1 class="slab-title">File {{.Name}}</h1>
				<a href="{{.RootLoc}}index.html">Index</a>
			</div>
		</div>

		<div class="wrapper">
	        <section class="doc">
				<h2>Overview</h2>
				<ul>
					{{range .VariableDecls}}<li><a href="#{{.Ident}}">{{.Snippet}}</a></li>{{end}}
					{{range .StructDecls}}<li><a href="#{{.Ident}}">{{.Snippet}}</a></li>{{end}}
					{{range .FunctionDecls}}<li><a href="#{{.Ident}}">{{.Snippet}}</a></li>{{end}}
				</ul>
			</section>

			<section class="doc">
				<h2>Variables</h2>
					{{range .VariableDecls}}
					<h3 id="{{.Ident}}">{{.Ident}}</h3>
					<pre class="snippet"><code>{{.Snippet}}</code></pre>
					<div class="doccomment">{{.ParsedDocs}}</div>
				{{end}}
			</section>

			<section class="doc">
				<h2>Structs</h2>
					{{range .StructDecls}}
					<h3 id="{{.Ident}}">{{.Ident}}</h3>
					<pre class="snippet"><code>{{.Snippet}}</code></pre>
					<div class="doccomment">{{.ParsedDocs}}</div>
				{{end}}
			</section>

			<section class="doc">
				<h2>Functions</h2>
					{{range .FunctionDecls}}
					<h3 id="{{.Ident}}">{{.Ident}}</h3>
					<pre class="snippet"><code>{{.Snippet}}</code></pre>
					<div class="doccomment">{{.ParsedDocs}}</div>
				{{end}}
			</section>
		</div>
	</body>
</html>`
