# Dockerfile

# Use the official image as a parent image
FROM golang:latest

# create a new directory for our app
WORKDIR /app

# copy go mod and sum files
COPY . .

#build golang app
RUN go build -o main ./cmd/main.go

# expose port 8080
EXPOSE 8080

# run the app
CMD ["./main"]
