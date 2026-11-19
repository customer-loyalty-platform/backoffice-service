package model

import "time"

type MakerInfo struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Role     string `json:"role"`
}

type CheckerInfo struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

type PendingAction struct {
	ID              string                 `json:"id"`
	Maker           MakerInfo              `json:"maker"`
	Checker         *CheckerInfo           `json:"checker,omitempty"`
	ActionType      string                 `json:"action_type"`
	TargetResource  string                 `json:"target_resource"`
	Payload         map[string]interface{} `json:"payload"`
	Status          string                 `json:"status"`
	RejectionReason *string                `json:"rejection_reason,omitempty"`
	CreatedAt       time.Time              `json:"created_at"`
	ResolvedAt      *time.Time             `json:"resolved_at,omitempty"`
}

type ApproveActionResponse struct {
	Status      string `json:"status"`
	ActionID    string `json:"action_id"`
	ExecutedAt  string `json:"executed_at"`
	Message     string `json:"message"`
}

type RejectActionRequest struct {
	RejectionReason string `json:"rejection_reason" binding:"required"`
}

