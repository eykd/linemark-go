package archtest_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

const moduleRoot = "github.com/eykd/linemark-go"

// Architectural layers from inner to outer.
const (
	layerDomain         = "domain"
	layerApplication    = "application"
	layerInfrastructure = "infrastructure"
	layerPresentation   = "presentation"
)

// packageLayer maps relative package paths to their architectural layer.
var packageLayer = map[string]string{
	"internal/domain":      layerDomain,
	"internal/outline":     layerApplication,
	"internal/frontmatter": layerInfrastructure,
	"internal/lock":        layerInfrastructure,
	"internal/slug":        layerInfrastructure,
	"internal/sid":         layerInfrastructure,
	"internal/fs":          layerInfrastructure,
	"cmd":                  layerPresentation,
}

// allowedImports defines the dependency matrix per the clean architecture rules:
//
//	Domain         → Domain only
//	Application    → Domain, Application
//	Infrastructure → Domain, Application, Infrastructure
//	Presentation   → Domain, Application, Infrastructure, Presentation
var allowedImports = map[string]map[string]bool{
	layerDomain:         {layerDomain: true},
	layerApplication:    {layerDomain: true, layerApplication: true},
	layerInfrastructure: {layerDomain: true, layerApplication: true, layerInfrastructure: true},
	layerPresentation:   {layerDomain: true, layerApplication: true, layerInfrastructure: true, layerPresentation: true},
}

// projectRoot returns the absolute path to the project root by navigating
// up from the test file location (internal/archtest/).
func projectRoot(t *testing.T) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("unable to determine test file path")
	}
	return filepath.Join(filepath.Dir(filename), "..", "..")
}

// collectInternalImports parses all non-test Go files in dir and returns
// the project-internal import paths (those starting with moduleRoot).
func collectInternalImports(t *testing.T, dir string) []string {
	t.Helper()
	fset := token.NewFileSet()
	//lint:ignore SA1019 ParseDir is sufficient for import scanning in tests
	pkgs, err := parser.ParseDir(fset, dir, func(fi os.FileInfo) bool {
		return !strings.HasSuffix(fi.Name(), "_test.go")
	}, parser.ImportsOnly)
	if err != nil {
		t.Fatalf("parsing %s: %v", dir, err)
	}

	var imports []string
	for _, pkg := range pkgs {
		for _, file := range pkg.Files {
			for _, imp := range file.Imports {
				path := strings.Trim(imp.Path.Value, `"`)
				if strings.HasPrefix(path, moduleRoot+"/") {
					imports = append(imports, path)
				}
			}
		}
	}
	return imports
}

// collectAllImports parses all non-test Go files in dir and returns
// every import path (internal and external).
func collectAllImports(t *testing.T, dir string) []string {
	t.Helper()
	fset := token.NewFileSet()
	//lint:ignore SA1019 ParseDir is sufficient for import scanning in tests
	pkgs, err := parser.ParseDir(fset, dir, func(fi os.FileInfo) bool {
		return !strings.HasSuffix(fi.Name(), "_test.go")
	}, parser.ImportsOnly)
	if err != nil {
		t.Fatalf("parsing %s: %v", dir, err)
	}

	var imports []string
	for _, pkg := range pkgs {
		for _, file := range pkg.Files {
			for _, imp := range file.Imports {
				path := strings.Trim(imp.Path.Value, `"`)
				imports = append(imports, path)
			}
		}
	}
	return imports
}

// relPackage strips the module root prefix to get a relative package path.
func relPackage(importPath string) string {
	return strings.TrimPrefix(importPath, moduleRoot+"/")
}

// TestDomainLayerHasNoInternalDependencies verifies the domain package
// imports only Go standard library packages, no other project packages.
func TestDomainLayerHasNoInternalDependencies(t *testing.T) {
	root := projectRoot(t)
	domainDir := filepath.Join(root, "internal", "domain")
	imports := collectInternalImports(t, domainDir)

	for _, imp := range imports {
		t.Errorf("domain layer has forbidden internal import: %s", imp)
	}
}

