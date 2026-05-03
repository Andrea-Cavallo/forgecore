FROM golang:1.26-alpine AS builder
WORKDIR /workspace
COPY shared/ shared/
COPY services/forgecore-jobs/ services/forgecore-jobs/
WORKDIR /workspace/services/forgecore-jobs
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -o /worker ./cmd/worker

FROM gcr.io/distroless/static-debian12
COPY --from=builder /worker /worker
EXPOSE 8092
ENTRYPOINT ["/worker"]
