package doc

import (
	"html/template"

	"github.com/ark-lang/ark/parser"
)

type Decl struct {
	Node       parser.Documentable
	Docs       string
	ParsedDocs template.HTML // docs after markdown parsing
	Ident      string        // identifier
	Snippet    string        // code snippet of declaration
}

func (v *Decl) process() {
	v.ParsedDocs = template.HTML(parseMarkdown(v.Docs))

	switch v.Node.(type) {
	case *parser.FunctionDecl:
		v.Ident, v.Snippet = generateFunctionDeclSnippet(v.Node.(*parser.FunctionDecl))
	case *parser.StructDecl:
		v.Ident, v.Snippet = generateStructDeclSnippet(v.Node.(*parser.StructDecl))
	case *parser.TraitDecl:
		v.Ident, v.Snippet = generateTraitDeclSnippet(v.Node.(*parser.TraitDecl))
	case *parser.ImplDecl:
		v.Ident, v.Snippet = generateImplDeclSnippet(v.Node.(*parser.ImplDecl))
	case *parser.VariableDecl:
		v.Ident, v.Snippet = generateVariableDeclSnippet(v.Node.(*parser.VariableDecl))
	default:
		panic("unimplimented decl type in doc")
	}
}

func generateFunctionDeclSnippet(decl *parser.FunctionDecl) (ident, snippet string) {
	ident = decl.Function.Name

	snippet = parser.KEYWORD_FUNC + " " + ident + "("
	for i, par := range decl.Function.Parameters {
		snippet += par.Variable.Name + ": " + par.Variable.Type.TypeName()
		if i < len(decl.Function.Parameters)-1 {
			snippet += ", "
		}
	}
	snippet += ")"
	if decl.Function.ReturnType != nil {
		snippet += ": " + decl.Function.ReturnType.TypeName()

	}

	return
}

func generateStructDeclSnippet(decl *parser.StructDecl) (ident, snippet string) {
	ident = decl.Struct.Name
	snippet = "struct " + decl.Struct.Name + " {"
	for i, member := range decl.Struct.Variables {
		snippet += "\n    " + member.Variable.Name + ": " + member.Variable.Type.TypeName()
		if member.Assignment != nil {
			snippet += " = [TODO values]"
		}
		if i < len(decl.Struct.Variables)-1 {
			snippet += ","
		} else {
			snippet += "\n"
		}
	}
	snippet += "}"
	return
}

func generateTraitDeclSnippet(decl *parser.TraitDecl) (ident, snippet string) {
	ident = decl.Trait.Name
	snippet = "trait " + decl.Trait.Name + " {"
	for _, member := range decl.Trait.Functions {
		_, functionSnippet := generateFunctionDeclSnippet(member)
		snippet += "\n    " + functionSnippet + "\n"
	}
	snippet += "}"
	return
}

func generateImplDeclSnippet(decl *parser.ImplDecl) (ident, snippet string) {
	ident = decl.StructName
	snippet = "impl " + decl.StructName
	if decl.TraitName != "" {
		snippet += " for " + decl.TraitName
	}
	snippet += " {"
	for _, member := range decl.Functions {
		_, functionSnippet := generateFunctionDeclSnippet(member)
		snippet += "\n    " + functionSnippet + "\n"
	}
	snippet += "}"
	return
}

func generateVariableDeclSnippet(decl *parser.VariableDecl) (ident, snippet string) {
	ident = decl.Variable.Name
	snippet = ident + ": " + decl.Variable.Type.TypeName()
	if decl.Assignment != nil {
		snippet += " = [TODO values]"
	}
	return
}
