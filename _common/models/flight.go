package models

type Flight struct {
	ID        string       `json:"id"`
	Departure string       `json:"departure"`
	Seats     []FlightSeat `json:"seats"`
}

type FlightSeat struct {
	ID          string `json:"id"`
	Letter      string `json:"letter"`
	PassengerID string `json:"passenger_id"`
	Row         int    `json:"row"`
}
