package main

import (
	"context"
	"encoding/json"
	"errors"
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

type FlightsRepository interface {
	Find(id string) (model.Flight, error)
	ReserveSeat(flightID string, seatID string, passengerID string) error
}

type Request struct {
	FlightID    string `json:"flight_id"`
	SeatID      string `json:"seat_id"`
	PassengerID string `json:"passenger_id"`
}

func Adapter(flightsRepo FlightsRepository) Handler {
	return func(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		request := Request{}
		err := json.Unmarshal([]byte(req.Body), &request)
		if err != nil {
			return internal.Error(http.StatusBadRequest, err), nil
		}

		// Validations
		if internal.TrimLines(request.FlightID) == "" ||
			internal.TrimLines(request.SeatID) == "" ||
			internal.TrimLines(request.PassengerID) == "" {
			return internal.Error(http.StatusBadRequest, errors.New("missing required fields")), nil
		}

		// Find the flight
		flight, err := flightsRepo.Find(request.FlightID)
		if err == repository.ErrNoFlightsFound {
			return internal.Error(http.StatusNotFound, err), nil
		}
		if err != nil {
			return internal.Error(http.StatusInternalServerError, err), nil
		}

		// Reserve seat
		err = flightsRepo.ReserveSeat(flight.ID, request.SeatID, request.PassengerID)
		if err == repository.ErrNoSeatFoundInFlight {
			return internal.Error(http.StatusNotFound, err), nil
		}
		if err == repository.ErrSeatNotAvailable {
			return internal.Error(http.StatusUnprocessableEntity, err), nil
		}
		if err != nil {
			return internal.Error(http.StatusInternalServerError, err), nil
		}

		return internal.Respond(http.StatusOK, ""), nil
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
