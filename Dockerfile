FROM golang:1.14.0-alpine as builder
LABEL maintainer="alexander.knipping@noris.de"

ENV GOOS=linux
ENV GOARCH=amd64

RUN apk add --no-cache git

WORKDIR /app
COPY cmd/ cmd/
COPY *.go go.mod go.sum ./

RUN go mod download
RUN go build -o app ./cmd

# ---

FROM alpine:3.11
LABEL maintainer="alexander.knipping@noris.de"

COPY --from=builder /app/app /usr/bin/generate-ingestion

ENTRYPOINT ["/usr/bin/generate-ingestion"]