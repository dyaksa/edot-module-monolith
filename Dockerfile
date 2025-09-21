FROM golang:1.24-alpine
# Install Goose CLI
USER root
# This binary will be copied to the final stage.
RUN go install github.com/pressly/goose/v3/cmd/goose@latest
# Install dependencies
RUN apk add --no-cache build-base

#Set Working Directory
WORKDIR /usr/src/app

COPY . .

RUN mkdir "tmp"

# Build Go
RUN CGO_ENABLED=1 go build -tags musl -o main main.go

# Expose Application Port
EXPOSE 8080

# Run The Application
CMD ["./main"]
