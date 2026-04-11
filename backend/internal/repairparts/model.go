package repairparts

type RepairPart struct {
	RepairPartID int64  `json:"repair_part_id"`
	Name         string `json:"name"`
	Quantity     int    `json:"quantity"`
	DeliveryDate string `json:"delivery_date"`
}

type CreateRepairPartInput struct {
	Name         string `json:"name"`
	Quantity     int    `json:"quantity"`
	DeliveryDate string `json:"delivery_date"`
}

type RestockRepairPartInput struct {
	Quantity     int    `json:"quantity"`
	DeliveryDate string `json:"delivery_date"`
}
