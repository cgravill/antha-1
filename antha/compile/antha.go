// antha.go: Part of the Antha language
// Copyright (C) 2017 The Antha authors. All rights reserved.
//
// This program is free software; you can redistribute it and/or
// modify it under the terms of the GNU General Public License
// as published by the Free Software Foundation; either version 2
// of the License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program; if not, write to the Free Software
// Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA  02110-1301, USA.
//
// For more information relating to the software or licensing issues please
// contact license@antha-lang.org or write to the Antha team c/o
// Synthace Ltd. The London Bioscience Innovation Centre
// 2 Royal College St, London NW1 0NH UK

package compile

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"text/template"

	"github.com/antha-lang/antha/antha/ast"
	"github.com/antha-lang/antha/antha/parser"
	"github.com/antha-lang/antha/antha/token"
)

const (
	elementFilename = "element.go"
)

const (
	tabWidth    = 8
	printerMode = UseSpaces | TabIndent
)

var (
	errNotAnthaFile = errors.New("not antha file")
)

type posError struct {
	message string
	pos     token.Pos
}

func (e posError) Error() string {
	return e.message
}

func (e posError) Pos() token.Pos {
	return e.pos
}

func throwErrorf(pos token.Pos, format string, args ...interface{}) {
	panic(posError{
		message: fmt.Sprintf(format, args...),
		pos:     pos,
	})
}

// A Field is a field of a message
type Field struct {
	Name string
	Type ast.Expr // Fully qualified go type name
	Doc  string
	Tag  string
}

// A Message is an input or an output or user defined type
type Message struct {
	Name   string
	Doc    string
	Fields []*Field
	Kind   token.Token // One of token.{DATA, PARAMETERS, OUTPUTS, INPUTS}
}

func isOutput(tok token.Token) bool {
	switch tok {
	case token.OUTPUTS, token.DATA:
		return true
	default:
		return false
	}
}

func isInput(tok token.Token) bool {
	switch tok {
	case token.INPUTS, token.PARAMETERS:
		return true
	default:
		return false
	}
}

func isAnthaGenDeclToken(tok token.Token) bool {
	switch tok {
	case token.OUTPUTS, token.DATA, token.PARAMETERS, token.INPUTS:
		return true
	default:
		return false
	}
}

// An ImportReq is a request to add an import
type ImportReq struct {
	Path    string // Package path
	Name    string // Package alias
	useExpr string // Dummy expression to supress unused imports
}

func (r *ImportReq) ImportName() string {
	if len(r.Name) != 0 {
		return r.Name
	}

	return path.Base(r.Path)
}

type ImportReqs []*ImportReq

func (irs ImportReqs) Sort() {
	sort.Slice(irs, func(i, j int) bool {
		iSplit := strings.Split(irs[i].Path, "/")
		jSplit := strings.Split(irs[j].Path, "/")
		iLen, jLen := len(iSplit), len(jSplit)
		l := iLen
		if jLen < iLen {
			l = jLen
		}
		for idx := 0; idx < l; idx++ {
			if iSplit[idx] < jSplit[idx] {
				return true
			} else if iSplit[idx] > jSplit[idx] {
				return false
			}
		}
		return iLen < jLen
	})
}

// Antha is a preprocessing pass from antha file to go file
type Antha struct {
	fileSet *token.FileSet
	file    *ast.File
	// Description of this element
	description string
	// Path to element file
	elementPath string
	// messages of an element as well as inputs and outputs
	messages []*Message
	// Protocol name as given in Antha file
	protocolName string

	// inputs or outputs of an element but not messages
	TokenByParamName map[string]token.Token
	// Imports in protocol and imports to add
	ImportReqs   ImportReqs
	importByName map[string]struct{}
}

