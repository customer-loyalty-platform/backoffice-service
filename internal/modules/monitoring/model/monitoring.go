package model

import "time"

type MemberMetrics struct {
	TotalRegistered int64 `json:"total_registered"`
	Active30d       int64 `json:"active_30d"`
	Active60d       int64 `json:"active_60d"`
	Active90d       int64 `json:"active_90d"`
	NewToday        int64 `json:"new_today"`
}

type PointMetrics struct {
	TotalCirculating int64 `json:"total_circulating"`
	EarnedToday      int64 `json:"earned_today"`
	RedeemedToday    int64 `json:"redeemed_today"`
	ExpiredToday     int64 `json:"expired_today"`
}

type VoucherMetrics struct {
	TotalActiveBatches  int64 `json:"total_active_batches"`
	TotalAvailableStock int64 `json:"total_available_stock"`
	ClaimedToday        int64 `json:"claimed_today"`
}

type CampaignMetrics struct {
	TotalActive       int64 `json:"total_active"`
	TotalParticipants int64 `json:"total_participants"`
	CompletionsToday  int64 `json:"completions_today"`
}

type DashboardSummary struct {
	Members   MemberMetrics   `json:"members"`
	Points    PointMetrics    `json:"points"`
	Vouchers  VoucherMetrics  `json:"vouchers"`
	Campaigns CampaignMetrics `json:"campaigns"`
}

type ChartDataPoint struct {
	Date          string `json:"date"`
	TotalEarned   int64  `json:"total_earned"`
	TotalRedeemed int64  `json:"total_redeemed"`
	TotalExpired  int64  `json:"total_expired"`
}

type AdminActor struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Role     string `json:"role"`
}

type AuditLogEntry struct {
	ID             string                 `json:"id"`
	Admin          AdminActor             `json:"admin"`
	Action         string                 `json:"action"`
	TargetResource string                 `json:"target_resource"`
	OldValue       map[string]interface{} `json:"old_value,omitempty"`
	NewValue       map[string]interface{} `json:"new_value,omitempty"`
	IPAddress      string                 `json:"ip_address"`
	CreatedAt      time.Time              `json:"created_at"`
}

