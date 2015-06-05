package doc

import "github.com/ark-lang/ark/parser"

type Decl struct {
	AstDecl parser.Decl
	Docs    string
}
