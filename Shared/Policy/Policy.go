package Policy

import "time"

type Action string

const (
	ActionLogin       Action = "login"
	ActionCreateOrder Action = "create_order"
)

const (
	PlanFree  = "FREE_PLAN_USER"
	PlanPro   = "PRO_PLAN_USER"
	PlanAdmin = "ADMIN"
)

type Rule struct {
	Limit  int
	Window time.Duration
}

var Unlimited = Rule{Limit: 0}

type Provider interface {
	RuleFor(plan string, action Action) Rule
}

// пока хардкод, позже уберу в конфиг

type StaticProvider struct {
	rules map[string]map[Action]Rule
}

func NewStaticProvider() *StaticProvider {
	return &StaticProvider{
		rules: map[string]map[Action]Rule{
			PlanFree: {
				ActionLogin:       {Limit: 100, Window: time.Hour},
				ActionCreateOrder: {Limit: 10, Window: 24 * time.Hour},
			},
			PlanPro:   {},
			PlanAdmin: {},
		},
	}
}

func (p *StaticProvider) RuleFor(plan string, action Action) Rule {
	actions, ok := p.rules[plan]
	if !ok {
		actions = p.rules[PlanFree]
	}
	if rule, ok := actions[action]; ok {
		return rule
	}
	return Unlimited
}