var (
	// Replacements for identifiers in expressions in functions
	intrinsics = map[string]string{
		"Centrifuge":    "execute.Centrifuge",
		"Electroshock":  "execute.Electroshock",
		"ExecuteMixes":  "execute.ExecuteMixes",
		"Incubate":      "execute.Incubate",
		"Mix":           "execute.Mix",
		"MixInto":       "execute.MixInto",
		"MixNamed":      "execute.MixNamed",
		"MixTo":         "execute.MixTo",
		"MixerPrompt":   "execute.MixerPrompt",
		"NewComponent":  "execute.NewComponent",
		"NewPlate":      "execute.NewPlate",
		"Prompt":        "execute.Prompt",
		"ReadEM":        "execute.ReadEM",
		"Sample":        "execute.Sample",
		"SetInputPlate": "execute.SetInputPlate",
		"SplitSample":   "execute.SplitSample",
	}

	// Replacement for go type names in type expressions and types and type lists
	types = map[string]string{
		"Amount":               "wunit.Amount",
		"Angle":                "wunit.Angle",
		"AngularVelocity":      "wunit.AngularVelocity",
		"Area":                 "wunit.Area",
		"Capacitance":          "wunit.Capacitance",
		"Concentration":        "wunit.Concentration",
		"DNASequence":          "wtype.DNASequence",
		"Density":              "wunit.Density",
		"DeviceMetadata":       "api.DeviceMetadata",
		"Energy":               "wunit.Energy",
		"File":                 "wtype.File",
		"FlowRate":             "wunit.FlowRate",
		"Force":                "wunit.Force",
		"JobID":                "jobfile.JobID",
		"IncubateOpt":          "execute.IncubateOpt",
		"LHComponent":          "wtype.Liquid",
		"LHPlate":              "wtype.LHPlate",
		"LHTip":                "wtype.LHTip",
		"LHTipbox":             "wtype.LHTipbox",
		"LHWell":               "wtype.LHWell",
		"Length":               "wunit.Length",
		"Liquid":               "wtype.Liquid",
		"LiquidType":           "wtype.LiquidType",
		"Mass":                 "wunit.Mass",
		"Moles":                "wunit.Moles",
		"PolicyName":           "wtype.PolicyName",
		"Plate":                "wtype.Plate",
		"Pressure":             "wunit.Pressure",
		"Rate":                 "wunit.Rate",
		"Resistance":           "wunit.Resistance",
		"SpecificHeatCapacity": "wunit.SpecificHeatCapacity",
		"SubstanceQuantity":    "wunit.SubstanceQuantity",
		"Temperature":          "wunit.Temperature",
		"Time":                 "wunit.Time",
		"Velocity":             "wunit.Velocity",
		"Voltage":              "wunit.Voltage",
		"Volume":               "wunit.Volume",
		"Warning":              "wtype.Warning",
	}
)

// NewAntha creates a new antha pass
func NewAntha(fileSet *token.FileSet, src *ast.File) (*Antha, error) {
	if src.Tok != token.PROTOCOL {
		return nil, errNotAnthaFile
	}

	p := &Antha{
		fileSet:      fileSet,
		file:         src,
		protocolName: src.Name.Name,
		elementPath:  fileSet.File(src.Package).Name(),
		description:  src.Doc.Text(),
		importByName: make(map[string]struct{}),
	}

	// TODO: add usage tracking to replace useExpr
	p.addImportReq(&ImportReq{
		Path: "github.com/antha-lang/antha/laboratory",
	})
	p.addImportReq(&ImportReq{
		Path: "github.com/antha-lang/antha/workflow",
	})
	p.addImportReq(&ImportReq{
		Path:    "github.com/antha-lang/antha/antha/anthalib/wtype",
		useExpr: "wtype.FALSE",
	})
	p.addImportReq(&ImportReq{
		Path:    "github.com/antha-lang/antha/antha/AnthaStandardLibrary/Packages/jobfile",
		useExpr: "jobfile.DefaultClient",
	})
	p.addImportReq(&ImportReq{
		Path:    "github.com/antha-lang/antha/antha/anthalib/wunit",
		useExpr: "wunit.GetGlobalUnitRegistry",
	})
	p.addImportReq(&ImportReq{
		Path:    "github.com/antha-lang/antha/execute",
		useExpr: "execute.MixInto",
	})

	p.file.Tok = token.PACKAGE
	// Case-insensitive comparison because some filesystems are
	// case-insensitive
	packagePath := filepath.Base(filepath.Dir(p.elementPath))
	if e, f := p.protocolName, packagePath; strings.ToLower(e) != strings.ToLower(f) {
		return nil, fmt.Errorf("%s: expecting protocol %s to be in directory %s", p.elementPath, e, f)
	}

	p.recordImports()
	p.recordMessages()
	if err := p.validateMessages(); err != nil {
		return nil, err
	}

	return p, nil
}

