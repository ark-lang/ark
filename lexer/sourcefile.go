package lexer

import (
	"io/ioutil"
)

type Sourcefile struct {
	Filename string
	Contents []rune
	Tokens   []*Token
}

func NewSourcefile(file string) (*Sourcefile, error) {
	sf := &Sourcefile{Filename: file}

	contents, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	sf.Contents = []rune(string(contents))
	return sf, nil
}
