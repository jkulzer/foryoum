run:
	go fmt .
	templ fmt *
	templ generate
	go run .
