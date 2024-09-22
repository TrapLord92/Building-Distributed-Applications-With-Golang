package models

// API user credentials
// It is used to sign in
//
// swagger:model user

type User struct {
	// required: true
	Password string `json:"password"`
	// required: true
	Username string `json:"username"`
}
