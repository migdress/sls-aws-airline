package repository

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/meetupaws/flight_seat_reservation/flights/internal/model"
)

var (
	ErrNoFlightsFound      = errors.New("no_flights_found")
	ErrNoSeatFoundInFlight = errors.New("no_seats_found_in_the_given_flight")
	ErrSeatNotAvailable    = errors.New("seat_not_available")
)

type FlightsRepository struct {
	client *dynamodb.DynamoDB
	table  string
}

func (r *FlightsRepository) Save(m model.Flight) (model.Flight, error) {
	hasFreeSeats := 0
	if m.HasFreeSeats {
		hasFreeSeats = 1
	}

	seats := make([]*dynamodb.AttributeValue, len(m.Seats))
	for i, s := range m.Seats {
		seats[i] = &dynamodb.AttributeValue{
			M: map[string]*dynamodb.AttributeValue{
				"id": {
					S: aws.String(s.ID),
				},
				"letter": {
					S: aws.String(s.Letter),
				},
				"row": {
					N: aws.String(strconv.Itoa(s.Row)),
				},
				"passenger_id": {
					S: aws.String("-"),
				},
			},
		}
	}

	_, err := r.client.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String(r.table),
		Item: map[string]*dynamodb.AttributeValue{
			"id": {
				S: aws.String(m.ID),
			},
			"departure": {
				S: aws.String(m.Departure),
			},
			"has_free_seats": {
				N: aws.String(strconv.Itoa(hasFreeSeats)),
			},
			"seats": {
				L: seats,
			},
		},
	})

	if err != nil {
		return model.Flight{}, err
	}

	return m, nil
}

func (r *FlightsRepository) Find(id string) (model.Flight, error) {
	out, err := r.client.Query(&dynamodb.QueryInput{
		TableName: aws.String(r.table),
		KeyConditions: map[string]*dynamodb.Condition{
			"id": {
				ComparisonOperator: aws.String("EQ"),
				AttributeValueList: []*dynamodb.AttributeValue{
					{
						S: aws.String(id),
					},
				},
			},
		},
	})
	if err != nil {
		return model.Flight{}, err
	}

	if len(out.Items) == 0 {
		return model.Flight{}, ErrNoFlightsFound
	}

	flights, err := r.hydrate(out.Items)
	if err != nil {
		return model.Flight{}, err
	}
	return flights[0], nil
}

func (r *FlightsRepository) ListFlightsByDeparture(dateFrom string, dateTo string) ([]model.Flight, error) {
	out, err := r.client.Query(&dynamodb.QueryInput{
		TableName:              aws.String(r.table),
		IndexName:              aws.String("by_has_free_seats_and_departure"),
		KeyConditionExpression: aws.String("has_free_seats = :one AND departure BETWEEN :dateFrom AND :dateTo"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":one": {
				N: aws.String("1"),
			},
			":dateFrom": {
				S: aws.String(dateFrom),
			},
			":dateTo": {
				S: aws.String(dateTo),
			},
		},
	})
	if err != nil {
		return []model.Flight{}, err
	}

	if len(out.Items) == 0 {
		return []model.Flight{}, ErrNoFlightsFound
	}

	return r.hydrate(out.Items)
}

func (r *FlightsRepository) ReserveSeat(flightID string, seatID string, passengerID string) error {
	flight, err := r.Find(flightID)
	if err != nil {
		return err
	}

	foundSeat := model.FlightSeat{}
	foundSeatIndex := 0
	remainingSeats := 0
	for i, s := range flight.Seats {
		if foundSeat == (model.FlightSeat{}) && s.ID == seatID {
			foundSeat = s
			foundSeatIndex = i
		}
		if s.PassengerID == "" {
			remainingSeats++
		}
	}

	if foundSeat == (model.FlightSeat{}) {
		return ErrNoSeatFoundInFlight
	}

	if foundSeat.PassengerID != "" {
		return ErrSeatNotAvailable
	}

	updateExpression := aws.String(
		fmt.Sprintf("set seats[%v].passenger_id = :passengerID", foundSeatIndex),
	)
	expressionAttributeValues := map[string]*dynamodb.AttributeValue{
		":passengerID": {
			S: aws.String(passengerID),
		},
		":true": {
			N: aws.String("1"),
		},
	}
	if remainingSeats == 1 {
		updateExpression = aws.String(
			fmt.Sprintf("set has_free_seats = :zero, seats[%v].passenger_id = :passengerID", foundSeatIndex),
		)
		expressionAttributeValues[":zero"] = &dynamodb.AttributeValue{
			N: aws.String("0"),
		}
	}

	_, err = r.client.TransactWriteItems(&dynamodb.TransactWriteItemsInput{
		TransactItems: []*dynamodb.TransactWriteItem{
			{
				Update: &dynamodb.Update{
					TableName: aws.String(r.table),
					Key: map[string]*dynamodb.AttributeValue{
						"id": {
							S: aws.String(flightID),
						},
					},
					ConditionExpression:       aws.String("has_free_seats = :true"),
					UpdateExpression:          updateExpression,
					ExpressionAttributeValues: expressionAttributeValues,
				},
			},
		},
	})

	return err
}

func (r *FlightsRepository) hydrate(items []map[string]*dynamodb.AttributeValue) ([]model.Flight, error) {

	flights := make([]model.Flight, len(items))
	for i, item := range items {

		if v, ok := item["id"]; ok {
			flights[i].ID = *v.S
		}
		if v, ok := item["departure"]; ok {
			flights[i].Departure = *v.S
		}
		if v, ok := item["has_free_seats"]; ok {
			hasFreeSeats, err := strconv.ParseBool(*v.N)
			if err != nil {
				return []model.Flight{}, err
			}
			flights[i].HasFreeSeats = hasFreeSeats
		}

		if seatsList, ok := item["seats"]; ok {
			seats, err := r.hydrateSeats(seatsList.L)
			if err != nil {
				return []model.Flight{}, err
			}
			flights[i].Seats = seats
		}

	}
	return flights, nil

}

func (r *FlightsRepository) hydrateSeats(items []*dynamodb.AttributeValue) ([]model.FlightSeat, error) {

	seats := make([]model.FlightSeat, len(items))
	for i, item := range items {

		seatMap := item.M

		if v, ok := seatMap["id"]; ok {
			seats[i].ID = *v.S
		}
		if v, ok := seatMap["letter"]; ok {
			seats[i].Letter = *v.S
		}
		if v, ok := seatMap["passenger_id"]; ok && *v.S != "-" {
			seats[i].PassengerID = *v.S
		}
		if v, ok := seatMap["row"]; ok {
			intVal, err := strconv.Atoi(*v.N)
			if err != nil {
				return []model.FlightSeat{}, err
			}
			seats[i].Row = intVal
		}
	}

	return seats, nil
}

func NewFlightsRepository(client *dynamodb.DynamoDB, table string) *FlightsRepository {
	return &FlightsRepository{
		client: client,
		table:  table,
	}
}
