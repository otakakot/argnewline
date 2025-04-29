package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

type modification struct {
	start       int
	end         int
	replacement string
}

func main() {
	flag.Parse()
	if flag.NArg() < 1 {
		fmt.Println("Usage: argnewline <path-or-file>")
		os.Exit(1)
	}

	path := flag.Arg(0)
	info, err := os.Stat(path)
	if err != nil {
		fmt.Printf("Error accessing path: %v\n", err)
		os.Exit(1)
	}

	var files []string
	if info.IsDir() {
		err = filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() && info.Name() == "vendor" {
				return filepath.SkipDir
			}
			if !info.IsDir() && strings.HasSuffix(p, ".go") {
				files = append(files, p)
			}
			return nil
		})
		if err != nil {
			fmt.Printf("Error walking directory: %v\n", err)
			os.Exit(1)
		}
	} else {
		files = []string{path}
	}

	for _, file := range files {
		processFile(file)
	}
}

func processFile(filename string) {
	fmt.Printf("processing file: %s\n", filename)

	src, err := os.ReadFile(filename)
	if err != nil {
		fmt.Printf("Error reading file %s: %v\n", filename, err)
		return
	}

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filename, src, parser.ParseComments)
	if err != nil {
		fmt.Printf("Error parsing file %s: %v\n", filename, err)
		return
	}

	var modifications []modification

	// Walk the AST and reformat function parameter lists, call arguments, and interface method parameters that are on a single line.
	ast.Inspect(f, func(n ast.Node) bool {
		// Process function declarations.
		if funcDecl, ok := n.(*ast.FuncDecl); ok {
			if funcDecl.Type.Params == nil || len(funcDecl.Type.Params.List) == 0 {
				return true
			}

			lparen := funcDecl.Type.Params.Opening
			rparen := funcDecl.Type.Params.Closing

			posL := fset.Position(lparen)
			posR := fset.Position(rparen)

			// Only consider parameters that are declared on one line.
			if posL.Line == posR.Line {
				var buf bytes.Buffer
				buf.WriteString("(\n")
				for _, field := range funcDecl.Type.Params.List {
					if len(field.Names) > 0 {
						var names []string
						for _, name := range field.Names {
							var nameBuf bytes.Buffer
							printer.Fprint(&nameBuf, fset, name)
							names = append(names, nameBuf.String())
						}
						var typeBuf bytes.Buffer
						printer.Fprint(&typeBuf, fset, field.Type)
						fieldStr := fmt.Sprintf("%s %s", strings.Join(names, ", "), strings.TrimSpace(typeBuf.String()))
						buf.WriteString("  " + fieldStr + ",\n")
					} else {
						var typeBuf bytes.Buffer
						printer.Fprint(&typeBuf, fset, field.Type)
						fieldStr := strings.TrimSpace(typeBuf.String())
						buf.WriteString("  " + fieldStr + ",\n")
					}
				}
				buf.WriteString(")")
				newParams := buf.String()
				startOffset := fset.Position(lparen).Offset
				endOffset := fset.Position(rparen).Offset + 1 // include the closing parenthesis
				modifications = append(modifications, modification{start: startOffset, end: endOffset, replacement: newParams})
			}
		}

		// Process function call expressions.
		if callExpr, ok := n.(*ast.CallExpr); ok {
			// Check if the call expression has valid parentheses.
			if callExpr.Lparen.IsValid() && callExpr.Rparen.IsValid() && len(callExpr.Args) > 0 {
				posL := fset.Position(callExpr.Lparen)
				posR := fset.Position(callExpr.Rparen)
				// Only reformat if the arguments are declared on one line.
				if posL.Line == posR.Line {
					var buf bytes.Buffer
					buf.WriteString("(\n")
					for _, arg := range callExpr.Args {
						var argBuf bytes.Buffer
						printer.Fprint(&argBuf, fset, arg)
						buf.WriteString("  " + strings.TrimSpace(argBuf.String()) + ",\n")
					}
					buf.WriteString(")")
					newArgs := buf.String()
					startOffset := fset.Position(callExpr.Lparen).Offset
					endOffset := fset.Position(callExpr.Rparen).Offset + 1 // include the closing parenthesis
					modifications = append(modifications, modification{start: startOffset, end: endOffset, replacement: newArgs})
				}
			}
		}

		// Process interface method definitions.
		if iface, ok := n.(*ast.InterfaceType); ok {
			for _, field := range iface.Methods.List {
				ft, ok := field.Type.(*ast.FuncType)
				if !ok || ft.Params == nil || len(ft.Params.List) == 0 {
					continue
				}
				lparen := ft.Params.Opening
				rparen := ft.Params.Closing
				posL := fset.Position(lparen)
				posR := fset.Position(rparen)
				if posL.Line == posR.Line {
					var buf bytes.Buffer
					buf.WriteString("(\n")
					for _, param := range ft.Params.List {
						if len(param.Names) > 0 {
							var names []string
							for _, name := range param.Names {
								var nameBuf bytes.Buffer
								printer.Fprint(&nameBuf, fset, name)
								names = append(names, nameBuf.String())
							}
							var typeBuf bytes.Buffer
							printer.Fprint(&typeBuf, fset, param.Type)
							fieldStr := fmt.Sprintf("%s %s", strings.Join(names, ", "), strings.TrimSpace(typeBuf.String()))
							buf.WriteString("  " + fieldStr + ",\n")
						} else {
							var typeBuf bytes.Buffer
							printer.Fprint(&typeBuf, fset, param.Type)
							fieldStr := strings.TrimSpace(typeBuf.String())
							buf.WriteString("  " + fieldStr + ",\n")
						}
					}
					buf.WriteString(")")
					newParams := buf.String()
					startOffset := fset.Position(lparen).Offset
					endOffset := fset.Position(rparen).Offset + 1
					modifications = append(modifications, modification{start: startOffset, end: endOffset, replacement: newParams})
				}
			}
		}
		return true
	})

	if len(modifications) == 0 {
		return
	}

	newSrc := string(src)
	// Apply modifications in reverse order to maintain correct offsets.
	for i := len(modifications) - 1; i >= 0; i-- {
		mod := modifications[i]
		newSrc = newSrc[:mod.start] + mod.replacement + newSrc[mod.end:]
	}

	// Format the entire source file.
	formatted, err := format.Source([]byte(newSrc))
	if err != nil {
		fmt.Printf("Error formatting file %s: %v\n", filename, err)
		return
	}

	if err = os.WriteFile(filename, formatted, 0644); err != nil {
		fmt.Printf("Error writing file %s: %v\n", filename, err)
		return
	}
}
