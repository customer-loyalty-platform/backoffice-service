package model

import (
	"time"
)

type CurrentTier struct {
	ID       string `json:"id"`
	TierName string `json:"tier_name"`
	Rank     int    `json:"rank"`
}

type PointBalanceSnapshot struct {
	TotalPoints    int64 `json:"total_points"`
	LockedPoints   int64 `json:"locked_points"`
	ExpiringPoints int64 `json:"expiring_points"`
}

type LinkedAccount struct {
	ID                string `json:"id"`
	AccountType       string `json:"account_type"`
	AccountIdentifier string `json:"account_identifier"`
	LinkedAt          string `json:"linked_at"`
}

type PointLedgerEntry struct {
	ID             string    `json:"id"`
	MemberID       string    `json:"member_id"`
	Type           string    `json:"type"`
	Amount         int64     `json:"amount"`
	BalanceAfter   int64     `json:"balance_after"`
	ReferenceID    string    `json:"reference_id"`
	IdempotencyKey string    `json:"idempotency_key"`
	CreatedAt      time.Time `json:"created_at"`
}

type MemberListItem struct {
	ID            string               `json:"id"`
	AccountNumber string               `json:"account_number"`
	FullName      string               `json:"full_name"`
	Phone         string               `json:"phone"`
	Email         string               `json:"email"`
	CurrentTier   CurrentTier          `json:"current_tier"`
	Balance       PointBalanceSnapshot `json:"balance"`
	Status        string               `json:"status"`
	CreatedAt     time.Time            `json:"created_at"`
	UpdatedAt     time.Time            `json:"updated_at"`
}

type MemberDetail struct {
	MemberListItem
	LifetimeEarned       int64              `json:"lifetime_earned"`
	LifetimeRedeemed     int64              `json:"lifetime_redeemed"`
	TierQualifyingPoints int64              `json:"tier_qualifying_points"`
	TierEvaluationDate   string             `json:"tier_evaluation_date"`
	LinkedAccounts       []LinkedAccount    `json:"linked_accounts"`
	RecentTransactions   []PointLedgerEntry `json:"recent_transactions"`
}

type UpdateMemberStatusRequest struct {
	Status string `json:"status" binding:"required"`
	Reason string `json:"reason"`
}

type UpdateMemberStatusResponse struct {
	MemberID       string    `json:"member_id"`
	PreviousStatus string    `json:"previous_status"`
	NewStatus      string    `json:"new_status"`
	UpdatedBy      string    `json:"updated_by"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type PointAdjustmentRequest struct {
	AdjustmentType    string `json:"adjustment_type" binding:"required"`
	Amount            int64  `json:"amount" binding:"required"`
	ReasonCode        string `json:"reason_code" binding:"required"`
	Notes             string `json:"notes"`
	ExternalTicketRef string `json:"external_ticket_ref"`
}

type PointAdjustmentResponse struct {
	Status          string `json:"status"`
	ActionID        string `json:"action_id"`
	RequiresChecker bool   `json:"requires_checker"`
	MakerID         string `json:"maker_id"`
}

