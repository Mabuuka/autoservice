package users

type User struct {
	UserID              int64  `json:"user_id"`
	Role                string `json:"role"`
	OwnerID             *int64 `json:"owner_id,omitempty"`
	EmployeeID          *int64 `json:"employee_id,omitempty"`
	Email               string `json:"email"`
	FullName            string `json:"full_name"`
	Phone               string `json:"phone"`
	PreferredEntrypoint string `json:"preferred_entrypoint"`
	CreatedAt           string `json:"created_at"`
	UpdatedAt           string `json:"updated_at"`
	LastLoginAt         string `json:"last_login_at,omitempty"`
}

type AuthUser struct {
	UserID              int64
	Role                string
	OwnerID             *int64
	EmployeeID          *int64
	Email               string
	PasswordHash        string
	FullName            string
	Phone               string
	PreferredEntrypoint string
	CreatedAt           string
	UpdatedAt           string
	LastLoginAt         string
}

type OwnerProfile struct {
	OwnerID       int64  `json:"owner_id"`
	FullName      string `json:"full_name"`
	Phone         string `json:"phone"`
	DriverLicense string `json:"driver_license"`
	IsRegular     bool   `json:"is_regular"`
}

type EmployeeProfile struct {
	EmployeeID      int64  `json:"employee_id"`
	PersonnelNumber int    `json:"personnel_number"`
	Specialty       string `json:"specialty"`
	Phone           string `json:"phone"`
	FullName        string `json:"full_name"`
}

type Profile struct {
	User     User             `json:"user"`
	Owner    *OwnerProfile    `json:"owner,omitempty"`
	Employee *EmployeeProfile `json:"employee,omitempty"`
}

type RegisterInput struct {
	Email         string `json:"email"`
	Password      string `json:"password"`
	FullName      string `json:"full_name"`
	Phone         string `json:"phone"`
	DriverLicense string `json:"driver_license"`
}

type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UpdateProfileInput struct {
	FullName            string `json:"full_name"`
	Phone               string `json:"phone"`
	PreferredEntrypoint string `json:"preferred_entrypoint"`
	DriverLicense       string `json:"driver_license,omitempty"`
}
