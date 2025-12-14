FROM golang:1.18-bullseye AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/loadtest ./cmd/loadtest

FROM debian:bullseye-slim
WORKDIR /app
COPY --from=build /out/loadtest /app/loadtest

# Write outputs here (we'll mount a volume)
RUN mkdir -p /app/results

ENTRYPOINT ["/app/loadtest"]
