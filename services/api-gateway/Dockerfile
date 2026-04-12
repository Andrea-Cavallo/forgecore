FROM golang:1.24-alpine AS builder
WORKDIR /workspace
COPY shared/ shared/
COPY services/api-gateway/ services/api-gateway/
WORKDIR /workspace/services/api-gateway
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -o /server ./cmd/server

FROM gcr.io/distroless/static-debian12
COPY --from=builder /server /server
EXPOSE 8080
ENTRYPOINT ["/server"]
