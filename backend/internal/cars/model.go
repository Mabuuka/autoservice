package cars

type Car struct {
	CarID           int64  `json:"car_id"`
	OwnerID         int64  `json:"owner_id"`
	OwnerFullName   string `json:"owner_full_name"`
	Brand           string `json:"brand"`
	PlateNumber     string `json:"plate_number"`
	ManufactureYear int    `json:"manufacture_year"`
	Color           string `json:"color"`
}

type CreateCarInput struct {
	OwnerID         int64  `json:"owner_id"`
	Brand           string `json:"brand"`
	PlateNumber     string `json:"plate_number"`
	ManufactureYear int    `json:"manufacture_year"`
	Color           string `json:"color"`
}
