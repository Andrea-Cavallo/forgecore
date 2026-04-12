FROM golang:1.24-alpine AS builder
WORKDIR /workspace
COPY shared/ shared/
COPY services/permission-service/ services/permission-service/
WORKDIR /workspace/services/permission-service
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -o /server ./cmd/server

FROM gcr.io/distroless/static-debian12
COPY --from=builder /server /server
EXPOSE 8087 9097
ENTRYPOINT ["/server"]
