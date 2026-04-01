package employees

type Employee struct {
	EmployeeID      int64  `json:"employee_id"`
	PersonnelNumber int    `json:"personnel_number"`
	Specialty       string `json:"specialty"`
	Phone           string `json:"phone"`
	FullName        string `json:"full_name"`
}

type CreateEmployeeInput struct {
	PersonnelNumber int    `json:"personnel_number"`
	Specialty       string `json:"specialty"`
	Phone           string `json:"phone"`
	FullName        string `json:"full_name"`
}

type UpdateEmployeeInput struct {
	PersonnelNumber int    `json:"personnel_number"`
	Specialty       string `json:"specialty"`
	Phone           string `json:"phone"`
	FullName        string `json:"full_name"`
}
