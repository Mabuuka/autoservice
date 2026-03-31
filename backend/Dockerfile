FROM golang:1.26.1-alpine3.23 AS builder

WORKDIR /app/backend
COPY backend /app/backend

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/app main.go

FROM alpine:3.23

WORKDIR /app

RUN apk add --no-cache ca-certificates \
    && addgroup -S app \
    && adduser -S app -G app

COPY --from=builder /app/app /app/app
COPY frontend /app/frontend

ENV APP_PORT=8080
ENV FRONTEND_DIR=/app/frontend

EXPOSE 8080

USER app

CMD ["/app/app"]