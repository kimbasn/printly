package dto

type Credentials struct {
	Phone_number string `form:"phone_number"`
	Password     string `form:"password"`
}

type JWT struct {
	Token string `json:"token"`
}
