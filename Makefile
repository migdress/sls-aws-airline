.PHONY: build clean deploy remove

build: 
	make -C flights_list build

clean:
	make -C flights_list clean

deploy: clean build
	make -C flights_list deploy

remove: 
	make -C flights_list remove

