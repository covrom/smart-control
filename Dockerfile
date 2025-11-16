FROM golang:alpine AS builder
WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY internal .
COPY cmd .
RUN go build -a -o tgsmctl ./cmd/tgsmctl

FROM alpine:latest
RUN apk --no-cache add tzdata smartmontools
RUN mkdir -p /host/proc
RUN mkdir -p /var/lib/smart_reports_data
ENV TZ=Europe/Moscow
COPY --from=builder /app/tgsmctl /usr/local/bin/tgsmctl

ENTRYPOINT ["/usr/local/bin/tgsmctl"]