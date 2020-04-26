package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/meetupaws/flight_seat_reservation/flights/internal/model"
	"github.com/meetupaws/flight_seat_reservation/flights/internal/repository"
	"github.com/meetupaws/flight_seat_reservation/internal"
)

type Handler func(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error)

type FlightsRepository interface {
	Find(id string) (model.Flight, error)
	ReserveSeat(flightID string, seatID string, passengerID string) error
}

type Enqueuer interface {
	SendMsg(msg interface{}, queue string) error
}

type Request struct {
	FlightID    string `json:"flight_id"`
	SeatID      string `json:"seat_id"`
	PassengerID string `json:"passenger_id"`
}

func Adapter(flightsRepo FlightsRepository, enqueuer Enqueuer, notificationsQueue string) Handler {
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

		// Send message to queue
		seat := getSeat(flight, request.SeatID)
		err = enqueuer.SendMsg(
			model.QueueMsgReservedSeat{
				FlightID:        flight.ID,
				FlightDeparture: flight.Departure,
				SeatLetter:      seat.Letter,
				SeatRow:         seat.Row,
				UserID:          request.PassengerID,
			},
			notificationsQueue,
		)
		if err != nil {
			log.Printf("An error ocurred while sending message to queue %v: %v", notificationsQueue, err)
		}

		return internal.Respond(http.StatusOK, ""), nil
	}
}

func getSeat(flight model.Flight, seatID string) model.FlightSeat {
	for _, s := range flight.Seats {
		if s.ID == seatID {
			return s
		}
	}
	return model.FlightSeat{}
}

func main() {
	flightsTable := os.Getenv("DYNAMODB_FLIGHTS")
	if internal.TrimLines(flightsTable) == "" {
		panic("DYNAMODB_FLIGHTS is empty")
	}
	notificationsQueue := os.Getenv("NOTIFICATIONS_QUEUE")
	if internal.TrimLines(notificationsQueue) == "" {
		panic("NOTIFICATIONS_QUEUE is empty")
	}
	session := session.New()
	dynamodbClient := dynamodb.New(session)
	flightsRepo := repository.NewFlightsRepository(dynamodbClient, flightsTable)
	sqsClient := sqs.New(session)
	enqueuer := internal.NewEnqueuer(sqsClient)
	lambda.Start(Adapter(flightsRepo, enqueuer, notificationsQueue))
}