// getImportInsertPos returns position of last import decl or last decl if no
// import decl is present.
func getImportInsertPos(decls []ast.Decl) token.Pos {
	var lastNode ast.Node
	for _, d := range decls {
		gd, ok := d.(*ast.GenDecl)
		if !ok || gd.Tok != token.IMPORT {
			if lastNode == nil {
				lastNode = d
			}
			continue
		}
		lastNode = gd
	}

	if lastNode == nil {
		return token.NoPos
	}
	return lastNode.Pos()
}

// setImports merges multiple import blocks and then adds paths
func (p *Antha) setImports() {
	var nonImports []ast.Decl
	insertPos := getImportInsertPos(p.file.Decls)

	for _, d := range p.file.Decls {
		gd, ok := d.(*ast.GenDecl)
		if !ok || gd.Tok != token.IMPORT {
			nonImports = append(nonImports, d)
		}
	}

	p.ImportReqs.Sort()
	imports := make([]ast.Spec, 0, len(p.ImportReqs))
	for _, req := range p.ImportReqs {
		imp := &ast.ImportSpec{
			Path: &ast.BasicLit{
				Kind:     token.STRING,
				Value:    strconv.Quote(req.Path),
				ValuePos: insertPos,
			},
		}
		if len(req.Name) != 0 {
			imp.Name = ast.NewIdent(req.Name)
		}
		imports = append(imports, imp)
	}

	merged := &ast.GenDecl{
		Tok:    token.IMPORT,
		Lparen: insertPos,
		Rparen: insertPos,
		Specs:  imports,
	}

	p.file.Decls = append([]ast.Decl{merged}, nonImports...)
}

func (p *Antha) addImportReq(req *ImportReq) {
	name := req.ImportName()
	if _, found := p.importByName[name]; !found {
		p.ImportReqs = append(p.ImportReqs, req)
		p.importByName[name] = struct{}{}
	}
}

func (p *Antha) recordImports() {
	for _, decl := range p.file.Decls {
		gd, ok := decl.(*ast.GenDecl)
		if ok && gd.Tok == token.IMPORT {
			for _, spec := range gd.Specs {
				im := spec.(*ast.ImportSpec)

				value, _ := strconv.Unquote(im.Path.Value)
				req := &ImportReq{
					Path: value,
				}
				if im.Name != nil {
					req.Name = im.Name.String()
				}

				p.addImportReq(req)
			}
		}
	}
}

// recordMessages records all the spec definitions for inputs and outputs to element
func (p *Antha) recordMessages() {
	join := func(xs ...string) string {
		var ret []string
		for _, x := range xs {
			if len(x) == 0 {
				continue
			}
			ret = append(ret, x)
		}
		return strings.Join(ret, "\n")
	}

	for _, decl := range p.file.Decls {
		decl, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}

		if !isAnthaGenDeclToken(decl.Tok) {
			continue
		}

		for _, spec := range decl.Specs {
			spec, ok := spec.(*ast.TypeSpec)
			if !ok {
				throwErrorf(spec.Pos(), "expecting type")
			}
			typ, ok := spec.Type.(*ast.StructType)
			if !ok {
				throwErrorf(spec.Pos(), "expecting struct type")
			}

			var fields []*Field
			for _, field := range typ.Fields.List {
				for _, name := range field.Names {
					f := &Field{
						Name: name.String(),
						Type: p.desugarTypeExpr(field.Type),
						Doc:  join(field.Comment.Text(), field.Doc.Text()),
					}
					if field.Tag != nil {
						f.Tag = field.Tag.Value
					}
					fields = append(fields, f)
				}
			}

			p.messages = append(p.messages, &Message{
				Name:   spec.Name.String(),
				Fields: fields,
				Doc:    join(decl.Doc.Text(), spec.Comment.Text(), spec.Doc.Text()),
				Kind:   decl.Tok,
			})
		}
	}
}

