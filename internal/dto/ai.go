package dto

type AIRecommendRequest struct {
	Query string `json:"query"`
}

type AIRecommendation struct {
	ProductID int    `json:"product_id"`
	Reason    string `json:"reason"`
}

type AIRecommendResponse struct {
	Recommendations []AIRecommendation `json:"recommendations"`
}
