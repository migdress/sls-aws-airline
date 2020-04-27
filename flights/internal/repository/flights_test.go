package repository

import (
	"fmt"
	"log"
	"sync"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/google/go-cmp/cmp"
	"github.com/meetupaws/flight_seat_reservation/flights/internal/model"
	"github.com/meetupaws/flight_seat_reservation/internal"
	"github.com/stretchr/testify/require"
)

func createFlightsTable(client *dynamodb.DynamoDB, table string, t *testing.T) {
	_, err := client.CreateTable(&dynamodb.CreateTableInput{
		TableName: aws.String(table),
		AttributeDefinitions: []*dynamodb.AttributeDefinition{
			{
				AttributeName: aws.String("id"),
				AttributeType: aws.String("S"),
			},
			{
				AttributeName: aws.String("has_free_seats"),
				AttributeType: aws.String("N"),
			},
			{
				AttributeName: aws.String("departure"),
				AttributeType: aws.String("S"),
			},
		},
		KeySchema: []*dynamodb.KeySchemaElement{
			{
				AttributeName: aws.String("id"),
				KeyType:       aws.String("HASH"),
			},
		},
		ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(5),
			WriteCapacityUnits: aws.Int64(5),
		},
		GlobalSecondaryIndexes: []*dynamodb.GlobalSecondaryIndex{
			{
				IndexName: aws.String("by_has_free_seats_and_departure"),
				KeySchema: []*dynamodb.KeySchemaElement{
					{
						AttributeName: aws.String("has_free_seats"),
						KeyType:       aws.String("HASH"),
					},
					{
						AttributeName: aws.String("departure"),
						KeyType:       aws.String("RANGE"),
					},
				},
				Projection: &dynamodb.Projection{
					ProjectionType: aws.String("ALL"),
				},
				ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
					ReadCapacityUnits:  aws.Int64(5),
					WriteCapacityUnits: aws.Int64(5),
				},
			},
		},
	})
	if err != nil {
		t.Errorf("Error while creating flights table: %v\n", err)
	}
}

func TestFlightsRepository_SaveAndFind(t *testing.T) {

	table := "flights"
	closer, client := internal.DynamodbStart(t)
	defer closer()
	createFlightsTable(client, table, t)
	flightsRepo := NewFlightsRepository(client, table)

	flightsToSave := []model.Flight{
		{
			ID:        "f2",
			Departure: "2019-11-26T09:05:00+0000",
			Seats: []model.FlightSeat{
				{
					ID:     "s1",
					Letter: "A",
					Row:    1,
				},
				{
					ID:     "s1",
					Letter: "A",
					Row:    1,
				},
			},
		},
		{
			ID:        "f2",
			Departure: "2019-11-26T09:05:00+0000",
			Seats: []model.FlightSeat{
				{
					ID:     "s1",
					Letter: "A",
					Row:    1,
				},
				{
					ID:     "s1",
					Letter: "A",
					Row:    1,
				},
			},
		},
	}

	for _, f := range flightsToSave {
		_, err := flightsRepo.Save(f)
		if err != nil {
			t.Errorf("Error while saving flight: %v", err)
		}
	}

	for i, f := range flightsToSave {
		foundFlight, err := flightsRepo.Find(f.ID)
		if err != nil {
			t.Errorf("Error while finding flight: %v\n", err)
		}
		if diff := cmp.Diff(flightsToSave[i], foundFlight); diff != "" {
			t.Errorf("Error while finding flight: (-want,+got)\n%s", diff)
		}
	}

}

func TestFlightsRepository_ListFlightsByDeparture(t *testing.T) {

	// Arrange
	table := "flights"
	closer, client := internal.DynamodbStart(t)
	defer closer()
	createFlightsTable(client, table, t)
	flightsRepo := NewFlightsRepository(client, table)

	flightsToSave := []model.Flight{
		{
			ID:           "f2",
			Departure:    "2019-11-26T09:05:00+0000",
			HasFreeSeats: true,
			Seats: []model.FlightSeat{
				{
					ID:     "s1",
					Letter: "A",
					Row:    1,
				},
				{
					ID:     "s1",
					Letter: "A",
					Row:    1,
				},
			},
		},
		{
			ID:           "f2",
			Departure:    "2019-11-22T09:05:00+0000",
			HasFreeSeats: true,
			Seats: []model.FlightSeat{
				{
					ID:     "s1",
					Letter: "A",
					Row:    1,
				},
				{
					ID:     "s1",
					Letter: "A",
					Row:    1,
				},
			},
		},
		{
			ID:           "f3",
			Departure:    "2019-11-24T09:05:00+0000",
			HasFreeSeats: true,
			Seats: []model.FlightSeat{
				{
					ID:     "s1",
					Letter: "A",
					Row:    1,
				},
				{
					ID:     "s1",
					Letter: "A",
					Row:    1,
				},
			},
		},
	}

	for _, f := range flightsToSave {
		_, err := flightsRepo.Save(f)
		require.NoError(t, err)
	}

	// Act
	foundFlights, err := flightsRepo.ListFlightsByDeparture("2019-11-21T00:00:00+0000", "2019-11-25T00:00:00+0000")
	require.NoError(t, err)
	require.Len(t, foundFlights, 2)
	require.Contains(t, foundFlights, flightsToSave[1])
	require.Contains(t, foundFlights, flightsToSave[2])
}

func TestFlightsRepository_ReserveSeat(t *testing.T) {
	// Arrange
	table := "flights"
	closer, client := internal.DynamodbStart(t)
	defer closer()
	createFlightsTable(client, table, t)
	flightsRepo := NewFlightsRepository(client, table)

	// Save a flight so we can try to reserve a seat on it
	flightToSave := model.Flight{
		ID:           "f2",
		Departure:    "2019-11-26T09:05:00+0000",
		HasFreeSeats: true,
		Seats: []model.FlightSeat{
			{
				ID:     "s1",
				Letter: "A",
				Row:    1,
			},
			{
				ID:     "s2",
				Letter: "A",
				Row:    1,
			},
		},
	}
	_, err := flightsRepo.Save(flightToSave)
	if err != nil {
		t.Errorf("Error while saving flight: %v", err)
	}

	// Act, Concurrently try to reserve a seat
	limit := 100
	wg := sync.WaitGroup{}
	wg.Add(limit)
	mux := sync.Mutex{}
	success := 0
	passengerWinner := ""
	launchTime := time.Now()
	for i := 0; i < limit; i++ {
		go func(ii int) {
			time.Sleep(time.Until(launchTime))
			log.Printf("Launching test [%v] at %v\n", ii, time.Now())
			mux.Lock()
			err := flightsRepo.ReserveSeat("f2", "s1", fmt.Sprintf("%v", ii))
			if err != nil {
				log.Printf("[%v] unable to reserve seat\n", ii)
			} else {
				log.Printf("[%v] GOT THE SEAT!", ii)
				passengerWinner = fmt.Sprintf("%v", ii)
				success++
			}
			mux.Unlock()
			wg.Done()
		}(i)
	}
	wg.Wait()

	// Assert
	require.Equal(t, 1, success, "A seat was reserved more than once")
	updatedFlight, err := flightsRepo.Find("f2")
	require.NoError(t, err)
	require.Contains(t, updatedFlight.Seats, model.FlightSeat{
		ID:          "s1",
		Letter:      "A",
		Row:         1,
		PassengerID: passengerWinner,
	})

}
