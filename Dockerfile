FROM golang:1.25 AS builder
WORKDIR /app
COPY go.mod go.sum ./
#RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o icinga-dashboard .


# Stage 2: Create the final image
FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/icinga-dashboard .
COPY index.html/ ./
COPY favicon.ico/ ./
COPY assets/ ./assets/

ENV LISTEN_ADDRESS :8080
ENV ICINGA2_API_TIMEOUT 5
ENV MIN_STATE 1
ENV MAX_STATE 2
ENV MIN_STATE_TYPE 0


EXPOSE 8080
CMD ["./icinga-dashboard"]
