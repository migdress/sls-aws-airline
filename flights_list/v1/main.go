package main

import (
	"context"
	"errors"

	"github.com/aws/aws-lambda-go/events"
	common "github.com/meetupAWS20191126/try1/_common"
	"github.com/meetupAWS20191126/try1/_common/models"
)

type Handler func(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error)

type ResponseBody []ResponseBodyFlight

type ResponseBodyFlight struct {
	ID        string                   `json:"id"`
	Departure string                   `json:"departure"`
	Seats     []ResponseBodyFlightSeat `json:"seats"`
}

type ResponseBodyFlightSeat struct {
	ID          string `json:"id"`
	Letter      string `json:"letter"`
	Row         int    `json:"row"`
	PassengerID string `json:"passenger_id"`
}

type FlightsRepositoryInterface interface {
	ListFlightsByDeparture(dateFron string, dateTo string) ([]models.Flight, error)
}

func Adapter(flightsRepo FlightsRepositoryInterface) Handler {
	return func(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

		dateFrom := req.PathParameters["date_from"]
		dateTo := req.PathParameters["date_to"]
		flights, err := flightsRepo.ListFlightsByDeparture(
			dateFrom,
			dateTo,
		)
		if err != nil {
			return common.JSONError(500, err), nil
		}

		if len(flights) == 0 {
			return common.JSONError(404, errors.New("No flights found")), nil
		}

		response := make(ResponseBody, len(flights))
		for i, f := range flights {
			rSeats := make([]ResponseBodyFlightSeat, len(f.Seats))
			for j, s := range f.Seats {
				rSeat := ResponseBodyFlightSeat{}
				rSeat.ID = s.ID
				rSeat.Letter = s.Letter
				rSeat.Row = s.Row
				rSeat.PassengerID = s.PassengerID
				rSeats[j] = rSeat
			}
			rFlight := ResponseBodyFlight{}
			rFlight.ID = f.ID
			rFlight.Departure = f.Departure
			rFlight.Seats = rSeats
			response[i] = rFlight
		}

		return common.JSONStruct(200, response), nil
	}
}

/*
func main() {
	lambda.Start(Adapter())
}
*/