func (p *Antha) validateMessages() error {
	p.TokenByParamName = make(map[string]token.Token)

	seenMessage := make(map[token.Token]*Message)

	for _, msg := range p.messages {
		name := msg.Name
		if _, found := seenMessage[msg.Kind]; found {
			return fmt.Errorf("%s already declared", name)
		}

		seenMessage[msg.Kind] = msg

		for _, field := range msg.Fields {
			if tok, found := p.TokenByParamName[field.Name]; found {
				return fmt.Errorf("%s already declared as %v", name, tok)
			}
			p.TokenByParamName[field.Name] = msg.Kind

			if !ast.IsExported(field.Name) {
				return fmt.Errorf("field %s must begin with an upper case letter", name)
			}
		}
	}

	// add in empty ones if there are any missing
	for _, tok := range []token.Token{token.OUTPUTS, token.DATA, token.PARAMETERS, token.INPUTS} {
		if _, found := seenMessage[tok]; !found {
			p.messages = append(p.messages, &Message{
				Name: tok.String(),
				Kind: tok,
			})
			p.file.Decls = append(p.file.Decls, &ast.GenDecl{
				Tok: tok,
				Specs: []ast.Spec{
					&ast.TypeSpec{
						Name: ast.NewIdent(tok.String()),
						Type: &ast.StructType{
							Fields: &ast.FieldList{},
						},
					},
				},
			})
		}
	}

	return nil
}

// Transform rewrites AST to go standard primitives
func (p *Antha) Transform(files *AnthaFiles) error {
	p.desugar()
	p.setImports()
	p.addUses()

	buf := new(bytes.Buffer)
	compiler := &Config{
		Mode:     printerMode,
		Tabwidth: tabWidth,
	}
	lineMap, err := compiler.Fprint(buf, p.fileSet, p.file)
	if err != nil {
		return err
	}

	if err := p.printFunctions(buf, lineMap); err != nil {
		return err
	}

	elementName := path.Join(p.protocolName, elementFilename)

	files.addFile(elementName, buf)
	return nil
}

func (p *Antha) addUses() {
	for _, req := range p.ImportReqs {
		if len(req.useExpr) == 0 {
			continue
		}
		decl := &ast.GenDecl{
			Tok: token.VAR,
		}

		decl.Specs = append(decl.Specs, &ast.ValueSpec{
			Names:  identList("_"),
			Values: []ast.Expr{mustParseExpr(req.useExpr)},
		})
		p.file.Decls = append(p.file.Decls, decl)
	}
}

// printFunctions generates synthetic antha functions and data stuctures
func (p *Antha) printFunctions(out io.Writer, lineMap map[int]int) error {
	var tmpl = `
type {{.ElementTypeName}} struct {
	name workflow.ElementInstanceName

	Inputs     Inputs
	Outputs    Outputs
	Parameters Parameters
	Data       Data
}

func New(installer laboratory.ElementInstaller, name workflow.ElementInstanceName) *{{.ElementTypeName}} {
	element := &{{.ElementTypeName}}{name: name}
	installer.InstallElement(element)
	return element
}

func (element *{{.ElementTypeName}}) Name() workflow.ElementInstanceName {
	return element.name
}

func (element *{{.ElementTypeName}}) TypeName() workflow.ElementTypeName {
	return {{printf "%q" .ElementTypeName}}
}

func RegisterLineMap(labBuild *laboratory.LaboratoryBuilder) {
	lineMap := map[int]int{
		{{range $key, $value := .LineMap}}{{$key}}: {{$value}}, {{end}}
	}
	labBuild.RegisterLineMap(
		{{printf "%q" .GeneratedPath}},
		{{printf "%q" .Path}},
		{{printf "%q" .ElementTypeName}},
		lineMap)
}
`
	type TVars struct {
		ElementTypeName string
		GeneratedPath   string
		Path            string
		LineMap         map[int]int
	}

	tv := TVars{
		ElementTypeName: p.protocolName,
		GeneratedPath:   filepath.Join(filepath.Dir(p.elementPath), elementFilename),
		Path:            p.elementPath,
		LineMap:         lineMap,
	}

	return template.Must(template.New("").Parse(tmpl)).Execute(out, tv)
}

