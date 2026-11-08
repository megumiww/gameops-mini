FROM golang:1.27-alpine AS build
WORKDIR /src
COPY go.mod ./
COPY . .
RUN go build -o /gameops-server ./cmd/server

FROM alpine:3.20
WORKDIR /app
COPY --from=build /gameops-server ./gameops-server
ENV GAMEOPS_ADDR=:8080
EXPOSE 8080
CMD ["./gameops-server"]
