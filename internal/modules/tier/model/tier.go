package model

import "time"

type TierPerk struct {
	PerkCode    string `json:"perk_code"`
	PerkName    string `json:"perk_name"`
	Description string `json:"description"`
	IsActive    bool   `json:"is_active"`
}

type TierDefinition struct {
	ID                string     `json:"id"`
	TierName          string     `json:"tier_name"`
	Rank              int        `json:"rank"`
	QualificationType string     `json:"qualification_type"`
	MinThreshold      int64      `json:"min_threshold"`
	Multiplier        float64    `json:"multiplier"`
	PerksConfig       []TierPerk `json:"perks_config"`
	MemberCount       int64      `json:"member_count"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

type CreateTierRequest struct {
	TierName          string     `json:"tier_name" binding:"required"`
	Rank              int        `json:"rank" binding:"required"`
	QualificationType string     `json:"qualification_type" binding:"required"`
	MinThreshold      int64      `json:"min_threshold"`
	Multiplier        float64    `json:"multiplier" binding:"required"`
	PerksConfig       []TierPerk `json:"perks_config"`
}

type UpdateTierRequest = CreateTierRequest

