package generator

import (
	"go/ast"
	goparser "go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"

	mailparser "github.com/elliot40404/mailc/internal/parser"
)

func TestGenerateCode_SimpleAndInvite(t *testing.T) {
	// Create temp templates
	dir := t.TempDir()
	mustWrite := func(name, body string) {
		p := filepath.Join(dir, name)
		if err := os.WriteFile(p, []byte(body), 0o600); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}
	mustWrite("simple.html", `<!-- $Subject: Welcome {{username}} -->

<html>
<body>
Hi {{firstName}}
</body>
</html>
`)
	mustWrite("invite.html", `<!-- $Subject: Invite -->
<!-- @type inviteLink string -->
<html><body><a href="{{inviteLink}}"></a></body></html>`)

	pts, err := mailparser.ParseDir(dir)
	if err != nil {
		t.Fatalf("ParseDir: %v", err)
	}
	out := t.TempDir()
	if err := GenerateCode(pts, out, "TEST"); err != nil {
		t.Fatalf("GenerateCode: %v", err)
	}

	// Parse generated simple file via AST
	simplePath := filepath.Join(out, "simple.email.go")
	fset := token.NewFileSet()
	file, err := goparser.ParseFile(fset, simplePath, nil, goparser.ParseComments)
	if err != nil {
		t.Fatalf("parse generated simple: %v", err)
	}
	// Assert type SimpleEmailData with fields Username and FirstName
	hasUsername := false
	hasFirstName := false
	for _, decl := range file.Decls {
		gd, ok := decl.(*ast.GenDecl)
		if !ok || gd.Tok != token.TYPE {
			continue
		}
		for _, spec := range gd.Specs {
			ts, ok := spec.(*ast.TypeSpec)
			if !ok || ts.Name.Name != "SimpleEmailData" {
				continue
			}
			st, ok := ts.Type.(*ast.StructType)
			if !ok {
				continue
			}
			for _, f := range st.Fields.List {
				for _, n := range f.Names {
					if n.Name == "Username" {
						hasUsername = true
					}
					if n.Name == "FirstName" {
						hasFirstName = true
					}
				}
			}
		}
	}
	if !hasUsername || !hasFirstName {
		t.Fatalf("expected SimpleEmailData to include Username and FirstName fields")
	}
	// Assert normalized variables in const content
	htmlConstVal := findConstValue(t, file, "simpleEmailHTMLTemplate")
	if !strings.Contains(htmlConstVal, "{{ .FirstName}}") {
		t.Fatalf("expected body template to contain normalized FirstName reference, got: %q", htmlConstVal)
	}
	subjConstVal := findConstValue(t, file, "simpleEmailSubjectTemplate")
	if !strings.Contains(subjConstVal, "{{ .Username}}") {
		t.Fatalf("expected subject template to contain normalized Username reference, got: %q", subjConstVal)
	}

	// Invite assertions
	invitePath := filepath.Join(out, "invite.email.go")
	file2, err := goparser.ParseFile(fset, invitePath, nil, 0)
	if err != nil {
		t.Fatalf("parse generated invite: %v", err)
	}
	if !typeHasField(file2, "InviteEmailData", "InviteLink") {
		t.Fatalf("expected InviteEmailData to include InviteLink string field")
	}
	if findConstValue(t, file2, "inviteEmailHTMLTemplate") == "" || findConstValue(t, file2, "inviteEmailSubjectTemplate") == "" {
		t.Fatalf("expected invite to have file-specific const names")
	}
}

func TestGenerateCode_NoSubject_NoTextTemplateImport(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nosubject.html")
	if err := os.WriteFile(path, []byte("<html>\n<body>\nHi {{firstName}}\n</body>\n</html>\n"), 0o600); err != nil {
		t.Fatalf("write nosubject: %v", err)
	}
	pts, err := mailparser.ParseDir(dir)
	if err != nil {
		t.Fatalf("ParseDir: %v", err)
	}
	out := t.TempDir()
	if err := GenerateCode(pts, out, "TEST"); err != nil {
		t.Fatalf("GenerateCode: %v", err)
	}
	genPath := filepath.Join(out, "nosubject.email.go")
	fset := token.NewFileSet()
	file, err := goparser.ParseFile(fset, genPath, nil, 0)
	if err != nil {
		t.Fatalf("parse generated nosubject: %v", err)
	}
	// Ensure no text/template import
	for _, imp := range file.Imports {
		if imp.Path != nil && imp.Path.Value == "\"text/template\"" {
			t.Fatalf("text/template should not be imported when subject is empty")
		}
	}
	// Ensure result struct exists
	if !typeHasField(file, "NosubjectEmailResult", "Subject") || !typeHasField(file, "NosubjectEmailResult", "HTML") {
		t.Fatalf("expected result struct to be generated")
	}
}

// Helpers

func findConstValue(t *testing.T, f *ast.File, name string) string {
	t.Helper()
	for _, decl := range f.Decls {
		gd, ok := decl.(*ast.GenDecl)
		if !ok || gd.Tok != token.CONST {
			continue
		}
		for _, spec := range gd.Specs {
			vs, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}
			for i, n := range vs.Names {
				if n.Name == name && i < len(vs.Values) {
					bl, ok := vs.Values[i].(*ast.BasicLit)
					if ok {
						// strip surrounding quotes/backticks
						return strings.Trim(bl.Value, "`\"")
					}
				}
			}
		}
	}
	return ""
}

func typeHasField(f *ast.File, typeName, fieldName string) bool {
	for _, decl := range f.Decls {
		gd, ok := decl.(*ast.GenDecl)
		if !ok || gd.Tok != token.TYPE {
			continue
		}
		for _, spec := range gd.Specs {
			ts, ok := spec.(*ast.TypeSpec)
			if !ok || ts.Name.Name != typeName {
				continue
			}
			st, ok := ts.Type.(*ast.StructType)
			if !ok {
				continue
			}
			for _, f := range st.Fields.List {
				for _, n := range f.Names {
					if n.Name == fieldName {
						return true
					}
				}
			}
		}
	}
	return false
}
