package model

type Flight struct {
	ID           string       `json:"id"`
	Departure    string       `json:"departure"`
	HasFreeSeats bool         `json:"has_free_seats"`
	Seats        []FlightSeat `json:"seats"`
}

type FlightSeat struct {
	ID          string `json:"id"`
	Letter      string `json:"letter"`
	PassengerID string `json:"passenger_id"`
	Row         int    `json:"row"`
}
