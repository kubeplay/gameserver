package auth

import (
	"github.com/casbin/casbin"
	"github.com/casbin/casbin/persist/file-adapter"
)

const (
	casbinModel = `
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = r.sub == p.sub && keyMatch2(r.obj, p.obj) && regexMatch(r.act, p.act)
`
)

var (
	guestPerms = []interface{}{
		[]string{"/v1/policies", "GET"},
		[]string{"/v1/events", "GET"},
		[]string{"/v1/events:resourceName", "GET"},
		[]string{"/v1/events/:parent/games", "(GET)|(POST)"},
		[]string{"/v1/events/:parent/games/:resourceName", "GET"},
		[]string{"/v1/events/:parent/games/:resourceName/solve", "POST"},
	}
	hostPerms = []interface{}{
		[]string{"/v1/policies", "(GET)|(POST)"},
		[]string{"/v1/challenges", "(GET)|(POST)"},
		[]string{"/v1/policies/:resourceName,", "(GET)|(DELETE)|(PUT)"},
		[]string{"/v1/challenges", "(GET)|(POST)"},
		[]string{"/v1/challenges/:resourceName", "(GET)|(PUT)|(DELETE)"},
		[]string{"/v1/events", "(GET)|(POST)"},
		[]string{"/v1/events:resourceName", "(GET)|(PUT)|(DELETE)"},
		[]string{"/v1/events/:parent/games", "(GET)|(POST)"},
		[]string{"/v1/events/:parent/games/:resourceName", "(GET)|(DELETE)"},
		[]string{"/v1/events/:parent/games/:resourceName/solve", "POST"},
		[]string{"/v1/events/:parent/games/:resourceName/start", "POST"},
	}
)

func NewUserPolicy(policyPath, username string, isHost bool) error {
	e, err := casbin.NewEnforcerSafe(
		casbin.NewModel(casbinModel),
		fileadapter.NewAdapter(policyPath),
	)
	if err != nil {
		return err
	}
	if isHost {
		for _, perms := range hostPerms {
			newPerms := append([]string{username}, perms.([]string)...)
			e.AddPolicySafe([]interface{}{newPerms}...)
		}
	} else {
		for _, perms := range guestPerms {
			newPerms := append([]string{username}, perms.([]string)...)
			e.AddPolicySafe([]interface{}{newPerms}...)
		}
	}
	return e.SavePolicy()
}

func NewEnforcer(policyPath string) (*casbin.Enforcer, error) {
	return casbin.NewEnforcerSafe(
		casbin.NewModel(casbinModel),
		fileadapter.NewAdapter(policyPath),
	)
}
