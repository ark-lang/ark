package parser

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/ark-lang/ark/src/lexer"
	"github.com/ark-lang/ark/src/util/log"
)

type Module struct {
	Nodes       []Node
	File        *lexer.Sourcefile
	Path        string // this stores the path too, e.g src/main
	Name        *ModuleName
	GlobalScope *Scope
}

type ModuleLookup struct {
	Name     string
	Tree     *ParseTree
	Module   *Module
	Children map[string]*ModuleLookup
}

func NewModuleLookup(name string) *ModuleLookup {
	res := &ModuleLookup{
		Name:     name,
		Children: make(map[string]*ModuleLookup),
	}
	return res
}

func (v *ModuleLookup) Get(name *ModuleName) (*ModuleLookup, error) {
	ml := v

	for idx, part := range name.Parts {
		nml, ok := ml.Children[part]
		if ok {
			ml = nml
		} else {
			return nil, fmt.Errorf("Module not found in lookup: %s",
				(&ModuleName{Parts: name.Parts[0 : idx+1]}).String())
		}
	}

	return ml, nil
}

func (v *ModuleLookup) Create(name *ModuleName) *ModuleLookup {
	ml := v

	for _, part := range name.Parts {
		nml, ok := ml.Children[part]
		if !ok {
			nml = NewModuleLookup(part)
			ml.Children[part] = nml
		}
		ml = nml
	}

	return ml
}

func (v *ModuleLookup) Dump(i int) {
	if v.Name != "" {
		log.Debug("main", "%s", strings.Repeat(" ", i))
		log.Debugln("main", "%s", v.Name)
	}

	for _, child := range v.Children {
		child.Dump(i + 1)
	}
}

type ModuleName struct {
	Parts []string
}

func NewModuleName(node *NameNode) *ModuleName {
	res := &ModuleName{
		Parts: make([]string, len(node.Modules)+1),
	}

	for idx, mod := range node.Modules {
		res.Parts[idx] = mod.Value
	}
	res.Parts[len(res.Parts)-1] = node.Name.Value

	return res
}

func JoinModuleName(modName *ModuleName, part string) *ModuleName {
	res := &ModuleName{
		Parts: make([]string, len(modName.Parts)+1),
	}
	copy(res.Parts, modName.Parts)
	res.Parts[len(res.Parts)-1] = part
	return res
}

func (v *ModuleName) Last() string {
	idx := len(v.Parts) - 1
	return v.Parts[idx]
}

func (v *ModuleName) String() string {
	buf := new(bytes.Buffer)
	for idx, parent := range v.Parts {
		buf.WriteString(parent)
		if idx < len(v.Parts)-1 {
			buf.WriteString("::")
		}

	}
	return buf.String()
}

func (v *ModuleName) ToPath() string {
	buf := new(bytes.Buffer)
	for idx, parent := range v.Parts {
		buf.WriteString(parent)
		if idx < len(v.Parts)-1 {
			buf.WriteRune(filepath.Separator)
		}

	}
	return buf.String()
}
