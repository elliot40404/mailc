package model

// TemplateFile is the parsed template representation.
type TemplateFile struct {
	Name            string     // filename without extension, e.g. "welcome"
	SubjectTemplate string     // extracted subject template (without <!-- -->), may be empty
	Body            string     // HTML body with the subject comment removed
	Variables       []Variable // parsed variables (from subject + body)
}

type Variable struct {
	Name     string
	Type     string     // e.g. "string", "int", "[]Item"
	Children []Variable // nested fields -> structs
	IsSlice  bool
}
