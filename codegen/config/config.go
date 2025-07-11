package config

import (
	"bytes"
	"errors"
	"fmt"
	"go/types"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/vektah/gqlparser/v2"
	"github.com/vektah/gqlparser/v2/ast"
	"golang.org/x/tools/go/packages"
	"gopkg.in/yaml.v3"

	"github.com/99designs/gqlgen/codegen/templates"
	"github.com/99designs/gqlgen/internal/code"
)

type Config struct {
	SchemaFilename                       StringList                 `yaml:"schema,omitempty"`
	Exec                                 ExecConfig                 `yaml:"exec"`
	Model                                PackageConfig              `yaml:"model,omitempty"`
	Federation                           PackageConfig              `yaml:"federation,omitempty"`
	Resolver                             ResolverConfig             `yaml:"resolver,omitempty"`
	AutoBind                             []string                   `yaml:"autobind"`
	Models                               TypeMap                    `yaml:"models,omitempty"`
	StructTag                            string                     `yaml:"struct_tag,omitempty"`
	Directives                           map[string]DirectiveConfig `yaml:"directives,omitempty"`
	LocalPrefix                          string                     `yaml:"local_prefix,omitempty"`
	GoBuildTags                          StringList                 `yaml:"go_build_tags,omitempty"`
	GoInitialisms                        GoInitialismsConfig        `yaml:"go_initialisms,omitempty"`
	OmitSliceElementPointers             bool                       `yaml:"omit_slice_element_pointers,omitempty"`
	OmitGetters                          bool                       `yaml:"omit_getters,omitempty"`
	OmitInterfaceChecks                  bool                       `yaml:"omit_interface_checks,omitempty"`
	OmitComplexity                       bool                       `yaml:"omit_complexity,omitempty"`
	OmitGQLGenFileNotice                 bool                       `yaml:"omit_gqlgen_file_notice,omitempty"`
	OmitGQLGenVersionInFileNotice        bool                       `yaml:"omit_gqlgen_version_in_file_notice,omitempty"`
	OmitRootModels                       bool                       `yaml:"omit_root_models,omitempty"`
	OmitResolverFields                   bool                       `yaml:"omit_resolver_fields,omitempty"`
	OmitPanicHandler                     bool                       `yaml:"omit_panic_handler,omitempty"`
	UseFunctionSyntaxForExecutionContext bool                       `yaml:"use_function_syntax_for_execution_context,omitempty"`
	// If this is set to true, argument directives that
	// decorate a field with a null value will still be called.
	//
	// This enables argument directives to not just mutate
	// argument values but to set them even if they're null.
	CallArgumentDirectivesWithNull bool           `yaml:"call_argument_directives_with_null,omitempty"`
	StructFieldsAlwaysPointers     bool           `yaml:"struct_fields_always_pointers,omitempty"`
	ReturnPointersInUnmarshalInput bool           `yaml:"return_pointers_in_unmarshalinput,omitempty"`
	ResolversAlwaysReturnPointers  bool           `yaml:"resolvers_always_return_pointers,omitempty"`
	NullableInputOmittable         bool           `yaml:"nullable_input_omittable,omitempty"`
	EnableModelJsonOmitemptyTag    *bool          `yaml:"enable_model_json_omitempty_tag,omitempty"`
	EnableModelJsonOmitzeroTag     *bool          `yaml:"enable_model_json_omitzero_tag,omitempty"`
	SkipValidation                 bool           `yaml:"skip_validation,omitempty"`
	SkipModTidy                    bool           `yaml:"skip_mod_tidy,omitempty"`
	Sources                        []*ast.Source  `yaml:"-"`
	Packages                       *code.Packages `yaml:"-"`
	Schema                         *ast.Schema    `yaml:"-"`

	// Deprecated: use Federation instead. Will be removed next release
	Federated bool `yaml:"federated,omitempty"`
}

var cfgFilenames = []string{".gqlgen.yml", "gqlgen.yml", "gqlgen.yaml"}

// templatePackageNames is a list of packages names that the default templates use, in order to preload those for performance considerations
// any additional package added to the base templates should be added here to improve performance and load all packages in bulk
var templatePackageNames = []string{
	"context", "fmt", "io", "strconv", "time", "sync", "strings", "sync/atomic", "embed", "golang.org/x/sync/semaphore",
	"errors", "bytes", "github.com/vektah/gqlparser/v2", "github.com/vektah/gqlparser/v2/ast",
	"github.com/99designs/gqlgen/graphql", "github.com/99designs/gqlgen/graphql/introspection",
}

