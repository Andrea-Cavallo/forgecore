# ForgeCore Client Generation

ForgeCore mantiene due livelli di client interni:

- gRPC/proto: i contratti canonici vivono in `shared/proto/*.proto`.
- SDK Go HTTP: i client stabili vivono in `sdk/go/<bounded-context>`.

Regole:

- Ogni proto deve usare `package <context>.v1`.
- Ogni proto deve avere `option go_package` verso `shared/proto/<context>/v1`.
- I client Go devono usare `sdk/go/clientretry` e `sdk/go/clienttransport`.
- Non duplicare retry, circuit breaker, bearer injection o propagazione request id nei singoli client.
- I client generati o standardizzati devono compilare con `go build ./...` dentro `sdk/go`.

Comandi di verifica:

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\check-proto-contracts.ps1
cd sdk/go
go build ./...
```
