package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
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

func (m *FlightsRepositoryMock) Find(id string) (model.Flight, error) {
	ret := m.Called(id)
	return ret.Get(0).(model.Flight), ret.Error(1)
}

func (m *FlightsRepositoryMock) ReserveSeat(flightID string, seatID string, passengerID string) error {
	ret := m.Called(flightID, seatID, passengerID)
	return ret.Error(0)
}

func TestAdapter(t *testing.T) {

	type mocks struct {
		flightsRepo *FlightsRepositoryMock
	}

	tests := []struct {
		name   string
		req    events.APIGatewayProxyRequest
		want   events.APIGatewayProxyResponse
		mocks  mocks
		mocker func(m mocks)
	}{
		{
			name: "Get a 200 status code after succesfully reserve a seat",
			req: events.APIGatewayProxyRequest{
				Body: `{
						"flight_id": "f1",
						"seat_id": "s1",
						"passenger_id": "p1"
					}`,
			},
			want: events.APIGatewayProxyResponse{
				StatusCode: http.StatusOK,
				Headers: map[string]string{
					"Content-Type": "application/json",
				},
			},
			mocks: mocks{
				flightsRepo: &FlightsRepositoryMock{},
			},
			mocker: func(m mocks) {
				m.flightsRepo.On(
					"Find",
					"f1",
				).Return(
					model.Flight{
						ID: "f1",
					},
					nil,
				).Once()

				m.flightsRepo.On(
					"ReserveSeat",
					"f1",
					"s1",
					"p1",
				).Return(nil).Once()
			},
		},
		{
			name: "Get a 400 status because request body is malformed",
			req: events.APIGatewayProxyRequest{
				Body: `{
						"flight_id": "f1",
						"seat_id": "s1",
						"passenger_id": "p1",
					}`,
			},
			want: events.APIGatewayProxyResponse{
				StatusCode: http.StatusBadRequest,
				Headers: map[string]string{
					"Content-Type": "application/json",
				},
				Body: internal.TrimLines(`{"errors":["invalid character '}' looking for beginning of object key string"]}`),
			},
			mocks: mocks{
				flightsRepo: &FlightsRepositoryMock{},
			},
			mocker: func(m mocks) {},
		},
		{
			name: "Get a 400 status because flight_id field is missing",
			req: events.APIGatewayProxyRequest{
				Body: `{
						"seat_id": "s1",
						"passenger_id": "p1"
					}`,
			},
			want: events.APIGatewayProxyResponse{
				StatusCode: http.StatusBadRequest,
				Headers: map[string]string{
					"Content-Type": "application/json",
				},
				Body: internal.TrimLines(`{"errors":["missing required fields"]}`),
			},
			mocks: mocks{
				flightsRepo: &FlightsRepositoryMock{},
			},
			mocker: func(m mocks) {},
		},
		{
			name: "Get a 404 status because the flight was not found",
			req: events.APIGatewayProxyRequest{
				Body: `{
					"flight_id": "f1",
					"seat_id": "s1",
					"passenger_id": "p1"
				}`,
			},
			want: events.APIGatewayProxyResponse{
				StatusCode: http.StatusNotFound,
				Headers: map[string]string{
					"Content-Type": "application/json",
				},
				Body: internal.TrimLines(
					fmt.Sprintf(`{"errors":["%s"]}`, repository.ErrNoFlightsFound),
				),
			},
			mocks: mocks{
				flightsRepo: &FlightsRepositoryMock{},
			},
			mocker: func(m mocks) {
				m.flightsRepo.On(
					"Find",
					"f1",
				).Return(
					model.Flight{},
					repository.ErrNoFlightsFound,
				).Once()
			},
		},
		{
			name: "Get a 500 status because the repo returned an unexpected error trying to find the flight",
			req: events.APIGatewayProxyRequest{
				Body: `{
					"flight_id": "f1",
					"seat_id": "s1",
					"passenger_id": "p1"
				}`,
			},
			want: events.APIGatewayProxyResponse{
				StatusCode: http.StatusInternalServerError,
				Headers: map[string]string{
					"Content-Type": "application/json",
				},
				Body: internal.TrimLines(`{"errors":["unexpected"]}`),
			},
			mocks: mocks{
				flightsRepo: &FlightsRepositoryMock{},
			},
			mocker: func(m mocks) {
				m.flightsRepo.On(
					"Find",
					"f1",
				).Return(
					model.Flight{},
					errors.New("unexpected"),
				).Once()
			},
		},
		{
			name: "Get a 500 status because the repository returned an unexpected error after trying to reserve a seat",
			req: events.APIGatewayProxyRequest{
				Body: `{
					"flight_id": "f1",
					"seat_id": "s1",
					"passenger_id": "p1"
				}`,
			},
			want: events.APIGatewayProxyResponse{
				StatusCode: http.StatusInternalServerError,
				Headers: map[string]string{
					"Content-Type": "application/json",
				},
				Body: internal.TrimLines(`{"errors":["unexpected_reserve"]}`),
			},
			mocks: mocks{
				flightsRepo: &FlightsRepositoryMock{},
			},
			mocker: func(m mocks) {
				m.flightsRepo.On(
					"Find",
					"f1",
				).Return(
					model.Flight{
						ID: "f1",
					},
					nil,
				).Once()

				m.flightsRepo.On(
					"ReserveSeat",
					"f1",
					"s1",
					"p1",
				).Return(errors.New("unexpected_reserve")).Once()
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
				t.Errorf("Differences found: (-want,+got)\n%s", diff)
			}
			tt.mocks.flightsRepo.AssertExpectations(t)
		})
	}

}
