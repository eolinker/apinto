package price_calcular

type PricingData struct {
	BasicInfo *BasicInfo            `json:"basic_info"`
	Strategy  map[string]*PricePlan `json:"strategy"`
}

type BasicInfo struct {
	Version         string `json:"version"`
	Rely            string `json:"rely"`
	ResourceGroupID string `json:"resource_group_id"`
	TenantID        string `json:"tenant_id"`
}

type PricePlan struct {
	Cost     map[string]float64 `json:"cost"`
	Sale     map[string]float64 `json:"sale"`
	Official map[string]float64 `json:"official"`
}
