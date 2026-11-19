package model

import "time"

type TargetCriteria struct {
	TargetType      string   `json:"target_type"`
	TargetValue     float64  `json:"target_value"`
	CategoryFilters []string `json:"category_filters,omitempty"`
	TimeWindowDays  *int     `json:"time_window_days,omitempty"`
}

type RewardDefinition struct {
	RewardType       string  `json:"reward_type"`
	BonusPoints      *int64  `json:"bonus_points,omitempty"`
	VoucherBatchCode *string `json:"voucher_batch_code,omitempty"`
}

type Campaign struct {
	ID               string           `json:"id"`
	CampaignCode     string           `json:"campaign_code"`
	Title            string           `json:"title"`
	Description      string           `json:"description"`
	BannerURL        string           `json:"banner_url,omitempty"`
	CampaignType     string           `json:"campaign_type"`
	Status           string           `json:"status"`
	StartDate        string           `json:"start_date"`
	EndDate          string           `json:"end_date"`
	TargetCriteria   TargetCriteria   `json:"target_criteria"`
	RewardDefinition RewardDefinition `json:"reward_definition"`
	ParticipantCount int64            `json:"participant_count"`
	CompletionCount  int64            `json:"completion_count"`
	CreatedBy        string           `json:"created_by"`
	CreatedAt        time.Time        `json:"created_at"`
	UpdatedAt        time.Time        `json:"updated_at"`
}

type CreateCampaignRequest struct {
	CampaignCode     string           `json:"campaign_code" binding:"required"`
	Title            string           `json:"title" binding:"required"`
	Description      string           `json:"description"`
	BannerURL        string           `json:"banner_url"`
	CampaignType     string           `json:"campaign_type" binding:"required"`
	StartDate        string           `json:"start_date" binding:"required"`
	EndDate          string           `json:"end_date" binding:"required"`
	TargetCriteria   TargetCriteria   `json:"target_criteria" binding:"required"`
	RewardDefinition RewardDefinition `json:"reward_definition" binding:"required"`
}

type UpdateCampaignRequest = CreateCampaignRequest

type UpdateCampaignStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

