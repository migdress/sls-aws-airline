package main

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/google/go-cmp/cmp"
	"github.com/meetupaws/flight_seat_reservation/flights/internal/model"
	"github.com/meetupaws/flight_seat_reservation/flights/internal/repository"
	"github.com/meetupaws/flight_seat_reservation/internal"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type FlightsRepositoryMock struct {
	mock.Mock
}

func (m *FlightsRepositoryMock) ListFlightsByDeparture(dateFrom string, dateTo string) ([]model.Flight, error) {
	args := m.Called(dateFrom, dateTo)
	return args.Get(0).([]model.Flight), args.Error(1)
}

func TestAdapter(t *testing.T) {

	type mocks struct {
		flightsRepo *FlightsRepositoryMock
	}

	tests := []struct {
		name   string
		mocks  mocks
		req    events.APIGatewayProxyRequest
		want   events.APIGatewayProxyResponse
		mocker func(mocks mocks)
	}{
		{
			name: "Return a 200 status code after succesfully list flights by a given departure",
			req: events.APIGatewayProxyRequest{
				PathParameters: map[string]string{
					"dateFrom": "2019-11-25",
					"dateTo":   "2019-11-27",
				},
			},
			mocks: mocks{
				flightsRepo: &FlightsRepositoryMock{},
			},
			want: events.APIGatewayProxyResponse{
				StatusCode: 200,
				Headers: map[string]string{
					"Content-Type": "application/json",
				},
				Body: internal.TrimLines(`[
					{
						"id":"flight-1",
						"departure":"2019-11-26T09:25:00+0000",
						"has_free_seats": true,
						"seats":[
							{
								"id":"seat-1",
								"letter":"A",
								"row":1,
								"passenger_id":""
							},
							{
								"id":"seat-2",
								"letter":"B",
								"row":1,
								"passenger_id":"p1"
							}
						]
					},
					{
						"id":"flight-2",
						"departure":"2019-11-26T09:25:00+0000",
						"has_free_seats": true,
						"seats":[
							{
								"id":"seat-1",
								"letter":"B",
								"row":1,
								"passenger_id":""
							},
							{
								"id":"seat-2",
								"letter":"B",
								"row":2,
								"passenger_id":"p2"
							}
						]
					}
				]`),
			},
			mocker: func(m mocks) {
				m.flightsRepo.On(
					"ListFlightsByDeparture",
					"2019-11-25",
					"2019-11-27",
				).Return([]model.Flight{
					{
						ID:           "flight-1",
						Departure:    "2019-11-26T09:25:00+0000",
						HasFreeSeats: true,
						Seats: []model.FlightSeat{
							{
								ID:          "seat-1",
								Letter:      "A",
								Row:         1,
								PassengerID: "",
							},
							{
								ID:          "seat-2",
								Letter:      "B",
								Row:         1,
								PassengerID: "p1",
							},
						},
					},
					{
						ID:           "flight-2",
						Departure:    "2019-11-26T09:25:00+0000",
						HasFreeSeats: true,
						Seats: []model.FlightSeat{
							{
								ID:          "seat-1",
								Letter:      "B",
								Row:         1,
								PassengerID: "",
							},
							{
								ID:          "seat-2",
								Letter:      "B",
								Row:         2,
								PassengerID: "p2",
							},
						},
					},
				}, nil).Once()
			},
		}, {
			name: "Return a 500 status code after an error with the repository",
			req: events.APIGatewayProxyRequest{
				PathParameters: map[string]string{
					"dateFrom": "2019-11-25",
					"dateTo":   "2019-11-27",
				},
			},
			mocks: mocks{
				flightsRepo: &FlightsRepositoryMock{},
			},
			want: events.APIGatewayProxyResponse{
				StatusCode: 500,
				Headers: map[string]string{
					"Content-Type": "application/json",
				},
				Body: internal.TrimLines(`{
						"errors":["Some error"]
					}`),
			},
			mocker: func(m mocks) {
				m.flightsRepo.On(
					"ListFlightsByDeparture",
					"2019-11-25",
					"2019-11-27",
				).Return(
					[]model.Flight{},
					errors.New("Some error"),
				).Once()
			},
		},
		{
			name: "Return a 404 status code because there are not flights between the given dates",
			req: events.APIGatewayProxyRequest{
				PathParameters: map[string]string{
					"dateFrom": "2019-11-25",
					"dateTo":   "2019-11-27",
				},
			},
			mocks: mocks{
				flightsRepo: &FlightsRepositoryMock{},
			},
			want: events.APIGatewayProxyResponse{
				StatusCode: 404,
				Headers: map[string]string{
					"Content-Type": "application/json",
				},
				Body: internal.TrimLines(`{
						"errors":["no_flights_found"]
					}`),
			},
			mocker: func(m mocks) {
				m.flightsRepo.On(
					"ListFlightsByDeparture",
					"2019-11-25",
					"2019-11-27",
				).Return(
					[]model.Flight{},
					repository.ErrNoFlightsFound,
				).Once()
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			tt.mocker(tt.mocks)

			// Act
			handler := Adapter(tt.mocks.flightsRepo)
			got, err := handler(context.Background(), tt.req)

			// Assert
			require.NoError(t, err)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("Differences: (-want,+got)\n%s", diff)
			}

			tt.mocks.flightsRepo.AssertExpectations(t)
		})
	}

}
