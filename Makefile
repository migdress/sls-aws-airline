.PHONY: deploy_flights remove_flights

deploy_flights: 
	make -C flights/list deploy
	make -C flights/reserve_seat deploy
	make -C flights/send_reservation_email deploy

remove_flights: 
	make -C flights/list remove
	make -C flights/reserve_seat remove
	make -C flights/send_reservation_email remove

