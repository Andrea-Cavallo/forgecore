FROM golang:1.26-alpine AS builder
WORKDIR /workspace
COPY shared/ shared/
COPY services/forgecore-permissions/ services/forgecore-permissions/
WORKDIR /workspace/services/forgecore-permissions
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -o /server ./cmd/server

FROM gcr.io/distroless/static-debian12
COPY --from=builder /server /server
EXPOSE 8087 9097
ENTRYPOINT ["/server"]
