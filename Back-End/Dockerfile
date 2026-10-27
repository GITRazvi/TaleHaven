FROM golang:1.23-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/talehaven-api main.go

FROM alpine:3.21

WORKDIR /app
RUN addgroup -S talehaven && adduser -S talehaven -G talehaven && mkdir images && chown talehaven:talehaven images

COPY --from=build /out/talehaven-api ./talehaven-api

USER talehaven
EXPOSE 3001

CMD ["./talehaven-api"]