// desugar updates AST for antha semantics
func (p *Antha) desugar() {
	for idx, d := range p.file.Decls {
		switch d := d.(type) {

		case *ast.GenDecl:
			ast.Inspect(d, p.inspectTypes)
			p.desugarGenDecl(d)

		case *ast.AnthaDecl:
			ast.Inspect(d.Body, p.inspectIntrinsics)
			ast.Inspect(d.Body, p.inspectParamUses)
			ast.Inspect(d.Body, p.inspectTypes)
			p.file.Decls[idx] = p.desugarAnthaDecl(d)

		default:
			ast.Inspect(d, p.inspectTypes)
		}
	}
}

func identList(name string) []*ast.Ident {
	return []*ast.Ident{ast.NewIdent(name)}
}

func mustParseExpr(x string) ast.Expr {
	r, err := parser.ParseExpr(x)
	if err != nil {
		panic(fmt.Errorf("cannot parse %s: %s", x, err))
	}
	return r
}

// desugarGenDecl returns standard go ast for antha GenDecls
func (p *Antha) desugarGenDecl(d *ast.GenDecl) {
	if !isAnthaGenDeclToken(d.Tok) {
		return
	}

	d.Tok = token.TYPE
}

// desugarAnthaDecl returns standard go ast for antha decl.
//
// E.g.,
//   Validation
// to
//  func (element *ElementTypeName) Validation(lab *laboratory.Laboratory) error
func (p *Antha) desugarAnthaDecl(d *ast.AnthaDecl) ast.Decl {
	body := d.Body
	body.List = append(body.List,
		&ast.ReturnStmt{
			Return:  d.Pos(),
			Results: []ast.Expr{mustParseExpr("nil")},
		})
	f := &ast.FuncDecl{
		Doc: d.Doc,
		Recv: &ast.FieldList{
			Opening: d.Pos(),
			List: []*ast.Field{
				{
					Names: identList("element"),
					Type:  mustParseExpr("*" + p.protocolName),
				},
			},
		},
		Name: ast.NewIdent(d.Tok.String()),
		Body: body,
	}

	f.Type = &ast.FuncType{
		Func: d.Pos(),
		Params: &ast.FieldList{
			Opening: d.Pos(),
			List: []*ast.Field{
				{
					Names: identList("lab"),
					Type:  mustParseExpr("*laboratory.Laboratory"),
				},
			},
		},
		Results: &ast.FieldList{
			Opening: d.Pos(),
			List: []*ast.Field{
				{
					Type: mustParseExpr("error"),
				},
			},
		},
	}

	return f
}

// desugarTypeIdent returns appropriate nested SelectorExpr for the replacement for
// Identifier
func (p *Antha) desugarTypeIdent(t *ast.Ident) ast.Expr {
	v, ok := types[t.Name]
	if !ok {
		return t
	}

	return mustParseExpr(v)
}

// desugarTypeExpr returns appropriate go type for an antha (type) expr
func (p *Antha) desugarTypeExpr(t ast.Node) ast.Expr {
	switch t := t.(type) {
	case nil:
		return nil

	case *ast.Ident:
		return p.desugarTypeIdent(t)

	case *ast.ParenExpr:
		t.X = p.desugarTypeExpr(t.X)

	case *ast.SelectorExpr:

	case *ast.StarExpr:
		t.X = p.desugarTypeExpr(t.X)

	case *ast.ArrayType:
		t.Elt = p.desugarTypeExpr(t.Elt)

	case *ast.StructType:
		ast.Inspect(t, p.inspectTypes)

	case *ast.FuncType:
		ast.Inspect(t, p.inspectTypes)

	case *ast.InterfaceType:
		ast.Inspect(t, p.inspectTypes)

	case *ast.MapType:
		t.Key = p.desugarTypeExpr(t.Key)
		t.Value = p.desugarTypeExpr(t.Value)

	case *ast.ChanType:
		t.Value = p.desugarTypeExpr(t.Value)

	case *ast.Ellipsis:

	default:
		throwErrorf(t.Pos(), "unexpected expression %s of type %T", t, t)
	}

	return t.(ast.Expr)
}

