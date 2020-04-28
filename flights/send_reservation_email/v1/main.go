package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/meetupaws/flight_seat_reservation/flights/internal/model"
	"github.com/meetupaws/flight_seat_reservation/internal"
)

type Handler func(ctx context.Context, event events.SQSEvent) error

type FlightsRepository interface {
	Find(id string) (model.Flight, error)
	ReserveSeat(flightID string, seatID string, passengerID string) error
}

type Mailer interface {
	SendEmail(subject string, body string, from string, to []string, cc []string) error
}

var emailTemplate = `
Hello! %v.
Your resevartion is confirmed, seat %v%v for the fly with id %v on on %v!
`

type Request struct {
	FlightID    string `json:"flight_id"`
	SeatID      string `json:"seat_id"`
	PassengerID string `json:"passenger_id"`
}

func Adapter(mailer Mailer, senderEmail string) Handler {
	return func(ctx context.Context, event events.SQSEvent) error {
		msgBody := model.QueueMsgReservedSeat{}
		err := json.Unmarshal([]byte(event.Records[0].Body), &msgBody)
		if err != nil {
			return err
		}

		emailBody := fmt.Sprintf(
			msgBody.UserID,
			msgBody.SeatRow,
			msgBody.SeatLetter,
			msgBody.FlightID,
			msgBody.FlightDeparture,
		)

		err = mailer.SendEmail(
			"Flight seat reservation",
			emailBody,
			senderEmail,
			[]string{msgBody.UserID},
			nil,
		)
		return err
	}
}

func main() {
	senderEmail := os.Getenv("SENDER_EMAIL")
	if internal.TrimLines(senderEmail) == "" {
		panic("SENDER_EMAIL is empty")
	}
	session := session.New()
	sesClient := ses.New(session)
	mailer := internal.NewMailer(sesClient)
	lambda.Start(Adapter(mailer, senderEmail))
}
