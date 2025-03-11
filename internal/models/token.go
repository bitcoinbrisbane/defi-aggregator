package models

// TokenRequest defines the structure for the token post request
type TokenRequest struct {
	Address string `json:"address" binding:"required"`
}