func inspectExprList(exprs []ast.Expr, w func(ast.Node) bool) {
	for _, expr := range exprs {
		ast.Inspect(expr, w)
	}
}

// inspectTypes replaces bare antha types with go qualified names.
//
// Changing all idents blindly would be simpler but opt instead with only
// replacing idents that appear in types.
func (p *Antha) inspectTypes(n ast.Node) bool {
	switch n := n.(type) {
	case nil:

	case *ast.Field:
		n.Type = p.desugarTypeExpr(n.Type)

	case *ast.TypeSpec:
		n.Type = p.desugarTypeExpr(n.Type)

	case *ast.MapType:
		n.Key = p.desugarTypeExpr(n.Key)
		n.Value = p.desugarTypeExpr(n.Value)

	case *ast.ArrayType:
		n.Elt = p.desugarTypeExpr(n.Elt)

	case *ast.ChanType:
		n.Value = p.desugarTypeExpr(n.Value)

	case *ast.FuncLit:
		n.Type = p.desugarTypeExpr(n.Type).(*ast.FuncType)
		ast.Inspect(n.Body, p.inspectTypes)

	case *ast.CompositeLit:
		n.Type = p.desugarTypeExpr(n.Type)
		inspectExprList(n.Elts, p.inspectTypes)

	case *ast.TypeAssertExpr:
		n.Type = p.desugarTypeExpr(n.Type)

	case *ast.ValueSpec:
		n.Type = p.desugarTypeExpr(n.Type)
		inspectExprList(n.Values, p.inspectTypes)

	default:
		return true
	}

	return false
}

// inspectParamUses replaces bare antha identifiers with go qualified names
func (p *Antha) inspectParamUses(node ast.Node) bool {
	// desugar if it is a known param
	rewriteIdent := func(node *ast.Ident) {
		tok, ok := p.TokenByParamName[node.Name]
		if !ok {
			return
		}

		node.Name = "element." + tok.String() + "." + node.Name
	}

	rewriteAssignLHS := func(node *ast.AssignStmt) {
		for _, lhs := range node.Lhs {
			ident, ok := lhs.(*ast.Ident)
			if !ok {
				continue
			}

			tok, found := p.TokenByParamName[ident.Name]
			if !found || !isOutput(tok) {
				continue
			}

			ident.Name = "element." + tok.String() + "." + ident.Name
		}
	}

	switch n := node.(type) {

	case nil:
		return false

	case *ast.AssignStmt:
		rewriteAssignLHS(n)

	case *ast.KeyValueExpr:
		if _, identKey := n.Key.(*ast.Ident); identKey {
			// Skip identifiers that are keys
			ast.Inspect(n.Value, p.inspectParamUses)
			return false
		}
	case *ast.Ident:
		rewriteIdent(n)

	case *ast.SelectorExpr:
		// Skip identifiers that are field accesses
		ast.Inspect(n.X, p.inspectParamUses)
		return false
	}
	return true
}

// inspectIntrinsics replaces bare antha function names with go qualified
// names
func (p *Antha) inspectIntrinsics(node ast.Node) bool {
	switch n := node.(type) {
	case *ast.CallExpr:
		ident, direct := n.Fun.(*ast.Ident)
		if !direct {
			break
		}

		if desugar, ok := intrinsics[ident.Name]; ok {
			ident.Name = desugar
			n.Args = append([]ast.Expr{ast.NewIdent("lab")}, n.Args...) // only for now.
		}
	}
	return true
}
