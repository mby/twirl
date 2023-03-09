.PHONY: help  # list targets
help:
	@grep '^.PHONY: ' Makefile | sed 's/.PHONY: /  /g' | awk '{split($$0,a," # "); print a[1],"\t",a[2]}'

.PHONY: run   # listening @:8080
run: deps swag
	go run .

.PHONY: swag  # update swagger
swag:
	swag init

.PHONY: deps  # installs dependencies
deps:
	go install github.com/swaggo/swag/cmd/swag@latest
