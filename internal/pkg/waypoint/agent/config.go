package agent

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/buildkite/shellwords"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclsimple"
	"github.com/zclconf/go-cty/cty"
)

type Config struct {
	groupNames []string
	groups     map[string]*hclGroup

	forceShell string
	opWrapper  func(Operation) Operation
}

func ParseConfig(input string) (*Config, error) {
	var hc hclConfig

	var ctx hcl.EvalContext

	err := hclsimple.Decode("blah.hcl", []byte(input), &ctx, &hc)
	if err != nil {
		return nil, err
	}

	var cfg Config
	cfg.groups = make(map[string]*hclGroup)
	cfg.opWrapper = func(o Operation) Operation { return o }

	for _, grp := range hc.Groups {
		cfg.groupNames = append(cfg.groupNames, grp.Name)
		cfg.groups[grp.Name] = grp
	}

	sort.Strings(cfg.groupNames)

	return &cfg, nil
}

func ParseConfigFile(path string) (*Config, error) {
	var hc hclConfig

	var ctx hcl.EvalContext

	input, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	err = hclsimple.Decode(path, input, &ctx, &hc)
	if err != nil {
		return nil, err
	}

	var cfg Config
	cfg.groups = make(map[string]*hclGroup)
	cfg.opWrapper = func(o Operation) Operation { return o }

	for _, grp := range hc.Groups {
		cfg.groupNames = append(cfg.groupNames, grp.Name)
		cfg.groups[grp.Name] = grp
	}

	sort.Strings(cfg.groupNames)

	return &cfg, nil
}

func (c *Config) Groups() []string {
	return c.groupNames
}

func (c *Config) IsAvailable(group, id string) (bool, error) {
	grp, ok := c.groups[group]
	if !ok {
		return false, nil
	}

	for _, act := range grp.Actions {
		if act.Name == id {
			return true, nil
		}
	}

	return false, nil
}

func (c *Config) Action(group, id string, hctx *hcl.EvalContext) (Operation, error) {
	grp, ok := c.groups[group]
	if !ok {
		return nil, nil
	}

	if hctx == nil {
		hctx = &hcl.EvalContext{}
	}

	for _, act := range grp.Actions {
		if act.Name == id {
			return c.convertAction(hctx, act.Body)
		}
	}

	return nil, nil
}

var actionSchema = hcl.BodySchema{
	Blocks: []hcl.BlockHeaderSchema{
		{
			Type: "run",
		},
		{
			Type: "http",
		},
		{
			Type: "status",
		},
		{
			Type: "operation",
		},
	},
}

func (c *Config) convertAction(hctx *hcl.EvalContext, body hcl.Body) (Operation, error) {
	content, diag := body.Content(&actionSchema)
	if diag.HasErrors() {
		return nil, diag
	}

	for _, blk := range content.Blocks {
		switch blk.Type {
		case "run":
			return c.convertRun(hctx, blk)
		case "http":
			return c.convertHTTPAction(hctx, blk)
		case "status":
			return c.convertStatus(hctx, blk)
		}
	}

	// See if there operations

	var ops []*hcl.Block

	for _, blk := range content.Blocks {
		if blk.Type == "operation" {
			ops = append(ops, blk)
		}
	}

	if len(ops) > 0 {
		return c.convertCompound(hctx, ops)
	}

	return nil, fmt.Errorf("no operation specified")
}

func (c *Config) convertCompound(hctx *hcl.EvalContext, blks []*hcl.Block) (Operation, error) {
	var co CompoundOperation

	for _, blk := range blks {
		op, err := c.convertAction(hctx, blk.Body)
		if err != nil {
			return nil, err
		}

		co.Operations = append(co.Operations, op)
	}

	return c.opWrapper(&co), nil
}

var runActionSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{
			Name:     "command",
			Required: true,
		},
		{
			Name: "env",
		},
	},
	Blocks: []hcl.BlockHeaderSchema{
		{
			Type: "docker",
		},
	},
}

const posixSpecialChars = "!\"#$&'()*,;<=>?[]\\^`{}|~"

