package parser

import (
	"os"

	"github.com/ark-lang/ark/src/lexer"
	"github.com/ark-lang/ark/src/util"
	"github.com/ark-lang/ark/src/util/log"
)

type NodeType int

const (
	NODE_UNKOWN NodeType = iota
	NODE_PRIMITIVE
	NODE_FUNCTION
	NODE_ENUM
	NODE_ENUM_MEMBER
	NODE_STRUCT
	NODE_STRUCT_STATIC
	NODE_VARIABLE
	NODE_MODULE
	NODE_TYPE
)

func (v NodeType) IsType() bool {
	return v == NODE_PRIMITIVE || v == NODE_ENUM || v == NODE_STRUCT || v == NODE_TYPE
}

func (v NodeType) String() string {
	switch v {
	case NODE_PRIMITIVE:
		return "primitive"

	case NODE_FUNCTION:
		return "function"

	case NODE_ENUM:
		return "enum"

	case NODE_ENUM_MEMBER:
		return "enum member"

	case NODE_STRUCT:
		return "struct"

	case NODE_STRUCT_STATIC:
		return "static struct member"

	case NODE_VARIABLE:
		return "variable"

	case NODE_MODULE:
		return "module"

	default:
		return "unkown"
	}
}

type NameMap struct {
	parent  *NameMap
	tree    *ParseTree
	types   map[string]NodeType
	modules map[string]*NameMap
}

func (v *NameMap) typeOf(name LocatedString) NodeType {
	typ, ok := v.types[name.Value]
	if !ok && v.parent != nil {
		typ = v.parent.TypeOf(name)
	}
	return typ
}

func (v *NameMap) TypeOf(name LocatedString) NodeType {
	typ := v.typeOf(name)
	if typ == NODE_UNKOWN {
		startPos := name.Where.Start()
		log.Errorln("constructor", "[%s:%d:%d] Undeclared name `%s`",
			startPos.Filename, startPos.Line, startPos.Char, name.Value)
		log.Error("constructor", v.tree.Source.MarkSpan(name.Where))
	}
	return typ
}

func (v *NameMap) module(name LocatedString) *NameMap {
	mod, ok := v.modules[name.Value]
	if !ok && v.parent != nil {
		mod = v.parent.Module(name)
	}
	return mod
}

func (v *NameMap) Module(name LocatedString) *NameMap {
	mod := v.module(name)
	if mod == nil {
		startPos := name.Where.Start()
		log.Errorln("constructor", "[%s:%d:%d] Unknown module `%s`",
			startPos.Filename, startPos.Line, startPos.Char, name.Value)
		log.Error("constructor", v.tree.Source.MarkSpan(name.Where))
	}
	return mod
}

func (v *NameMap) TypeOfNameNode(name *NameNode) NodeType {
	mod := v
	typ := NODE_MODULE
	for _, modName := range name.Modules {
		if typ == NODE_MODULE {
			typ = mod.typeOf(modName)
			if typ == NODE_MODULE {
				mod = mod.module(modName)
				if mod == nil {
					return NODE_UNKOWN
				}
			}
		} else {
			startPos := modName.Where.Start()
			log.Errorln("constructor", "[%s:%d:%d] Invalid use of `::`. `%s` is not a module",
				startPos.Filename, startPos.Line, startPos.Char, modName.Value)
			log.Error("constructor", v.tree.Source.MarkSpan(modName.Where))
		}
	}

	if typ == NODE_ENUM {
		return NODE_ENUM_MEMBER
	} else if typ == NODE_STRUCT {
		return NODE_STRUCT_STATIC
	}

	typ = mod.typeOf(name.Name)
	if typ == NODE_UNKOWN {
		startPos := name.Name.Where.Start()
		log.Errorln("constructor", "[%s:%d:%d] Undeclared name `%s`",
			startPos.Filename, startPos.Line, startPos.Char, name.Name.Value)
		log.Error("constructor", v.tree.Source.MarkSpan(name.Name.Where))
	}
	return typ
}

