# sls-aws-airline
This is a serverless project

# architecture

![arch](https://github.com/migdress/sls-aws-airline-docs/blob/master/default/guides/assets/static/flight-seat-reservation/arch.png?raw=true)

## Flights

A subdoman with 3 microservices
  * **list**: list the flight by departure given a range of dates
    * The `passenger_id` is an email
  * **reserve_seat**: reserves a seat in a flight
  * **send_email**: sends an email to the user confirming the reservation