// DefaultConfig creates a copy of the default config
func DefaultConfig() *Config {
	falseValue := false
	return &Config{
		SchemaFilename:                 StringList{"schema.graphql"},
		Model:                          PackageConfig{Filename: "models_gen.go"},
		Exec:                           ExecConfig{Filename: "generated.go"},
		Directives:                     map[string]DirectiveConfig{},
		Models:                         TypeMap{},
		StructFieldsAlwaysPointers:     true,
		ReturnPointersInUnmarshalInput: false,
		ResolversAlwaysReturnPointers:  true,
		NullableInputOmittable:         false,
		EnableModelJsonOmitzeroTag:     &falseValue,
	}
}

// LoadDefaultConfig loads the default config so that it is ready to be used
func LoadDefaultConfig() (*Config, error) {
	config := DefaultConfig()

	for _, filename := range config.SchemaFilename {
		filename = filepath.ToSlash(filename)
		var err error
		var schemaRaw []byte
		schemaRaw, err = os.ReadFile(filename)
		if err != nil {
			return nil, fmt.Errorf("unable to open schema: %w", err)
		}

		config.Sources = append(config.Sources, &ast.Source{Name: filename, Input: string(schemaRaw)})
	}

	return config, nil
}

// LoadConfigFromDefaultLocations looks for a config file in the current directory, and all parent directories
// walking up the tree. The closest config file will be returned.
func LoadConfigFromDefaultLocations() (*Config, error) {
	cfgFile, err := findCfg()
	if err != nil {
		return nil, err
	}

	err = os.Chdir(filepath.Dir(cfgFile))
	if err != nil {
		return nil, fmt.Errorf("unable to enter config dir: %w", err)
	}
	return LoadConfig(cfgFile)
}