func MapNames(nodes []ParseNode, tree *ParseTree, modules map[string]*ParseTree, parent *NameMap) *NameMap {
	nameMap := &NameMap{}
	nameMap.parent = parent
	nameMap.tree = tree
	nameMap.types = make(map[string]NodeType)
	nameMap.modules = make(map[string]*NameMap)
	previousLocation := make(map[string]lexer.Span)
	shouldExit := false

	var cModule *NameMap
	if parent == nil {
		for i := 0; i < len(_PrimitiveType_index); i++ {
			typ := PrimitiveType(i)
			name := typ.TypeName()
			nameMap.types[name] = NODE_PRIMITIVE
			previousLocation[name] = lexer.Span{Filename: "_builtin"}
		}
		cModule = &NameMap{
			types:   make(map[string]NodeType),
			modules: make(map[string]*NameMap),
		}
		nameMap.types["C"] = NODE_MODULE
		nameMap.modules["C"] = cModule
	} else {
		cModule = parent.module(LocatedString{Value: "C"})
	}

	for _, node := range nodes {
		var name LocatedString
		var typ NodeType

		switch node.(type) {
		case *FunctionDeclNode:
			fdn := node.(*FunctionDeclNode)
			name, typ = fdn.Header.Name, NODE_FUNCTION

			if fdn.Header.Attrs().Contains("c") || fdn.Attrs().Contains("c") {
				_, occupied := nameMap.modules["C"].types[name.Value]
				if occupied {
					startPos := name.Where.Start()
					log.Errorln("constructor", "[%s:%d:%d] Found duplicate definition of `%s`",
						tree.Source.Path, startPos.Line, startPos.Char, name.Value)
					log.Error("constructor", tree.Source.MarkSpan(name.Where))

					prevPos := previousLocation[name.Value]
					log.Errorln("constructor", "[%s:%d:%d] Previous declaration was here",
						tree.Source.Path, prevPos.StartLine, prevPos.StartChar)
					log.Error("constructor", tree.Source.MarkSpan(prevPos))
					shouldExit = true
				}
				cModule.types[name.Value] = typ
				previousLocation[name.Value] = name.Where
				continue
			}

		case *VarDeclNode:
			vd := node.(*VarDeclNode)
			name, typ = vd.Name, NODE_VARIABLE

		case *TypeDeclNode:
			vd := node.(*TypeDeclNode)
			name, typ = vd.Name, NODE_TYPE

			switch vd.Type.(type) {
			case *EnumTypeNode:
				typ = NODE_ENUM
			case *StructTypeNode:
				typ = NODE_STRUCT
			}

		case *UseDeclNode:
			udn := node.(*UseDeclNode)

			baseModuleName := udn.Module.Name
			if len(udn.Module.Modules) > 0 {
				baseModuleName = udn.Module.Modules[0]
			}

			mod, ok := modules[baseModuleName.Value]
			if !ok {
				startPos := baseModuleName.Where.Start()
				log.Errorln("constructor", "[%s:%d:%d] Unknown module `%s`",
					tree.Source.Path, startPos.Line, startPos.Char, baseModuleName.Value)
				log.Error("constructor", tree.Source.MarkSpan(baseModuleName.Where))
				shouldExit = true
			}
			nameMap.modules[baseModuleName.Value] = MapNames(mod.Nodes, mod, modules, nil)

			modNameMap := nameMap.modules[baseModuleName.Value]
			for _, mod := range udn.Module.Modules {
				nextMap, ok := modNameMap.modules[mod.Value]
				if !ok {
					startPos := mod.Where.Start()
					log.Errorln("constructor", "[%s:%d:%d] Unknown module `%s`",
						tree.Source.Path, startPos.Line, startPos.Char, mod.Value)
					log.Error("constructor", tree.Source.MarkSpan(mod.Where))
					shouldExit = true
					break
				}
				modNameMap = nextMap
			}
			name, typ = udn.Module.Name, NODE_MODULE

		default:
			continue
		}

		_, occupied := nameMap.types[name.Value]
		if occupied {
			startPos := name.Where.Start()
			log.Errorln("constructor", "[%s:%d:%d] Found duplicate definition of `%s`",
				tree.Source.Path, startPos.Line, startPos.Char, name.Value)
			log.Error("constructor", tree.Source.MarkSpan(name.Where))

			prevPos := previousLocation[name.Value]
			log.Errorln("constructor", "[%s:%d:%d] Previous declaration was here",
				tree.Source.Path, prevPos.StartLine, prevPos.StartChar)
			log.Error("constructor", tree.Source.MarkSpan(prevPos))
			shouldExit = true
		}
		nameMap.types[name.Value] = typ
		previousLocation[name.Value] = name.Where
	}

	if shouldExit {
		os.Exit(util.EXIT_FAILURE_CONSTRUCTOR)
	}

	return nameMap
}
