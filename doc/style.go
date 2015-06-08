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
a, p, h1, h2, h3, h4, h5, h6 {
	font-family: sans-serif;
}

a {
	text-decoration: none;
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
