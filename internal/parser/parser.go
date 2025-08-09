package parser

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/elliot40404/mailc/internal/util"
)

// inferSimpleVariables scans the subject and HTML body for simple template
// variables like {{var}} and, if not already declared via @type, records them
// as top-level variables of type string.
func inferSimpleVariables(pt *ParsedTemplate) {
	// Build a set of existing variable names for quick lookup
	existing := map[string]struct{}{}
	for _, v := range pt.Variables {
		existing[v.Name] = struct{}{}
	}
	// Extract from subject and HTML
	candidates := make(map[string]struct{})
	for _, m := range reSimpleVar.FindAllStringSubmatch(pt.Subject, -1) {
		if len(m) > 1 {
			candidates[m[1]] = struct{}{}
		}
	}
	for _, m := range reSimpleVar.FindAllStringSubmatch(pt.HTML, -1) {
		if len(m) > 1 {
			candidates[m[1]] = struct{}{}
		}
	}
	// Add missing as string-typed variables
	for name := range candidates {
		if _, ok := existing[name]; ok {
			continue
		}
		// Skip names that collide with declared structs (since they would be ambiguous)
		collision := false
		for _, s := range pt.Structs {
			if util.UpperFirst(name) == s.Name {
				collision = true
				break
			}
		}
		if collision {
			continue
		}
		pt.Variables = append(pt.Variables, ParsedVariable{Name: name, Type: "string"})
		pt.Types = append(pt.Types, ParsedType{Type: "string"})
	}
}

type ParsedTemplate struct {
	FilePath  string
	Subject   string
	HTML      string
	Structs   []ParsedStruct
	Types     []ParsedType
	Variables []ParsedVariable
}

type ParsedStruct struct {
	Name   string
	Fields []ParsedField
}

type ParsedField struct {
	Name string
	Type string
}

type ParsedType struct {
	Type string
}

type ParsedVariable struct {
	Name string
	Type string
}

var (
	reSubject = regexp.MustCompile(`<!--\s*\$Subject:\s*(.*?)\s*-->`)
	reTypeDef = regexp.MustCompile(`<!--\s*@type\s+([A-Za-z0-9_.]+)\s*([A-Za-z0-9_.]*)\s*-->`)
)

// Matches simple variables like {{var}} or {{   var   }} (no dots/functions).
var reSimpleVar = regexp.MustCompile(`\{\{\s*-?\s*([A-Za-z][A-Za-z0-9_]*)\s*-?\s*\}\}`)

func ParseFile(path string) (*ParsedTemplate, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}

	pt := &ParsedTemplate{
		FilePath: path,
	}

	scanner := bufio.NewScanner(bytes.NewReader(data))
	structMap := make(map[string]*ParsedStruct)
	typeSet := make(map[string]struct{})
	htmlBuf := &bytes.Buffer{}

	for scanner.Scan() {
		line := scanner.Text()

		if m := reSubject.FindStringSubmatch(line); len(m) > 1 {
			pt.Subject = strings.TrimSpace(m[1])
			continue
		}

		if m := reTypeDef.FindStringSubmatch(line); len(m) > 0 {
			fullName := strings.TrimSpace(m[1])
			fieldType := strings.TrimSpace(m[2])

			if !strings.Contains(fullName, ".") {
				// No dot: either a struct declaration (no type) or a single variable (has type)
				if fieldType == "" {
					structName := util.UpperFirst(fullName)
					if _, exists := structMap[structName]; !exists {
						structMap[structName] = &ParsedStruct{Name: structName}
					}
				} else {
					// Single top-level variable
					pt.Variables = append(pt.Variables, ParsedVariable{
						Name: fullName,
						Type: fieldType,
					})
					typeSet[fieldType] = struct{}{}
				}
			} else {
				parts := strings.SplitN(fullName, ".", 2)
				structName := util.UpperFirst(parts[0])
				fieldName := util.UpperFirst(parts[1])

				if _, exists := structMap[structName]; !exists {
					structMap[structName] = &ParsedStruct{Name: structName}
				}

				structMap[structName].Fields = append(structMap[structName].Fields, ParsedField{
					Name: fieldName,
					Type: fieldType,
				})

				if fieldType != "" {
					typeSet[fieldType] = struct{}{}
				}
			}
			continue
		}

		htmlBuf.WriteString(line + "\n")
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scanning file: %w", err)
	}

	for _, s := range structMap {
		pt.Structs = append(pt.Structs, *s)
	}

	for t := range typeSet {
		pt.Types = append(pt.Types, ParsedType{Type: t})
	}

	pt.HTML = htmlBuf.String()

	// Infer undeclared simple variables from subject and HTML
	inferSimpleVariables(pt)
	return pt, nil
}

func ParseDir(dir string) ([]*ParsedTemplate, error) {
	var templates []*ParsedTemplate

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".html") {
			pt, err := ParseFile(path)
			if err != nil {
				return fmt.Errorf("parsing %s: %w", path, err)
			}
			templates = append(templates, pt)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return templates, nil
}