var path2regex = strings.NewReplacer(
	`.`, `\.`,
	`*`, `.+`,
	`\`, `[\\/]`,
	`/`, `[\\/]`,
)

// LoadConfig reads the gqlgen.yml config file
func LoadConfig(filename string) (*Config, error) {
	b, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("unable to read config: %w", err)
	}

	return ReadConfig(bytes.NewReader(b))
}

func ReadConfig(cfgFile io.Reader) (*Config, error) {
	config := DefaultConfig()

	dec := yaml.NewDecoder(cfgFile)
	dec.KnownFields(true)

	if err := dec.Decode(config); err != nil {
		return nil, fmt.Errorf("unable to parse config: %w", err)
	}

	if err := CompleteConfig(config); err != nil {
		return nil, err
	}

	return config, nil
}

// CompleteConfig fills in the schema and other values to a config loaded from
// YAML.
func CompleteConfig(config *Config) error {
	defaultDirectives := map[string]DirectiveConfig{
		"skip":        {SkipRuntime: true},
		"include":     {SkipRuntime: true},
		"deprecated":  {SkipRuntime: true},
		"specifiedBy": {SkipRuntime: true},
		"oneOf":       {SkipRuntime: true},
	}

	for key, value := range defaultDirectives {
		if _, defined := config.Directives[key]; !defined {
			config.Directives[key] = value
		}
	}

	preGlobbing := config.SchemaFilename
	config.SchemaFilename = StringList{}
	for _, f := range preGlobbing {
		var matches []string

		// for ** we want to override default globbing patterns and walk all
		// subdirectories to match schema files.
		if strings.Contains(f, "**") {
			pathParts := strings.SplitN(f, "**", 2)
			rest := strings.TrimPrefix(strings.TrimPrefix(pathParts[1], `\`), `/`)
			// turn the rest of the glob into a regex, anchored only at the end because ** allows
			// for any number of dirs in between and walk will let us match against the full path name
			globRe := regexp.MustCompile(path2regex.Replace(rest) + `$`)

			if err := filepath.Walk(pathParts[0], func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}

				if globRe.MatchString(strings.TrimPrefix(path, pathParts[0])) {
					matches = append(matches, path)
				}

				return nil
			}); err != nil {
				return fmt.Errorf("failed to walk schema at root %s: %w", pathParts[0], err)
			}
		} else {
			var err error
			matches, err = filepath.Glob(f)
			if err != nil {
				return fmt.Errorf("failed to glob schema filename %s: %w", f, err)
			}
		}

		for _, m := range matches {
			if config.SchemaFilename.Has(m) {
				continue
			}
			config.SchemaFilename = append(config.SchemaFilename, m)
		}
	}

	for _, filename := range config.SchemaFilename {
		filename = filepath.ToSlash(filename)
		var err error
		var schemaRaw []byte
		schemaRaw, err = os.ReadFile(filename)
		if err != nil {
			return fmt.Errorf("unable to open schema: %w", err)
		}

		config.Sources = append(config.Sources, &ast.Source{Name: filename, Input: string(schemaRaw)})
	}

	config.GoInitialisms.setInitialisms()

	return nil
}

func (c *Config) Init() error {
	if c.Packages == nil {
		c.Packages = code.NewPackages(
			code.WithBuildTags(c.GoBuildTags...),
			code.PackagePrefixToCache("github.com/99designs/gqlgen/graphql"),
			code.WithPreloadNames(templatePackageNames...),
		)
	}

	if c.Schema == nil {
		if err := c.LoadSchema(); err != nil {
			return err
		}
	}

	err := c.injectTypesFromSchema()
	if err != nil {
		return err
	}

	// prefetch all packages in one big packages.Load call
	c.Packages.LoadAll(c.packageList()...)

	err = c.autobind()
	if err != nil {
		return err
	}

	c.injectBuiltins()

	//  check everything is valid on the way out
	err = c.check()
	if err != nil {
		return err
	}

	return nil
}

func (c *Config) packageList() []string {
	pkgs := []string{
		"github.com/99designs/gqlgen/graphql",
		"github.com/99designs/gqlgen/graphql/introspection",
	}
	pkgs = append(pkgs, c.Models.ReferencedPackages()...)
	pkgs = append(pkgs, c.AutoBind...)
	return pkgs
}

func (c *Config) ReloadAllPackages() {
	c.Packages.ReloadAll(c.packageList()...)
}

func (c *Config) IsRoot(def *ast.Definition) bool {
	return def == c.Schema.Query || def == c.Schema.Mutation || def == c.Schema.Subscription
}

func (c *Config) injectTypesFromSchema() error {
	for _, d := range []string{"goModel", "goExtraField", "goField", "goTag", "goEnum"} {
		c.Directives[d] = DirectiveConfig{SkipRuntime: true}
	}

	for _, schemaType := range c.Schema.Types {
		if c.IsRoot(schemaType) {
			continue
		}

		if bd := schemaType.Directives.ForName("goModel"); bd != nil {
			if ma := bd.Arguments.ForName("model"); ma != nil {
				if mv, err := ma.Value.Value(nil); err == nil {
					c.Models.Add(schemaType.Name, mv.(string))
				}
			}

			if ma := bd.Arguments.ForName("models"); ma != nil {
				if mvs, err := ma.Value.Value(nil); err == nil {
					for _, mv := range mvs.([]any) {
						c.Models.Add(schemaType.Name, mv.(string))
					}
				}
			}

			if fg := bd.Arguments.ForName("forceGenerate"); fg != nil {
				if mv, err := fg.Value.Value(nil); err == nil {
					c.Models.ForceGenerate(schemaType.Name, mv.(bool))
				}
			}
		}

		if schemaType.Kind == ast.Object ||
			schemaType.Kind == ast.InputObject ||
			schemaType.Kind == ast.Interface {
			for _, field := range schemaType.Fields {
				if fd := field.Directives.ForName("goField"); fd != nil {
					// First, copy map entry for type and field to do modifications
					typeMapEntry := c.Models[schemaType.Name]
					typeMapFieldEntry := typeMapEntry.Fields[field.Name]

					if ta := fd.Arguments.ForName("type"); ta != nil {
						if c.Models.UserDefined(schemaType.Name) {
							return fmt.Errorf(
								"argument 'type' for directive @goField (src: %s, line: %d) not applicable for user-defined models",
								fd.Position.Src.Name,
								fd.Position.Line,
							)
						}

						if ft, err := ta.Value.Value(nil); err == nil {
							typeMapFieldEntry.Type = ft.(string)
						}
					}

					if ra := fd.Arguments.ForName("forceResolver"); ra != nil {
						if fr, err := ra.Value.Value(nil); err == nil {
							typeMapFieldEntry.Resolver = fr.(bool)
						}
					}

					if na := fd.Arguments.ForName("name"); na != nil {
						if fr, err := na.Value.Value(nil); err == nil {
							typeMapFieldEntry.FieldName = fr.(string)
						}
					}

					if arg := fd.Arguments.ForName("omittable"); arg != nil {
						if k, err := arg.Value.Value(nil); err == nil {
							val := k.(bool)
							typeMapFieldEntry.Omittable = &val
						}
					}

					// May be uninitialized, so do it now.
					if typeMapEntry.Fields == nil {
						typeMapEntry.Fields = make(map[string]TypeMapField)
					}

					// First, copy back probably modificated field settings
					typeMapEntry.Fields[field.Name] = typeMapFieldEntry

					// And final copy back probably modificated all type map
					c.Models[schemaType.Name] = typeMapEntry
				}
			}

			if efds := schemaType.Directives.ForNames("goExtraField"); len(efds) != 0 {
				for _, efd := range efds {
					if t := efd.Arguments.ForName("type"); t != nil {
						extraField := ModelExtraField{}

						if tv, err := t.Value.Value(nil); err == nil {
							extraField.Type = tv.(string)
						}

						if extraField.Type == "" {
							return fmt.Errorf(
								"argument 'type' for directive @goExtraField (src: %s, line: %d) cannot by empty",
								efd.Position.Src.Name,
								efd.Position.Line,
							)
						}

						if ot := efd.Arguments.ForName("overrideTags"); ot != nil {
							if otv, err := ot.Value.Value(nil); err == nil {
								extraField.OverrideTags = otv.(string)
							}
						}

						if d := efd.Arguments.ForName("description"); d != nil {
							if dv, err := d.Value.Value(nil); err == nil {
								extraField.Description = dv.(string)
							}
						}

						extraFieldName := ""
						if fn := efd.Arguments.ForName("name"); fn != nil {
							if fnv, err := fn.Value.Value(nil); err == nil {
								extraFieldName = fnv.(string)
							}
						}

						// First copy, then modify map entry.
						typeMapEntry := c.Models[schemaType.Name]

						if extraFieldName == "" {
							// Embeddable fields
							typeMapEntry.EmbedExtraFields = append(typeMapEntry.EmbedExtraFields, extraField)
						} else {
							// Regular fields
							if typeMapEntry.ExtraFields == nil {
								typeMapEntry.ExtraFields = make(map[string]ModelExtraField)
							}
							typeMapEntry.ExtraFields[extraFieldName] = extraField
						}

						// Copy back modified map entry
						c.Models[schemaType.Name] = typeMapEntry
					}
				}
			}
		}

		if schemaType.Kind == ast.Enum && !strings.HasPrefix(schemaType.Name, "__") {
			values := make(map[string]EnumValue)

			for _, value := range schemaType.EnumValues {
				if directive := value.Directives.ForName("goEnum"); directive != nil {
					if arg := directive.Arguments.ForName("value"); arg != nil {
						if v, err := arg.Value.Value(nil); err == nil {
							values[value.Name] = EnumValue{
								Value: v.(string),
							}
						}
					}
				}
			}

			if len(values) > 0 {
				model := c.Models[schemaType.Name]
				model.EnumValues = values
				c.Models[schemaType.Name] = model
			}
		}
	}

	return nil
}

type TypeMapEntry struct {
	Model         StringList              `yaml:"model,omitempty"`
	ForceGenerate bool                    `yaml:"forceGenerate,omitempty"`
	Fields        map[string]TypeMapField `yaml:"fields,omitempty"`
	EnumValues    map[string]EnumValue    `yaml:"enum_values,omitempty"`

	// Key is the Go name of the field.
	ExtraFields      map[string]ModelExtraField `yaml:"extraFields,omitempty"`
	EmbedExtraFields []ModelExtraField          `yaml:"embedExtraFields,omitempty"`
}

type TypeMapField struct {
	// Type is the Go type of the field.
	//
	// It supports the builtin basic types (like string or int64), named types
	// (qualified by the full package path), pointers to those types (prefixed
	// with `*`), and slices of those types (prefixed with `[]`).
	//
	// For example, the following are valid types:
	//  string
	//  *github.com/author/package.Type
	//  []string
	//  []*github.com/author/package.Type
	//
	// Note that the type will be referenced from the generated/graphql, which
	// means the package it lives in must not reference the generated/graphql
	// package to avoid circular imports.
	// restrictions.
	Type string `yaml:"type"`

	Resolver        bool   `yaml:"resolver"`
	FieldName       string `yaml:"fieldName"`
	Omittable       *bool  `yaml:"omittable"`
	GeneratedMethod string `yaml:"-"`
}

type EnumValue struct {
	Value string
}

type ModelExtraField struct {
	// Type is the Go type of the field.
	//
	// It supports the builtin basic types (like string or int64), named types
	// (qualified by the full package path), pointers to those types (prefixed
	// with `*`), and slices of those types (prefixed with `[]`).
	//
	// For example, the following are valid types:
	//  string
	//  *github.com/author/package.Type
	//  []string
	//  []*github.com/author/package.Type
	//
	// Note that the type will be referenced from the generated/graphql, which
	// means the package it lives in must not reference the generated/graphql
	// package to avoid circular imports.
	// restrictions.
	Type string `yaml:"type"`

	// OverrideTags is an optional override of the Go field tag.
	OverrideTags string `yaml:"overrideTags"`

	// Description is an optional the Go field doc-comment.
	Description string `yaml:"description"`
}

type StringList []string

func (a *StringList) UnmarshalYAML(unmarshal func(any) error) error {
	var single string
	err := unmarshal(&single)
	if err == nil {
		*a = []string{single}
		return nil
	}

	var multi []string
	err = unmarshal(&multi)
	if err != nil {
		return err
	}

	*a = multi
	return nil
}

func (a StringList) Has(file string) bool {
	for _, existing := range a {
		if existing == file {
			return true
		}
	}
	return false
}

func (c *Config) check() error {
	if c.Models == nil {
		c.Models = TypeMap{}
	}

	type FilenamePackage struct {
		Filename string
		Package  string
		Declaree string
	}

	fileList := map[string][]FilenamePackage{}

	if err := c.Models.Check(); err != nil {
		return fmt.Errorf("config.models: %w", err)
	}
	if err := c.Exec.Check(); err != nil {
		return fmt.Errorf("config.exec: %w", err)
	}
	fileList[c.Exec.ImportPath()] = append(fileList[c.Exec.ImportPath()], FilenamePackage{
		Filename: c.Exec.Filename,
		Package:  c.Exec.Package,
		Declaree: "exec",
	})

	if c.Model.IsDefined() {
		if err := c.Model.Check(); err != nil {
			return fmt.Errorf("config.model: %w", err)
		}
		fileList[c.Model.ImportPath()] = append(fileList[c.Model.ImportPath()], FilenamePackage{
			Filename: c.Model.Filename,
			Package:  c.Model.Package,
			Declaree: "model",
		})
	}
	if c.Resolver.IsDefined() {
		if err := c.Resolver.Check(); err != nil {
			return fmt.Errorf("config.resolver: %w", err)
		}
		fileList[c.Resolver.ImportPath()] = append(fileList[c.Resolver.ImportPath()], FilenamePackage{
			Filename: c.Resolver.Filename,
			Package:  c.Resolver.Package,
			Declaree: "resolver",
		})
	}
	if c.Federation.IsDefined() {
		if err := c.Federation.Check(); err != nil {
			return fmt.Errorf("config.federation: %w", err)
		}
		fileList[c.Federation.ImportPath()] = append(fileList[c.Federation.ImportPath()], FilenamePackage{
			Filename: c.Federation.Filename,
			Package:  c.Federation.Package,
			Declaree: "federation",
		})
		if c.Federation.ImportPath() != c.Exec.ImportPath() {
			return errors.New("federation and exec must be in the same package")
		}
	}
	if c.Federated {
		return errors.New("federated has been removed, instead use\nfederation:\n    filename: path/to/federated.go")
	}

	for importPath, pkg := range fileList {
		for _, file1 := range pkg {
			for _, file2 := range pkg {
				if file1.Package != file2.Package {
					return fmt.Errorf("%s and %s define the same import path (%s) with different package names (%s vs %s)",
						file1.Declaree,
						file2.Declaree,
						importPath,
						file1.Package,
						file2.Package,
					)
				}
			}
		}
	}

	return nil
}

type TypeMap map[string]TypeMapEntry

func (tm TypeMap) Exists(typeName string) bool {
	_, ok := tm[typeName]
	return ok
}

func (tm TypeMap) UserDefined(typeName string) bool {
	m, ok := tm[typeName]
	return ok && len(m.Model) > 0
}

func (tm TypeMap) Check() error {
	for typeName, entry := range tm {
		for _, model := range entry.Model {
			if strings.LastIndex(model, ".") < strings.LastIndex(model, "/") {
				return fmt.Errorf("model %s: invalid type specifier \"%s\" - you need to specify a struct to map to", typeName, entry.Model)
			}
		}

		if len(entry.Model) == 0 {
			for enum, v := range entry.EnumValues {
				if v.Value != "" {
					return fmt.Errorf("model is empty for: %v, but enum value is specified for %v", typeName, enum)
				}
			}
		}
	}
	return nil
}

func (tm TypeMap) ReferencedPackages() []string {
	var pkgs []string

	for _, typ := range tm {
		for _, model := range typ.Model {
			if model == "map[string]any" ||
				model == "map[string]interface{}" ||
				model == "any" ||
				model == "interface{}" {
				continue
			}
			pkg, _ := code.PkgAndType(model)
			if pkg == "" || inStrSlice(pkgs, pkg) {
				continue
			}
			pkgs = append(pkgs, code.QualifyPackagePath(pkg))
		}
	}

	sort.Slice(pkgs, func(i, j int) bool {
		return pkgs[i] > pkgs[j]
	})
	return pkgs
}

func (tm TypeMap) Add(name, goType string) {
	modelCfg := tm[name]
	modelCfg.Model = append(modelCfg.Model, goType)
	tm[name] = modelCfg
}

func (tm TypeMap) ForceGenerate(name string, forceGenerate bool) {
	modelCfg := tm[name]
	modelCfg.ForceGenerate = forceGenerate
	tm[name] = modelCfg
}

type DirectiveConfig struct {
	SkipRuntime bool `yaml:"skip_runtime"`

	// If the directive implementation is statically defined, don't provide a hook for it
	// in the generated server. This is useful for directives that are implemented
	// by plugins or the runtime itself.
	//
	// The function implemmentation should be provided here as a string.
	//
	// The function should have the following signature:
	// func(ctx context.Context, obj any, next graphql.Resolver[, directive arguments if any]) (res any, err error)
	Implementation *string
}

func inStrSlice(haystack []string, needle string) bool {
	for _, v := range haystack {
		if needle == v {
			return true
		}
	}

	return false
}

// findCfg searches for the config file in this directory and all parents up the tree
// looking for the closest match
func findCfg() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("unable to get working dir to findCfg: %w", err)
	}

	cfg := findCfgInDir(dir)

	for cfg == "" && dir != filepath.Dir(dir) {
		dir = filepath.Dir(dir)
		cfg = findCfgInDir(dir)
	}

	if cfg == "" {
		return "", os.ErrNotExist
	}

	return cfg, nil
}

func findCfgInDir(dir string) string {
	for _, cfgName := range cfgFilenames {
		path := filepath.Join(dir, cfgName)
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	return ""
}

func (c *Config) autobind() error {
	if len(c.AutoBind) == 0 {
		return nil
	}

	ps := c.Packages.LoadAll(c.AutoBind...)

	for _, t := range c.Schema.Types {
		if c.Models.UserDefined(t.Name) || c.Models[t.Name].ForceGenerate {
			continue
		}

		for i, p := range ps {
			if p == nil || p.Module == nil {
				return fmt.Errorf("unable to load %s - make sure you're using an import path to a package that exists", c.AutoBind[i])
			}

			autobindType := c.lookupAutobindType(p, t)
			if autobindType != nil {
				c.Models.Add(t.Name, autobindType.Pkg().Path()+"."+autobindType.Name())
				break
			}
		}
	}

	for i, t := range c.Models {
		if t.ForceGenerate {
			continue
		}

		for j, m := range t.Model {
			pkg, typename := code.PkgAndType(m)

			// skip anything that looks like an import path
			if strings.Contains(pkg, "/") {
				continue
			}

			for _, p := range ps {
				if p.Name != pkg {
					continue
				}
				if t := p.Types.Scope().Lookup(typename); t != nil {
					c.Models[i].Model[j] = t.Pkg().Path() + "." + t.Name()
					break
				}
			}
		}
	}

	return nil
}

func (c *Config) lookupAutobindType(p *packages.Package, schemaType *ast.Definition) types.Object {
	// Try binding to either the original schema type name, or the normalized go type name
	for _, lookupName := range []string{schemaType.Name, templates.ToGo(schemaType.Name)} {
		if t := p.Types.Scope().Lookup(lookupName); t != nil {
			return t
		}
	}

	return nil
}

func (c *Config) injectBuiltins() {
	builtins := TypeMap{
		"__Directive":         {Model: StringList{"github.com/99designs/gqlgen/graphql/introspection.Directive"}},
		"__DirectiveLocation": {Model: StringList{"github.com/99designs/gqlgen/graphql.String"}},
		"__Type":              {Model: StringList{"github.com/99designs/gqlgen/graphql/introspection.Type"}},
		"__TypeKind":          {Model: StringList{"github.com/99designs/gqlgen/graphql.String"}},
		"__Field":             {Model: StringList{"github.com/99designs/gqlgen/graphql/introspection.Field"}},
		"__EnumValue":         {Model: StringList{"github.com/99designs/gqlgen/graphql/introspection.EnumValue"}},
		"__InputValue":        {Model: StringList{"github.com/99designs/gqlgen/graphql/introspection.InputValue"}},
		"__Schema":            {Model: StringList{"github.com/99designs/gqlgen/graphql/introspection.Schema"}},
		"Float":               {Model: StringList{"github.com/99designs/gqlgen/graphql.FloatContext"}},
		"String":              {Model: StringList{"github.com/99designs/gqlgen/graphql.String"}},
		"Boolean":             {Model: StringList{"github.com/99designs/gqlgen/graphql.Boolean"}},
		"Int": {
			// FIXME: using int / int64 for Int is not spec compliant and introduces
			// security risks. We should default to int32.
			Model: StringList{
				"github.com/99designs/gqlgen/graphql.Int",
				"github.com/99designs/gqlgen/graphql.Int32",
				"github.com/99designs/gqlgen/graphql.Int64",
			},
		},
		"ID": {
			Model: StringList{
				"github.com/99designs/gqlgen/graphql.ID",
				"github.com/99designs/gqlgen/graphql.IntID",
			},
		},
	}

	for typeName, entry := range builtins {
		if !c.Models.Exists(typeName) {
			c.Models[typeName] = entry
		}
	}

	// These are additional types that are injected if defined in the schema as scalars.
	extraBuiltins := TypeMap{
		"Int64": {
			Model: StringList{
				"github.com/99designs/gqlgen/graphql.Int",
				"github.com/99designs/gqlgen/graphql.Int64",
			},
		},
		"Time":   {Model: StringList{"github.com/99designs/gqlgen/graphql.Time"}},
		"Map":    {Model: StringList{"github.com/99designs/gqlgen/graphql.Map"}},
		"Upload": {Model: StringList{"github.com/99designs/gqlgen/graphql.Upload"}},
		"Any":    {Model: StringList{"github.com/99designs/gqlgen/graphql.Any"}},
	}

	for typeName, entry := range extraBuiltins {
		if t, ok := c.Schema.Types[typeName]; !c.Models.Exists(typeName) && ok && t.Kind == ast.Scalar {
			c.Models[typeName] = entry
		}
	}
}

func (c *Config) LoadSchema() error {
	if c.Packages != nil {
		c.Packages = code.NewPackages(
			code.WithBuildTags(c.GoBuildTags...),
			code.PackagePrefixToCache("github.com/99designs/gqlgen/graphql"),
			code.WithPreloadNames(templatePackageNames...),
		)
	}

	if err := c.check(); err != nil {
		return err
	}

	schema, err := gqlparser.LoadSchema(c.Sources...)
	if err != nil {
		return err
	}

	if schema.Query == nil {
		schema.Query = &ast.Definition{
			Kind: ast.Object,
			Name: "Query",
		}
		schema.Types["Query"] = schema.Query
	}

	c.Schema = schema
	return nil
}

func abs(path string) string {
	absPath, err := filepath.Abs(path)
	if err != nil {
		panic(err)
	}
	return filepath.ToSlash(absPath)
}
