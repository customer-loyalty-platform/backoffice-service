package model

import "time"

type RuleCondition struct {
	Field    string      `json:"field"`
	Operator string      `json:"operator"`
	Value    interface{} `json:"value"`
}

type RuleConditionGroup struct {
	And []RuleCondition `json:"and,omitempty"`
	Or  []RuleCondition `json:"or,omitempty"`
}

type RuleAction struct {
	Type             string   `json:"type"`
	MultiplierFactor *float64 `json:"multiplier_factor,omitempty"`
	BonusPoints      *int64   `json:"bonus_points,omitempty"`
	MaxPointsCap     *int64   `json:"max_points_cap,omitempty"`
}

type DynamicRule struct {
	ID             string             `json:"id"`
	RuleCode       string             `json:"rule_code"`
	RuleName       string             `json:"rule_name"`
	Status         string             `json:"status"`
	Priority       int                `json:"priority"`
	StartTime      string             `json:"start_time"`
	EndTime        string             `json:"end_time"`
	ConditionsJSON RuleConditionGroup `json:"conditions_json"`
	ActionJSON     RuleAction         `json:"action_json"`
	CreatedBy      string             `json:"created_by"`
	CreatedAt      time.Time          `json:"created_at"`
	UpdatedAt      time.Time          `json:"updated_at"`
}

type CreateRuleRequest struct {
	RuleCode       string             `json:"rule_code" binding:"required"`
	RuleName       string             `json:"rule_name" binding:"required"`
	Priority       int                `json:"priority"`
	StartTime      string             `json:"start_time"`
	EndTime        string             `json:"end_time"`
	ConditionsJSON RuleConditionGroup `json:"conditions_json" binding:"required"`
	ActionJSON     RuleAction         `json:"action_json" binding:"required"`
}

type UpdateRuleRequest = CreateRuleRequest

type UpdateRuleStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

type CreateRuleResponse struct {
	RuleID   string `json:"rule_id"`
	RuleCode string `json:"rule_code"`
	State    string `json:"state"`
}

