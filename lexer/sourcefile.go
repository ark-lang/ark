package lexer

import (
	"io/ioutil"
	"strings"
	"path"
)

type Sourcefile struct {
	Path string
	Name string
	Contents []rune
	Tokens   []*Token
}

func NewSourcefile(filepath string) (*Sourcefile, error) {
	// TODO, get this to handle the rare //file//shit
	// cut out the filename from path
	// + 1 to cut out the slash.
	i, j := strings.LastIndex(filepath, "/") + 1, strings.LastIndex(filepath, path.Ext(filepath))

	// this is the name of the file, not the path
	name := filepath[i:j]

	sf := &Sourcefile{Name: name, Path: filepath}

	contents, err := ioutil.ReadFile(sf.Path)
	if err != nil {
		return nil, err
	}

	sf.Contents = []rune(string(contents))
	return sf, nil
}
