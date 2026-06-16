# build stage
FROM golang:1.26-alpine AS build

WORKDIR /app
COPY . .

RUN go build -v -o ./log-stats .

# minimum image stage
FROM alpine:latest
COPY --from=build /app/log-stats .
ENTRYPOINT ["./log-stats"]

