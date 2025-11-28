FROM golang:1.25.4 AS builder

WORKDIR /app

COPY src/* ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o grafana-demo-app

FROM gcr.io/distroless/base-debian12

WORKDIR /
COPY --from=builder /app/grafana-demo-app /grafana-demo-app

EXPOSE 8080
ENTRYPOINT ["/grafana-demo-app"]
