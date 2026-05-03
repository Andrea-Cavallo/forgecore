FROM golang:1.24-alpine AS builder
WORKDIR /workspace
COPY shared/ shared/
COPY services/job-service/ services/job-service/
WORKDIR /workspace/services/job-service
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -o /worker ./cmd/worker

FROM gcr.io/distroless/static-debian12
COPY --from=builder /worker /worker
ENTRYPOINT ["/worker"]
