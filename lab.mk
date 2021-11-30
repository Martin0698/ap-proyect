# build & run automation

APP_NAME=Pacman

build:
	go build -o ${APP_NAME}.o ${APP_NAME}.go

run: build
	@echo Running game
	./${APP_NAME}.o

clean:
	rm -rf *.o
