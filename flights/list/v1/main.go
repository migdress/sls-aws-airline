package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/meetupaws/flight_seat_reservation/flights/internal/model"
	"github.com/meetupaws/flight_seat_reservation/flights/internal/repository"
	"github.com/meetupaws/flight_seat_reservation/internal"
)

type Handler func(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error)

type Response []ResponseFlight

type ResponseFlight struct {
	ID           string               `json:"id"`
	Departure    string               `json:"departure"`
	HasFreeSeats bool                 `json:"has_free_seats"`
	Seats        []ResponseFlightSeat `json:"seats"`
}

type ResponseFlightSeat struct {
	ID          string `json:"id"`
	Letter      string `json:"letter"`
	Row         int    `json:"row"`
	PassengerID string `json:"passenger_id"`
}

type FlightsRepository interface {
	ListFlightsByDeparture(dateFrom string, dateTo string) ([]model.Flight, error)
}

func Adapter(flightsRepo FlightsRepository) Handler {
	return func(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		// Get request parameters
		dateFrom := req.PathParameters["dateFrom"]
		dateTo := req.PathParameters["dateTo"]

		// Look for flights
		flights, err := flightsRepo.ListFlightsByDeparture(
			dateFrom,
			dateTo,
		)
		if err == repository.ErrNoFlightsFound {
			return internal.Error(http.StatusNotFound, err), nil
		}
		if err != nil {
			return internal.Error(http.StatusInternalServerError, err), nil
		}

		// Prepare response
		response := make(Response, len(flights))
		for i, f := range flights {
			rSeats := make([]ResponseFlightSeat, len(f.Seats))
			for j, s := range f.Seats {
				rSeat := ResponseFlightSeat{}
				rSeat.ID = s.ID
				rSeat.Letter = s.Letter
				rSeat.Row = s.Row
				rSeat.PassengerID = s.PassengerID
				rSeats[j] = rSeat
			}
			rFlight := ResponseFlight{}
			rFlight.ID = f.ID
			rFlight.Departure = f.Departure
			rFlight.HasFreeSeats = f.HasFreeSeats
			rFlight.Seats = rSeats
			response[i] = rFlight
		}

		// Respond
		responseBytes, _ := json.Marshal(response)
		return internal.Respond(200, string(responseBytes)), nil
	}
}

func main() {
	flightsTable := os.Getenv("DYNAMODB_FLIGHTS")
	if internal.TrimLines(flightsTable) == "" {
		panic("DYNAMODB_FLIGHTS is empty")
	}
	session := session.New()
	dynamodbClient := dynamodb.New(session)
	flightsRepo := repository.NewFlightsRepository(dynamodbClient, flightsTable)
	lambda.Start(Adapter(flightsRepo))
}
