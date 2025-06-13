package dto

// UserRegistrationDTO separates API input concerns from the database model
type UserRegistrationDTO struct {
	Username  string `json:"username" binding:"required,min=3,max=50"`
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=6"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Role      string `json:"role" binding:"required"`
}

// UserLoginDTO for login requests
type UserLoginDTO struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// UserResponseDTO for responses without sensitive information
type UserResponseDTO struct {
	ID        uint   `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Role      string `json:"role"`
}

// UpdateUserDTO for profile updates
type UpdateUserDTO struct {
	Email     *string `json:"email" binding:"omitempty,email"`
	FirstName *string `json:"first_name"`
	LastName  *string `json:"last_name"`
	Password  *string `json:"password" binding:"omitempty,min=6"`
	Bio       *string `json:"bio"`
	Avatar    *string `json:"avatar"`
	Website   *string `json:"website"`
	Location  *string `json:"location"`
}
