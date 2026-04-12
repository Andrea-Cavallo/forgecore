FROM golang:1.24-alpine AS builder
WORKDIR /workspace
COPY shared/ shared/
COPY services/webhook-service/ services/webhook-service/
WORKDIR /workspace/services/webhook-service
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -o /server ./cmd/server

FROM gcr.io/distroless/static-debian12
COPY --from=builder /server /server
EXPOSE 8089
ENTRYPOINT ["/server"]
