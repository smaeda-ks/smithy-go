// +build ignore

package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"text/template"
)

// Import provides the import path and optional alias
type Import struct {
	Path  string
	Alias string
}

// Package returns the Go package name for the import. Returns alias if set.
func (i Import) Package() string {
	if v := i.Alias; len(v) != 0 {
		return v
	}

	if v := i.Path; len(v) != 0 {
		parts := strings.Split(v, "/")
		pkg := parts[len(parts)-1]
		return pkg
	}

	return ""
}

// Scalar provides the definition of a type to generate pointer utilities for.
type Scalar struct {
	Type   string
	Import *Import
}

// Name returns the exported function name for the type.
func (t Scalar) Name() string {
	return strings.Title(t.Type)
}

// Symbol returns the scalar's Go symbol with path if needed.
func (t Scalar) Symbol() string {
	if t.Import != nil {
		return t.Import.Package() + "." + t.Type
	}
	return t.Type
}

// Scalars is a list of scalars.
type Scalars []Scalar

// Imports returns all imports for the scalars.
func (ts Scalars) Imports() []*Import {
	imports := []*Import{}
	for _, t := range ts {
		if v := t.Import; v != nil {
			imports = append(imports, v)
		}
	}

	return imports
}

func main() {
	types := Scalars{
		{Type: "string"},
		{Type: "int"},
		{Type: "int8"},
		{Type: "int16"},
		{Type: "int32"},
		{Type: "int64"},
		{Type: "uint"},
		{Type: "uint8"},
		{Type: "uint16"},
		{Type: "uint32"},
		{Type: "uint64"},
		{Type: "float32"},
		{Type: "float64"},
		{Type: "Time", Import: &Import{Path: "time"}},
	}

	for filename, tmplName := range map[string]string{
		"to_ptr.go":   "scalar to pointer",
		"from_ptr.go": "scalar from pointer",
	} {
		if err := generateFile(filename, tmplName, types); err != nil {
			log.Fatalf("%s file generation failed, %v", filename, err)
		}
	}
}

func generateFile(filename string, tmplName string, types Scalars) error {
	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create %s file, %v", filename, err)
	}
	defer f.Close()

	if err := ptrTmpl.ExecuteTemplate(f, tmplName, types); err != nil {
		return fmt.Errorf("failed to generate %s file, %v", filename, err)
	}

	return nil
}

var ptrTmpl = template.Must(template.New("ptrTmpl").Parse(`
{{- define "header" }}
	// Code generated by smithy-go/ptr/generate.go DO NOT EDIT.
	package ptr

	import (
		{{- range $_, $import := $.Imports }}
			"{{ $import.Path }}"
		{{- end }}
	)
{{- end }}

{{- define "scalar from pointer" }}
	{{ template "header" $ }}

	{{ range $_, $type := $ }}
		{{ template "from pointer func" $type }}
		{{ template "from pointers func" $type }}
	{{- end }}
{{- end }}

{{- define "scalar to pointer" }}
	{{ template "header" $ }}

	{{ range $_, $type := $ }}
		{{ template "to pointer func" $type }}
		{{ template "to pointers func" $type }}
	{{- end }}
{{- end }}

{{- define "to pointer func" }}
	// {{ $.Name }} returns a pointer value for the {{ $.Symbol }} value passed in.
	func {{ $.Name }}(v {{ $.Symbol }}) *{{ $.Symbol }} {
		return &v
	}
{{- end }}

{{- define "to pointers func" }}
	// {{ $.Name }}Slice returns a slice of {{ $.Symbol }} pointers from the values
	// passed in.
	func {{ $.Name }}Slice(vs []{{ $.Symbol }}) []*{{ $.Symbol }} {
		ps := make([]*{{ $.Symbol }}, len(vs))
		for i, v := range vs {
			ps[i] = &v
		}

		return ps
	}

	// {{ $.Name }}Map returns a map of {{ $.Symbol }} pointers from the values
	// passed in.
	func {{ $.Name }}Map(vs map[string]{{ $.Symbol }}) map[string]*{{ $.Symbol }} {
		ps := make(map[string]*{{ $.Symbol }}, len(vs))
		for k, v := range vs {
			ps[k] = &v
		}

		return ps
	}
{{- end }}

{{- define "from pointer func" }}
	// To{{ $.Name }} returns {{ $.Symbol }} value dereferenced if the passed
	// in pointer was not nil. Returns a {{ $.Symbol }} zero value if the
	// pointer was nil.
	func To{{ $.Name }}(p *{{ $.Symbol }}) (v {{ $.Symbol }}) {
		if p == nil {
			return v
		}
			
		return *p
	}
{{- end }}

{{- define "from pointers func" }}
	// To{{ $.Name }}Slice returns a slice of {{ $.Symbol }} values, that are
	// dereferenced if the passed in pointer was not nil. Returns a {{ $.Symbol }}
	// zero value if the pointer was nil.
	func To{{ $.Name }}Slice(vs []*{{ $.Symbol }}) []{{ $.Symbol }} {
		ps := make([]{{ $.Symbol }}, len(vs))
		for i, v := range vs {
			ps[i] = To{{ $.Name }}(v)
		}

		return ps
	}

	// To{{ $.Name }}Map returns a map of {{ $.Symbol }} values, that are
	// dereferenced if the passed in pointer was not nil. The {{ $.Symbol }}
	// zero value is used if the pointer was nil.
	func To{{ $.Name }}Map(vs map[string]*{{ $.Symbol }}) map[string]{{ $.Symbol }} {
		ps := make(map[string]{{ $.Symbol }}, len(vs))
		for k, v := range vs {
			ps[k] = To{{ $.Name }}(v)
		}

		return ps
	}
{{- end }}
`))
