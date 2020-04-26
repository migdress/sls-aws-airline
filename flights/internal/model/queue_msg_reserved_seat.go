package model

type QueueMsgReservedSeat struct {
	FlightID        string `json:"flight_id"`
	FlightDeparture string `json:"flight_departure"`
	SeatLetter      string `json:"seat_letter"`
	SeatRow         int    `json:"seat_row"`
	UserID          string `json:"user_id"`
}
