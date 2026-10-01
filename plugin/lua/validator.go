package lua

import (
	"embed"
	"fmt"
	"regexp"
	"strings"

	"github.com/yuin/gopher-lua/ast"
	"github.com/yuin/gopher-lua/parse"
	"sigs.k8s.io/yaml"
)

//go:embed policies/*.yaml
var defaultPolicies embed.FS

type Violation struct {
	ID       string `yaml:"id"`
	Message  string `yaml:"message"`
	Severity string `yaml:"severity"`
	Line     int    `yaml:"line"` // Added line number
}

type Rule struct {
	ID          string `yaml:"id"`
	Name        string `yaml:"name"`
	Pattern     string `yaml:"pattern"`
	Severity    string `yaml:"severity"`
	Message     string `yaml:"message"`
	compiledReg *regexp.Regexp
}

type Profile struct {
	Name            string   `yaml:"name"`
	Description     string   `yaml:"description"`
	Extends         string   `yaml:"extends"`
	AllowedGlobals  []string `yaml:"allowed_globals"`
	AllowedPrefixes []string `yaml:"allowed_prefixes"`
}

type PolicyStore struct {
	Version  string    `yaml:"version"`
	Edition  string    `yaml:"edition"`
	Profiles []Profile `yaml:"profiles"`
	Rules    []Rule    `yaml:"rules"`
}

type Validator struct {
	store *PolicyStore
}

func NewValidator(edition string, _ string) (*Validator, error) {
	var data []byte
	var err error

	var fileName string
	switch edition {
	case "oss":
		fileName = "policies/kong_oss_3x.yaml"
	case "ee":
		fileName = "policies/kong_ee_3x.yaml"
	default:
		return nil, fmt.Errorf("unsupported Kong edition: %s", edition)
	}

	data, err = defaultPolicies.ReadFile(fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to read policy file %s: %w", fileName, err)
	}

	store := &PolicyStore{}
	if err := yaml.Unmarshal(data, store); err != nil {
		return nil, fmt.Errorf("failed to parse policy YAML: %w", err)
	}

	for i := range store.Rules {
		re, err := regexp.Compile(store.Rules[i].Pattern)
		if err != nil {
			return nil, fmt.Errorf("invalid regex in rule %s: %w", store.Rules[i].ID, err)
		}
		store.Rules[i].compiledReg = re
	}

	return &Validator{store: store}, nil
}

// resolveProfile builds a complete profile by merging current with all ancestors.
func (v *Validator) resolveProfile(name string) (*Profile, error) {
	var base *Profile
	for i := range v.store.Profiles {
		if v.store.Profiles[i].Name == name {
			base = &v.store.Profiles[i]
			break
		}
	}
	if base == nil {
		return nil, fmt.Errorf("profile %s not found", name)
	}

	// Deep copy base data
	res := &Profile{
		Name:            base.Name,
		AllowedGlobals:  append([]string{}, base.AllowedGlobals...),
		AllowedPrefixes: append([]string{}, base.AllowedPrefixes...),
	}

	// Recursive merge of parents
	if base.Extends != "" {
		parent, err := v.resolveProfile(base.Extends)
		if err != nil {
			return nil, err
		}
		res.AllowedGlobals = append(res.AllowedGlobals, parent.AllowedGlobals...)
		res.AllowedPrefixes = append(res.AllowedPrefixes, parent.AllowedPrefixes...)
	}
	return res, nil
}

func (v *Validator) Validate(code string, profileName string) ([]Violation, error) {
	profile, err := v.resolveProfile(profileName)
	if err != nil {
		return nil, err
	}

	violations := []Violation{}

	// Phase 1: Security Scan (Regex)
	for _, rule := range v.store.Rules {
		if rule.compiledReg.MatchString(code) {
			violations = append(violations, Violation{
				ID:       rule.ID,
				Message:  rule.Message,
				Severity: rule.Severity,
				Line:     0, // Regex doesn't provide line info easily
			})
		}
	}

	// Phase 2: Semantic Analysis (AST)
	astViolations := v.validateAST(code, profile)
	violations = append(violations, astViolations...)

	return violations, nil
}

func (v *Validator) validateAST(code string, profile *Profile) []Violation {
	violations := []Violation{}
	stats, err := parse.Parse(strings.NewReader(code), "<input>")
	if err != nil {
		violations = append(violations, Violation{
			ID: "LUA-SYNTAX", Message: fmt.Sprintf("Lua syntax error: %v", err), Severity: "ERROR",
		})
		return violations
	}

	inspector := func(expr ast.Expr) {
		if call, ok := expr.(*ast.FuncCallExpr); ok {
			var funcName string
			if call.Receiver != nil && call.Method != "" {
				funcName = v.extractFullName(call.Receiver) + ":" + call.Method
			} else {
				funcName = v.extractFullName(call.Func)
			}

			if funcName != "" && !v.isAllowed(funcName, profile) {
				violations = append(violations, Violation{
					ID:       "LUA-WHITELIST",
					Message:  fmt.Sprintf("Forbidden call detected: '%s' is not allowed in %s profile", funcName, profile.Name),
					Severity: "ERROR",
					Line:     call.Line(), // Extract line from AST
				})
			}
		}
	}

	for _, stmt := range stats {
		v.walkStatement(stmt, inspector)
	}
	return violations
}

func (v *Validator) walkStatement(stmt ast.Stmt, inspect func(ast.Expr)) {
	if stmt == nil {
		return
	}
	switch s := stmt.(type) {
	case *ast.FuncCallStmt:
		v.walkExpr(s.Expr, inspect)
	case *ast.AssignStmt:
		for _, expr := range s.Rhs {
			v.walkExpr(expr, inspect)
		}
	case *ast.LocalAssignStmt:
		for _, expr := range s.Exprs {
			v.walkExpr(expr, inspect)
		}
	case *ast.IfStmt:
		v.walkExpr(s.Condition, inspect)
		for _, sub := range s.Then {
			v.walkStatement(sub, inspect)
		}
		for _, sub := range s.Else {
			v.walkStatement(sub, inspect)
		}
	}
}

func (v *Validator) walkExpr(expr ast.Expr, inspect func(ast.Expr)) {
	if expr == nil {
		return
	}
	inspect(expr)

	switch e := expr.(type) {
	case *ast.FuncCallExpr:
		for _, arg := range e.Args {
			v.walkExpr(arg, inspect)
		}
	case *ast.AttrGetExpr: // Handle recursive index access: obj[1].prop
		v.walkExpr(e.Object, inspect)
		v.walkExpr(e.Key, inspect)
	case *ast.LogicalOpExpr:
		v.walkExpr(e.Lhs, inspect)
		v.walkExpr(e.Rhs, inspect)
	}
}

func (v *Validator) extractFullName(expr ast.Expr) string {
	if expr == nil {
		return ""
	}
	switch e := expr.(type) {
	case *ast.IdentExpr:
		return e.Value
	case *ast.StringExpr:
		return e.Value
	case *ast.AttrGetExpr:
		left := v.extractFullName(e.Object)
		if left == "" {
			return ""
		}
		right := v.extractFullName(e.Key)
		if right == "" {
			return left + ".*"
		}
		return left + "." + right
	}
	return ""
}

func (v *Validator) isAllowed(name string, profile *Profile) bool {
	for _, g := range profile.AllowedGlobals {
		if name == g {
			return true
		}
	}
	for _, p := range profile.AllowedPrefixes {
		if strings.HasPrefix(name, p) {
			return true
		}
	}
	return false
}

func (v *Validator) GetEdition() string {
	return v.store.Edition
}
