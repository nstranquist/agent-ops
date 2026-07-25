package mcpgate

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// Policy is a simple server/tool allow-deny DSL.
type Policy struct {
	Default string       `yaml:"default" json:"default"` // allow|deny
	Rules   []PolicyRule `yaml:"rules" json:"rules"`
}

type PolicyRule struct {
	Server string `yaml:"server" json:"server"`
	Tool   string `yaml:"tool" json:"tool"` // * = any
	Action string `yaml:"action" json:"action"` // allow|deny
}

// Decision is the primary product observable.
type Decision struct {
	Server  string `json:"server"`
	Tool    string `json:"tool"`
	Action  string `json:"action"` // allow|deny
	Matched string `json:"matched_rule"`
	Reason  string `json:"reason"`
}

// Evaluate returns allow/deny for a tool call.
func Evaluate(p Policy, server, tool string) Decision {
	def := strings.ToLower(strings.TrimSpace(p.Default))
	if def != "allow" && def != "deny" {
		def = "deny"
	}
	server = strings.TrimSpace(server)
	tool = strings.TrimSpace(tool)
	for i, r := range p.Rules {
		if !match(r.Server, server) {
			continue
		}
		if !match(r.Tool, tool) {
			continue
		}
		act := strings.ToLower(strings.TrimSpace(r.Action))
		if act != "allow" && act != "deny" {
			act = "deny"
		}
		return Decision{
			Server:  server,
			Tool:    tool,
			Action:  act,
			Matched: fmt.Sprintf("rules[%d]", i),
			Reason:  fmt.Sprintf("matched server=%q tool=%q", r.Server, r.Tool),
		}
	}
	return Decision{
		Server:  server,
		Tool:    tool,
		Action:  def,
		Matched: "default",
		Reason:  "no rule matched; using default",
	}
}

func match(pattern, value string) bool {
	pattern = strings.TrimSpace(pattern)
	if pattern == "" || pattern == "*" {
		return true
	}
	return strings.EqualFold(pattern, value)
}

func LoadPolicy(path string) (Policy, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return Policy{}, err
	}
	var p Policy
	if err := yaml.Unmarshal(b, &p); err != nil {
		return Policy{}, err
	}
	return p, nil
}
