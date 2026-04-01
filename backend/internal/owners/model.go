package owners

type Owner struct {
	OwnerID       int64  `json:"owner_id"`
	FullName      string `json:"full_name"`
	Phone         string `json:"phone"`
	DriverLicense string `json:"driver_license"`
	IsRegular     bool   `json:"is_regular"`
}

type CreateOwnerInput struct {
	FullName      string `json:"full_name"`
	Phone         string `json:"phone"`
	DriverLicense string `json:"driver_license"`
	IsRegular     bool   `json:"is_regular"`
}

type UpdateOwnerInput struct {
	FullName      string `json:"full_name"`
	Phone         string `json:"phone"`
	DriverLicense string `json:"driver_license"`
	IsRegular     bool   `json:"is_regular"`
}