// TestApplicationLayerDoesNotImportInfrastructure verifies the application
// layer (outline) does not directly depend on infrastructure packages.
func TestApplicationLayerDoesNotImportInfrastructure(t *testing.T) {
	root := projectRoot(t)
	outlineDir := filepath.Join(root, "internal", "outline")
	imports := collectInternalImports(t, outlineDir)

	for _, imp := range imports {
		rel := relPackage(imp)
		targetLayer, ok := packageLayer[rel]
		if !ok {
			continue
		}
		if targetLayer == layerInfrastructure {
			t.Errorf("application layer (outline) imports infrastructure package: %s", rel)
		}
	}
}

// TestLayerDependencyCompliance checks every package's imports against the
// full dependency matrix. Each package may only import packages in layers
// at the same level or below.
func TestLayerDependencyCompliance(t *testing.T) {
	root := projectRoot(t)

	for pkgPath, sourceLayer := range packageLayer {
		dir := filepath.Join(root, pkgPath)
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			continue
		}

		imports := collectInternalImports(t, dir)
		allowed := allowedImports[sourceLayer]

		for _, imp := range imports {
			rel := relPackage(imp)
			targetLayer, ok := packageLayer[rel]
			if !ok {
				continue
			}
			if !allowed[targetLayer] {
				t.Errorf("layer violation: %s (%s layer) imports %s (%s layer)",
					pkgPath, sourceLayer, rel, targetLayer)
			}
		}
	}
}

// fileContainsIdent parses a Go source file and returns true if any identifier
// in the AST matches the given name.
func fileContainsIdent(t *testing.T, path, name string) bool {
	t.Helper()
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, nil, parser.AllErrors)
	if err != nil {
		t.Fatalf("parsing %s: %v", path, err)
	}

	found := false
	ast.Inspect(f, func(n ast.Node) bool {
		if ident, ok := n.(*ast.Ident); ok && ident.Name == name {
			found = true
			return false
		}
		return true
	})
	return found
}

// TestMainPrintsErrorsWhenSilenceErrorsSet verifies that when SilenceErrors
// is set on the root command, main.go calls FormatError to print errors to
// stderr. Without this, CLI errors are silently swallowed.
func TestMainPrintsErrorsWhenSilenceErrorsSet(t *testing.T) {
	root := projectRoot(t)

	// Check if SilenceErrors is set in cmd/root.go
	if !fileContainsIdent(t, filepath.Join(root, "cmd", "root.go"), "SilenceErrors") {
		t.Skip("SilenceErrors not used")
	}

	// Verify main.go calls FormatError
	mainFile := filepath.Join(root, "main.go")
	if !fileContainsIdent(t, mainFile, "FormatError") {
		t.Error("main.go must call FormatError when SilenceErrors is true on root command — " +
			"otherwise CLI errors are silently swallowed")
	}
}

// TestExternalDependencyContainment verifies that third-party dependencies
// are only imported in their designated wrapper packages, not leaked across
// the codebase.
func TestExternalDependencyContainment(t *testing.T) {
	// Each external dependency maps to the one package allowed to import it.
	containment := map[string]string{
		"gopkg.in/yaml.v3":               "internal/frontmatter",
		"github.com/gofrs/flock":         "internal/lock",
		"golang.org/x/text/unicode/norm": "internal/slug",
		"github.com/spf13/cobra":         "cmd",
	}

	root := projectRoot(t)

	for pkgPath := range packageLayer {
		dir := filepath.Join(root, pkgPath)
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			continue
		}

		imports := collectAllImports(t, dir)
		for _, imp := range imports {
			allowedPkg, tracked := containment[imp]
			if !tracked {
				continue
			}
			if pkgPath != allowedPkg {
				t.Errorf("external dependency %q imported in %s (should only be in %s)",
					imp, pkgPath, allowedPkg)
			}
		}
	}
}
