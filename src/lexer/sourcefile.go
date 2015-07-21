package lexer

import (
	"bytes"
	"io/ioutil"
	"path"
	"strings"
)

type Sourcefile struct {
	Path     string
	Name     string
	Contents []rune
	NewLines []int
	Tokens   []*Token
}

func NewSourcefile(filepath string) (*Sourcefile, error) {
	// TODO, get this to handle the rare //file//shit
	// cut out the filename from path
	// + 1 to cut out the slash.
	i, j := strings.LastIndex(filepath, "/")+1, strings.LastIndex(filepath, path.Ext(filepath))

	// this is the name of the file, not the path
	name := filepath[i:j]

	sf := &Sourcefile{Name: name, Path: filepath}
	sf.NewLines = append(sf.NewLines, -1)
	sf.NewLines = append(sf.NewLines, -1)

	contents, err := ioutil.ReadFile(sf.Path)
	if err != nil {
		return nil, err
	}

	sf.Contents = []rune(string(contents))
	return sf, nil
}

func (s *Sourcefile) GetLine(line int) string {
	return string(s.Contents[s.NewLines[line]+1 : s.NewLines[line+1]])
}

func (s *Sourcefile) MarkPos(pos Position) string {
	buf := new(bytes.Buffer)

	lineString := s.GetLine(pos.Line)
	pad := pos.Char - 1

	buf.WriteString(strings.Replace(lineString, "%", "%%", -1))
	buf.WriteRune('\n')
	for i := 0; i < pad; i++ {
		buf.WriteRune(' ')
	}
	buf.WriteRune('^')
	buf.WriteRune('\n')

	return buf.String()

}

func (s *Sourcefile) MarkSpan(span Span) string {
	buf := new(bytes.Buffer)

	for line := span.StartLine; line <= span.EndLine; line++ {
		lineString := s.GetLine(line)

		var pad int
		if line == span.StartLine {
			pad = span.StartChar - 1
		} else {
			pad = 0
		}

		var length int
		if line == span.EndLine {
			length = span.EndChar - span.StartChar
		} else {
			length = len([]rune(lineString))
		}

		buf.WriteString(strings.Replace(lineString, "%", "%%", -1))
		buf.WriteRune('\n')
		for i := 0; i < pad; i++ {
			buf.WriteRune(' ')
		}
		for i := 0; i < length; i++ {
			buf.WriteRune('^')
		}
		buf.WriteRune('\n')
	}

	return buf.String()
}
