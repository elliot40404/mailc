package parser

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseFile_SimpleVariablesInferred(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "simple.html")
	err := os.WriteFile(path, []byte(`<!-- $Subject: Hello {{username}} -->
<html><body><p>Welcome {{firstName}}</p></body></html>`), 0o600)
	if err != nil {
		t.Fatalf("write temp template: %v", err)
	}
	pt, err := ParseFile(path)
	if err != nil {
		t.Fatalf("ParseFile error: %v", err)
	}
	if pt.Subject == "" {
		t.Fatalf("expected subject to be parsed")
	}
	got := map[string]bool{}
	for _, v := range pt.Variables {
		got[v.Name] = true
	}
	if !got["username"] || !got["firstName"] {
		t.Fatalf("expected variables username and firstName to be inferred, got: %#v", got)
	}
}

func TestParseFile_SingleTopLevelVariableWithTypeHint(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "invite.html")
	err := os.WriteFile(path, []byte(`<!-- $Subject: Use the link -->
<!-- @type inviteLink string -->
<html><body><a href="{{inviteLink}}">link</a></body></html>`), 0o600)
	if err != nil {
		t.Fatalf("write temp template: %v", err)
	}
	pt, err := ParseFile(path)
	if err != nil {
		t.Fatalf("ParseFile error: %v", err)
	}
	found := false
	for _, v := range pt.Variables {
		if v.Name == "inviteLink" && v.Type == "string" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected top-level variable inviteLink string to be parsed")
	}
}

func TestParseFile_StructsAndFields(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "order.html")
	tpl := `<!-- $Subject: Welcome {{User.Name}} -->
<!-- @type Order -->
<!-- @type Order.ID int -->
<!-- @type Order.Name string -->
<!-- @type Order.Qty int -->
<!-- @type Order.CreatedAt string -->
<!-- @type User -->
<!-- @type User.Name string -->
<html><body>
{{User.Name}} #{{Order.ID}} {{Order.Name}} x{{Order.Qty}} {{Order.CreatedAt}}
</body></html>`
	if err := os.WriteFile(path, []byte(tpl), 0o600); err != nil {
		t.Fatalf("write temp template: %v", err)
	}
	pt, err := ParseFile(path)
	if err != nil {
		t.Fatalf("ParseFile error: %v", err)
	}
	// Expect two structs
	if len(pt.Structs) != 2 {
		t.Fatalf("expected 2 structs, got %d", len(pt.Structs))
	}
	// Verify Order fields
	var order ParsedStruct
	for _, s := range pt.Structs {
		if s.Name == "Order" {
			order = s
		}
	}
	if len(order.Fields) != 4 {
		t.Fatalf("expected 4 fields in Order, got %d", len(order.Fields))
	}
	// quick presence checks
	wantFields := map[string]string{"ID": "int", "Name": "string", "Qty": "int", "CreatedAt": "string"}
	for _, f := range order.Fields {
		if typ, ok := wantFields[f.Name]; !ok || f.Type != typ {
			t.Fatalf("unexpected field: %s %s", f.Name, f.Type)
		}
	}
}
