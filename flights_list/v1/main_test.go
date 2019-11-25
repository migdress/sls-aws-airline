package main

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/google/go-cmp/cmp"
	common "github.com/meetupAWS20191126/try1/_common"
	"github.com/meetupAWS20191126/try1/_common/models"
	"github.com/stretchr/testify/mock"
)

type FlightsRepositoryMock struct {
	mock.Mock
}

func (m *FlightsRepositoryMock) ListFlightsByDeparture(dateFrom string, dateTo string) ([]models.Flight, error) {
	args := m.Called(dateFrom, dateTo)
	//return m.Called(dateFrom, dateTo).Get(0).([]models.Flight), m.Called.Error(1)
	return args.Get(0).([]models.Flight), args.Error(1)
}

func TestAdapter(t *testing.T) {

	type args struct {
		req events.APIGatewayProxyRequest
	}

	type mocks struct {
		flightsRepo *FlightsRepositoryMock
	}

	tests := []struct {
		name   string
		args   args
		mocks  mocks
		want   events.APIGatewayProxyResponse
		mocker func(mocks mocks, args args)
	}{
		{
			name: "Return a 200 status code after succesfully list flights by a given departure",
			args: args{
				req: events.APIGatewayProxyRequest{
					PathParameters: map[string]string{
						"date_from": "2019-11-25",
						"date_to":   "2019-11-27",
					},
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
				Body: common.TrimLines(`[
					{
						"id":"flight-1",
						"departure":"2019-11-26T09:25:00+0000",
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
			mocker: func(m mocks, a args) {
				m.flightsRepo.On(
					"ListFlightsByDeparture",
					"2019-11-25",
					"2019-11-27",
				).Return([]models.Flight{
					{
						ID:        "flight-1",
						Departure: "2019-11-26T09:25:00+0000",
						Seats: []models.FlightSeat{
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
						ID:        "flight-2",
						Departure: "2019-11-26T09:25:00+0000",
						Seats: []models.FlightSeat{
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
		},
		{
			name: "Return a 500 status code after an error with the repository",
			args: args{
				req: events.APIGatewayProxyRequest{
					PathParameters: map[string]string{
						"date_from": "2019-11-25",
						"date_to":   "2019-11-27",
					},
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
				Body: common.TrimLines(`{
					"error":"Some error"
				}`),
			},
			mocker: func(m mocks, a args) {
				m.flightsRepo.On(
					"ListFlightsByDeparture",
					"2019-11-25",
					"2019-11-27",
				).Return(
					[]models.Flight{},
					errors.New("Some error"),
				).Once()
			},
		},
		{
			name: "Return a 404 status code because there are not flights between the given dates",
			args: args{
				req: events.APIGatewayProxyRequest{
					PathParameters: map[string]string{
						"date_from": "2019-11-25",
						"date_to":   "2019-11-27",
					},
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
				Body: common.TrimLines(`{
					"error":"No flights found"
				}`),
			},
			mocker: func(m mocks, a args) {
				m.flightsRepo.On(
					"ListFlightsByDeparture",
					"2019-11-25",
					"2019-11-27",
				).Return(
					[]models.Flight{},
					nil,
				).Once()
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			tt.mocker(tt.mocks, tt.args)
			handler := Adapter(tt.mocks.flightsRepo)
			got, err := handler(context.Background(), tt.args.req)

			if err != nil {
				t.Errorf("Error in function Handler. %v\n", err)
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("Error in function Handler. (-want,+got)\n%s", diff)
			}
		})
	}

}
