package models

type Error struct {
	StatusCode int    `json:"statusCode" example:"500"`
	Error      string `json:"error" example:"Internal Server Error."`
}

type Message struct {
	StatusCode int    `json:"statusCode" example:"200"`
	Message    string `json:"message" example:"message"`
}
