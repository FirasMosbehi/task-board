# Build stage
FROM golang:1.25.4 AS build

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o taskboard-backend

# Run stage
FROM alpine:3.20

WORKDIR /app

COPY --from=build /app/taskboard-backend .

ENV PORT=8080

EXPOSE 8080

CMD ["/app/taskboard-backend"]
