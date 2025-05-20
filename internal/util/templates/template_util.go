package templates

import (
	"bytes"
	"errors"
	"html/template"
	"io/fs"
	"path/filepath"
	"strings"
)

// Errors
var (
	ErrTemplateNotFound = errors.New("template not found")
)

// TemplateUtil provides utility functions for template operations
type TemplateUtil struct {
	funcMap template.FuncMap
}

// NewTemplateUtil creates a new TemplateUtil
func NewTemplateUtil() *TemplateUtil {
	return &TemplateUtil{
		funcMap: make(template.FuncMap),
	}
}

// AddFunc adds a function to the template's function map
func (t *TemplateUtil) AddFunc(name string, fn interface{}) {
	if t.funcMap == nil {
		t.funcMap = make(template.FuncMap)
	}
	t.funcMap[name] = fn
}

// AddFuncMap adds multiple functions to the template's function map
func (t *TemplateUtil) AddFuncMap(funcMap template.FuncMap) {
	if t.funcMap == nil {
		t.funcMap = make(template.FuncMap)
	}
	for name, fn := range funcMap {
		t.funcMap[name] = fn
	}
}

// Execute executes a template with the given data
func (t *TemplateUtil) Execute(templateStr string, data interface{}) (string, error) {
	tmpl, err := template.New("").Funcs(t.funcMap).Parse(templateStr)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// ExecuteFile executes a template file with the given data
func (t *TemplateUtil) ExecuteFile(filename string, data interface{}) (string, error) {
	tmpl, err := template.New(filepath.Base(filename)).Funcs(t.funcMap).ParseFiles(filename)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// ExecuteFiles executes multiple template files with the given data
func (t *TemplateUtil) ExecuteFiles(filenames []string, data interface{}) (string, error) {
	tmpl, err := template.New(filepath.Base(filenames[0])).Funcs(t.funcMap).ParseFiles(filenames...)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// ExecuteGlob executes templates matching a glob pattern with the given data
func (t *TemplateUtil) ExecuteGlob(pattern string, data interface{}) (string, error) {
	tmpl, err := template.New("").Funcs(t.funcMap).ParseGlob(pattern)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// ExecuteTemplate executes a named template with the given data
func (t *TemplateUtil) ExecuteTemplate(tmpl *template.Template, name string, data interface{}) (string, error) {
	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, name, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// Parse parses a template string
func (t *TemplateUtil) Parse(templateStr string) (*template.Template, error) {
	return template.New("").Funcs(t.funcMap).Parse(templateStr)
}

// ParseFile parses a template file
func (t *TemplateUtil) ParseFile(filename string) (*template.Template, error) {
	return template.New(filepath.Base(filename)).Funcs(t.funcMap).ParseFiles(filename)
}

// ParseFiles parses multiple template files
func (t *TemplateUtil) ParseFiles(filenames ...string) (*template.Template, error) {
	return template.New(filepath.Base(filenames[0])).Funcs(t.funcMap).ParseFiles(filenames...)
}

// ParseGlob parses templates matching a glob pattern
func (t *TemplateUtil) ParseGlob(pattern string) (*template.Template, error) {
	return template.New("").Funcs(t.funcMap).ParseGlob(pattern)
}

// MustParse parses a template string and panics if there is an error
func (t *TemplateUtil) MustParse(templateStr string) *template.Template {
	return template.Must(template.New("").Funcs(t.funcMap).Parse(templateStr))
}

// MustParseFile parses a template file and panics if there is an error
func (t *TemplateUtil) MustParseFile(filename string) *template.Template {
	return template.Must(template.New(filepath.Base(filename)).Funcs(t.funcMap).ParseFiles(filename))
}

// MustParseFiles parses multiple template files and panics if there is an error
func (t *TemplateUtil) MustParseFiles(filenames ...string) *template.Template {
	return template.Must(template.New(filepath.Base(filenames[0])).Funcs(t.funcMap).ParseFiles(filenames...))
}

// MustParseGlob parses templates matching a glob pattern and panics if there is an error
func (t *TemplateUtil) MustParseGlob(pattern string) *template.Template {
	return template.Must(template.New("").Funcs(t.funcMap).ParseGlob(pattern))
}

// TemplateSet represents a set of templates
type TemplateSet struct {
	templates map[string]*template.Template
	funcMap   template.FuncMap
}

// NewTemplateSet creates a new TemplateSet
func (t *TemplateUtil) NewTemplateSet() *TemplateSet {
	return &TemplateSet{
		templates: make(map[string]*template.Template),
		funcMap:   t.funcMap,
	}
}

// Add adds a template to the set
func (ts *TemplateSet) Add(name, templateStr string) error {
	tmpl, err := template.New(name).Funcs(ts.funcMap).Parse(templateStr)
	if err != nil {
		return err
	}
	ts.templates[name] = tmpl
	return nil
}

// AddFile adds a template file to the set
func (ts *TemplateSet) AddFile(name, filename string) error {
	tmpl, err := template.New(name).Funcs(ts.funcMap).ParseFiles(filename)
	if err != nil {
		return err
	}
	ts.templates[name] = tmpl
	return nil
}

// Execute executes a template from the set
func (ts *TemplateSet) Execute(name string, data interface{}) (string, error) {
	tmpl, ok := ts.templates[name]
	if !ok {
		return "", ErrTemplateNotFound
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// ExecuteTemplate executes a named template from the set
func (ts *TemplateSet) ExecuteTemplate(name, templateName string, data interface{}) (string, error) {
	tmpl, ok := ts.templates[name]
	if !ok {
		return "", ErrTemplateNotFound
	}

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, templateName, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// LoadTemplates loads templates from a filesystem
func (t *TemplateUtil) LoadTemplates(fsys fs.FS, pattern string) (*TemplateSet, error) {
	ts := t.NewTemplateSet()
	err := fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		matched, err := filepath.Match(pattern, path)
		if err != nil {
			return err
		}
		if !matched {
			return nil
		}
		content, err := fs.ReadFile(fsys, path)
		if err != nil {
			return err
		}
		name := strings.TrimSuffix(path, filepath.Ext(path))
		return ts.Add(name, string(content))
	})
	if err != nil {
		return nil, err
	}
	return ts, nil
}
