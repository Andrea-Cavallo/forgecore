FROM golang:1.24-alpine AS builder
WORKDIR /workspace
COPY shared/ shared/
COPY services/auth-service/ services/auth-service/
WORKDIR /workspace/services/auth-service
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -o /server ./cmd/server

FROM gcr.io/distroless/static-debian12
COPY --from=builder /server /server
EXPOSE 8081 9091
ENTRYPOINT ["/server"]
