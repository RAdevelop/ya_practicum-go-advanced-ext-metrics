package main

import (
	"go/ast"
	"strings"

	"golang.org/x/tools/go/analysis"
)

const modulePrefix = "github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/"

// noExitAnalyzer запрещает прямой вызов os.Exit в функции main пакета main.
//
// Почему: os.Exit не выполняет defer-функции, не даёт graceful shutdown,
// не закрывает соединения и файлы. В main это особенно опасно — сервер
// завершается аварийно, без возможности корректно освободить ресурсы.
var noExitAnalyzer = &analysis.Analyzer{
	Name: "noexitmain",
	Doc:  "запрещает прямой вызов os.Exit в функции main пакета main",
	Run:  runNoExit,
}

func runNoExit(pass *analysis.Pass) (any, error) {
	// Анализатор применяется только к пакетам main.
	if pass.Pkg.Name() != "main" {
		return nil, nil
	}

	// тестовые бинарники (go test) — не наш код
	if strings.HasSuffix(pass.Pkg.Path(), ".test") {
		return nil, nil
	}

	// проверяем только пакеты нашего модуля
	if !strings.HasPrefix(pass.Pkg.Path(), modulePrefix) {
		return nil, nil
	}

	for _, file := range pass.Files {
		// Проверяем, что os импортирован в этом файле.
		osName, ok := findImportName(file, "os")
		if !ok {
			continue
		}

		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok {
				continue
			}
			// Ищем функцию main без ресивера (не метод).
			if fn.Recv != nil || fn.Name.Name != "main" {
				continue
			}

			ast.Inspect(fn.Body, func(n ast.Node) bool {
				call, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}

				sel, ok := call.Fun.(*ast.SelectorExpr)
				if !ok {
					return true
				}

				ident, ok := sel.X.(*ast.Ident)
				if !ok {
					return true
				}

				// os.Exit — с учётом возможного алиаса пакета os.
				if ident.Name == osName && sel.Sel.Name == "Exit" {
					pass.Reportf(call.Pos(),
						"прямой вызов os.Exit в main.main запрещён: "+
							"используйте return с кодом ошибки или log.Fatal через обёртку")
				}
				return true
			})
		}
	}

	return nil, nil
}

// findImportName возвращает имя, под которым импортирован пакет path,
// и признак, что пакет действительно импортирован.
// Учитывает алиасы (`import foo "os"`) и точечный импорт.
func findImportName(file *ast.File, path string) (string, bool) {
	for _, imp := range file.Imports {
		importPath := strings.Trim(imp.Path.Value, `"`)
		if importPath != path {
			continue
		}

		// import "os" — имя "os"
		// import foo "os" — имя "foo"
		// import . "os" — имя ".", нужно искать вызовы без квалификатора
		// import _ "os" — пакет не используется напрямую
		if imp.Name == nil {
			return path, true
		}
		switch imp.Name.Name {
		case "_":
			return "", false
		case ".":
			// точечный импорт: os.Exit вызывается как Exit()
			return ".", true
		default:
			return imp.Name.Name, true
		}
	}
	return "", false
}
