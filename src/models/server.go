package models

type ChallengeResponse struct {
	Challenge string `json:"challenge"`
}

type PopularityRequest struct {
	Options []string `form:"options[]" binding:"required,gt=0"`
}

type PopularityResponse struct {
	Options map[string]int64 `json:"options"`
}
