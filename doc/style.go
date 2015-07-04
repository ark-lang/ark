package doc

import (
	"os"
)

func (v *Docgen) generateStyle() {
	out, err := os.OpenFile(v.Dir+"style.css", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		panic(err)
	}
	defer out.Close()

	out.WriteString(STYLE)
}

const STYLE = `
* {
	margin: 0;
	padding: 0;
    box-sizing: border-box;
}

html {
	overflow-y: scroll;
}

body {
    font-family: "Source Serif Pro", Georgia, Times New Roman, serif;
}

body > .wrapper {
	background: #eee;
	margin-top: 20px;
}

.wrapper {
	display: block;
	width: 980px;
	margin: 0 auto;
	padding: 20px;
}

a {
	font-family: "Fira Sans", "Helvetica Neue", Helvetica, Arial, sans-serif;
}

.slab a {
	color: #fff;
	font-weight: 800;
}

.slab a:hover {
	color: #eee;
}

a {
	color: #24424F;
	text-decoration: none;
}

a:hover {
	color: #112833;
	text-decoration: underline;
}

section.doc {
	padding: 20px;
}

.slab {
	display: block;
	width: 100%;
	background: #546E7A;
	padding: 2rem;
	color: #fff;
}

.slab-title {
	font-size: 3em;
	color: #fff;
}

ul {
	margin-left: 20px;
}

h1, h2, h3, h4, h6 {
    font-family: "Fira Sans", "Helvetica Neue", Helvetica, Arial, sans-serif;
}

pre, code {
    font-family: "Fira Mono", monospace;
}

a {
	text-decoration: none;
}

.declname {
	margin-top: 1em;
}

.doccomment {
	padding-left: 25px;
}

.doccomment h1, .doccomment h2, .doccomment h3, .doccomment h4, .doccomment h5, .doccomment h6 {
	color: #444444;
}

.snippet {
	margin: 10px;
	padding: 10px;
	background-color: #DDDDDD;
}`
