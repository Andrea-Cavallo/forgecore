FROM golang:1.24-alpine AS builder
WORKDIR /workspace
COPY shared/ shared/
COPY services/notification-service/ services/notification-service/
WORKDIR /workspace/services/notification-service
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -o /server ./cmd/server

FROM gcr.io/distroless/static-debian12
COPY --from=builder /server /server
EXPOSE 8083
ENTRYPOINT ["/server"]
