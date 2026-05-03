FROM golang:1.26-alpine AS builder
WORKDIR /workspace
COPY shared/ shared/
COPY services/forgecore-payments/ services/forgecore-payments/
WORKDIR /workspace/services/forgecore-payments
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -o /server ./cmd/server

FROM gcr.io/distroless/static-debian12
COPY --from=builder /server /server
EXPOSE 8082 9092
ENTRYPOINT ["/server"]
