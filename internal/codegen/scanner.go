package codegen

import (
	"go/ast"
	"go/parser"
	"go/token"
	"reflect"
	"strings"
)

// FieldInfo contains metadata for a struct field.
type FieldInfo struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	JSONName string `json:"jsonName"`
}

// StructInfo describes a scanned Go struct.
type StructInfo struct {
	Name   string      `json:"name"`
	Fields []FieldInfo `json:"fields"`
}

// ActionInfo describes a function marked with // +zyraaction.
type ActionInfo struct {
	Package    string `json:"package"`
	Name       string `json:"name"`
	InputType  string `json:"inputType"`
	OutputType string `json:"outputType"`
}

// ScanResult holds the collected structs and actions.
type ScanResult struct {
	Structs []StructInfo `json:"structs"`
	Actions []ActionInfo `json:"actions"`
}

// ScanSource parses Go source code content and extracts structs and actions.
func ScanSource(filename string, src string) (*ScanResult, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filename, src, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	result := &ScanResult{}
	pkgName := node.Name.Name

	// Map comments to AST nodes
	commentMap := ast.NewCommentMap(fset, node, node.Comments)

	ast.Inspect(node, func(n ast.Node) bool {
		switch decl := n.(type) {
		case *ast.GenDecl:
			if decl.Tok == token.TYPE {
				for _, spec := range decl.Specs {
					if typeSpec, ok := spec.(*ast.TypeSpec); ok {
						if structType, ok := typeSpec.Type.(*ast.StructType); ok {
							st := parseStruct(typeSpec.Name.Name, structType)
							result.Structs = append(result.Structs, st)
						}
					}
				}
			}
		case *ast.FuncDecl:
			// Check if func has comment containing // +zyraaction
			comments := commentMap[decl]
			isAction := false
			if decl.Doc != nil {
				for _, comment := range decl.Doc.List {
					if strings.Contains(comment.Text, "+zyraaction") {
						isAction = true
						break
					}
				}
			}
			if !isAction {
				for _, cg := range comments {
					for _, comment := range cg.List {
						if strings.Contains(comment.Text, "+zyraaction") {
							isAction = true
							break
						}
					}
				}
			}

			if isAction {
				action := parseAction(pkgName, decl)
				result.Actions = append(result.Actions, action)
			}
		}
		return true
	})

	return result, nil
}

func parseStruct(name string, structType *ast.StructType) StructInfo {
	info := StructInfo{Name: name}
	for _, field := range structType.Fields.List {
		if len(field.Names) == 0 {
			continue
		}
		fieldName := field.Names[0].Name
		jsonTag := parseJSONTag(field.Tag)
		if jsonTag == "-" {
			continue
		}
		if jsonTag == "" {
			jsonTag = fieldName
		}

		fieldType := exprToString(field.Type)
		info.Fields = append(info.Fields, FieldInfo{
			Name:     fieldName,
			Type:     fieldType,
			JSONName: jsonTag,
		})
	}
	return info
}

func parseAction(pkg string, funcDecl *ast.FuncDecl) ActionInfo {
	action := ActionInfo{
		Package:    pkg,
		Name:       funcDecl.Name.Name,
		InputType:  "void",
		OutputType: "void",
	}

	if funcDecl.Type.Params != nil && len(funcDecl.Type.Params.List) > 0 {
		// Assume first non-context parameter is input
		for _, param := range funcDecl.Type.Params.List {
			paramTypeStr := exprToString(param.Type)
			if !strings.Contains(paramTypeStr, "Context") {
				action.InputType = paramTypeStr
				break
			}
		}
	}

	if funcDecl.Type.Results != nil && len(funcDecl.Type.Results.List) > 0 {
		// First non-error result is output
		for _, result := range funcDecl.Type.Results.List {
			resTypeStr := exprToString(result.Type)
			if !strings.Contains(resTypeStr, "error") {
				action.OutputType = resTypeStr
				break
			}
		}
	}

	return action
}

func parseJSONTag(tag *ast.BasicLit) string {
	if tag == nil {
		return ""
	}
	val := reflect.StructTag(strings.Trim(tag.Value, "`")).Get("json")
	parts := strings.Split(val, ",")
	return parts[0]
}

func exprToString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return exprToString(t.X)
	case *ast.ArrayType:
		return "[]" + exprToString(t.Elt)
	case *ast.SelectorExpr:
		return exprToString(t.X) + "." + t.Sel.Name
	case *ast.MapType:
		return "map[" + exprToString(t.Key) + "]" + exprToString(t.Value)
	default:
		return "any"
	}
}
