package maintenanceparts

type MaintenancePart struct {
	MaintenancePartID int64  `json:"maintenance_part_id"`
	Name              string `json:"name"`
	Quantity          int    `json:"quantity"`
	DeliveryDate      string `json:"delivery_date"`
}

type CreateMaintenancePartInput struct {
	Name         string `json:"name"`
	Quantity     int    `json:"quantity"`
	DeliveryDate string `json:"delivery_date"`
}