func (c *Config) convertRun(hctx *hcl.EvalContext, blk *hcl.Block) (Operation, error) {
	body, diag := blk.Body.Content(runActionSchema)
	if diag.HasErrors() {
		return nil, diag
	}

	val, diag := body.Attributes["command"].Expr.Value(hctx)
	if diag.HasErrors() {
		return nil, diag
	}

	var (
		words []string
		err   error
	)

	if val.Type() == cty.String {
		str := val.AsString()

		if strings.ContainsAny(str, posixSpecialChars) {
			shell := c.forceShell

			if shell == "" {
				shell = os.Getenv("SHELL")
			}

			if shell == "" {
				shell = "sh"
			}

			words = append(words, shell, "-c", str)
		} else {
			words, err = shellwords.Split(val.AsString())
			if err != nil {
				return nil, err
			}
		}
	} else if val.Type().IsTupleType() {
		for _, v := range val.AsValueSlice() {
			switch v.Type() {
			case cty.String:
				words = append(words, v.AsString())
			case cty.Number:
				words = append(words, v.AsBigFloat().String())
			default:
				return nil, fmt.Errorf("unsupported value type in arguments: %s", v.Type().GoString())
			}
		}
	}

	var do *DockerOptions

	if blks, ok := body.Blocks.ByType()["docker"]; ok {
		blk := blks[0]

		do = &DockerOptions{}

		diag := gohcl.DecodeBody(blk.Body, hctx, do)
		if diag.HasErrors() {
			return nil, diag
		}
	}

	env := map[string]string{}

	if v, ok := body.Attributes["env"]; ok {
		diag := gohcl.DecodeExpression(v.Expr, hctx, &env)
		if diag.HasErrors() {
			return nil, diag
		}
	}

	return c.opWrapper(&ShellOperation{
		Arguments:     words,
		Environment:   env,
		DockerOptions: do,
	}), nil
}

var httpActionSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{
			Name:     "url",
			Required: true,
		},
	},
}

func (c *Config) convertHTTPAction(hctx *hcl.EvalContext, blk *hcl.Block) (Operation, error) {
	body, diag := blk.Body.Content(httpActionSchema)
	if diag.HasErrors() {
		return nil, diag
	}

	val, diag := body.Attributes["url"].Expr.Value(hctx)
	if diag.HasErrors() {
		return nil, diag
	}

	if val.Type() != cty.String {
		return nil, fmt.Errorf("url must be a string")
	}

	return c.opWrapper(&HTTPOperation{
		URL: val.AsString(),
	}), nil
}

var statusSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{
			Name:     "message",
			Required: true,
		},
		{
			Name: "values",
		},
		{
			Name: "status",
		},
	},
}

func (c *Config) convertStatus(hctx *hcl.EvalContext, blk *hcl.Block) (Operation, error) {
	body, diag := blk.Body.Content(statusSchema)
	if diag.HasErrors() {
		return nil, diag
	}

	val, diag := body.Attributes["message"].Expr.Value(hctx)
	if diag.HasErrors() {
		return nil, diag
	}

	if val.Type() != cty.String {
		return nil, fmt.Errorf("message must be a string")
	}

	so := &StatusOperation{
		Message: val.AsString(),
	}

	if attr, ok := body.Attributes["status"]; ok {
		val, diag := attr.Expr.Value(hctx)
		if diag.HasErrors() {
			return nil, diag
		}

		if val.Type() != cty.String {
			return nil, fmt.Errorf("message must be a string")
		}

		so.Status = val.AsString()
	}

	if values, ok := body.Attributes["values"]; ok {
		so.Values = make(map[string]string)

		m, diag := values.Expr.Value(hctx)
		if diag.HasErrors() {
			return nil, diag
		}

		if !m.Type().IsObjectType() {
			return nil, fmt.Errorf("values must be a map/object")
		}

		i := m.ElementIterator()

		for i.Next() {
			key, val := i.Element()

			switch {
			case val.Type().Equals(cty.String):
				so.Values[key.AsString()] = val.AsString()

			case val.Type().Equals(cty.Number):
				bf := val.AsBigFloat()

				if bf.IsInt() {
					i, _ := bf.Int64()
					so.Values[key.AsString()] = fmt.Sprintf("%d", i)
				} else {
					f, _ := bf.Float64()
					so.Values[key.AsString()] = fmt.Sprintf("%f", f)
				}
			default:
				return nil, fmt.Errorf("values can only be strings or numbers")
			}
		}
	}

	return c.opWrapper(so), nil
}

type hclAction struct {
	Name string   `hcl:",label"`
	Body hcl.Body `hcl:",remain"`
}

type hclGroup struct {
	Name    string       `hcl:",label"`
	Actions []*hclAction `hcl:"action,block"`
}

type hclConfig struct {
	Groups []*hclGroup `hcl:"group,block"`
}
