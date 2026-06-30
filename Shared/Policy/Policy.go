package Policy

import (
	"fmt"
	"gRPCbigapp/Shared/Auth/AuthRoles"
	"time"
)

type Action string

const (
	ActionLogin       Action = "login"
	ActionCreateOrder Action = "create_order"
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
	rules map[AuthRoles.Plan]map[Action]Rule
}

func (p *StaticProvider) validation() error {
	for _, plan := range AuthRoles.All() {
		if _, ok := p.rules[plan]; ok {
			return fmt.Errorf("policy, для роли не заданы правила: %s", plan)
		}
	}
	return nil
}

func NewStaticProvider() (*StaticProvider, error) {
	provide := &StaticProvider{
		rules: map[AuthRoles.Plan]map[Action]Rule{
			AuthRoles.Free: {
				ActionLogin:       {Limit: 100, Window: time.Hour},
				ActionCreateOrder: {Limit: 10, Window: 24 * time.Hour},
			},
			AuthRoles.Pro:   {},
			AuthRoles.Admin: {},
		},
	}
	if err := provide.validation(); err != nil {
		return nil, err
	}
	return provide, nil
}

func (p *StaticProvider) RuleFor(plan string, action Action) Rule {
	actions, ok := p.rules[AuthRoles.Plan(plan)]
	if !ok {
		actions = p.rules[AuthRoles.Free]
	}
	if rule, ok := actions[action]; ok {
		return rule
	}
	return Unlimited
}
