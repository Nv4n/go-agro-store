package server

type UserRegister struct {
	Email     string `json:"email" form:"email" validate:"required,email"`
	Password  string `json:"password" form:"password" validate:"required,min=8,max=32"`
	FirstName string `json:"first_name" form:"fname" validate:"required,alpha,min=2,max=50"`
	LastName  string `json:"last_name" form:"lname" validate:"required,alpha,min=2,max=50"`
}

type UserLogin struct {
	Email    string `json:"email" form:"email" validate:"required,email"`
	Password string `json:"password" form:"password" validate:"required,min=8,max=32"`
}
