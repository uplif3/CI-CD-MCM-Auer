# Docker Analysis – Product Catalog API

## Part A – Dockerfile Analysis

### Multi-Stage Build

### Stage 1 – `builder` (`golang:1.26-alpine`)

| Schritt | Befehl | Zweck |
|---------|--------|-------|
| 1 | `COPY go.mod go.sum ./` | Nur Dependency-Dateien kopieren (Layer-Cache) |
| 2 | `RUN go mod download` | Abhängigkeiten herunterladen und cachen |
| 3 | `COPY . .` | Restlichen Source-Code kopieren |
| 4 | `RUN CGO_ENABLED=0 GOOS=linux go build -o /api-server ./cmd/api` | Statisch gelinktes Binary kompilieren |

**Warum zuerst nur `go.mod`/`go.sum` kopieren?**  
Docker cached jeden Layer. Ändert sich nur Source-Code (nicht die Dependencies), wird `go mod download` aus dem Cache geladen — der langsamste Schritt wird übersprungen.

### Stage 2 – Runtime (`alpine:3.19`)

Nur zwei Aktionen:
1. `COPY --from=builder /api-server .` → kopiert das fertige Binary aus Stage 1
2. `ENTRYPOINT ["./api-server"]` → startet den Server

Das finale Image enthält **keinen Go-Compiler**, keine Quellfiles, keine Build-Tools.

### Was macht `CGO_ENABLED=0`?

CGO ist die Schnittstelle zwischen Go und C-Bibliotheken. Wenn CGO aktiviert ist, linkt Go das Binary **dynamisch** gegen glibc. Alpine Linux verwendet jedoch **musl** statt glibc — das würde zur Laufzeit crashen.

Mit `CGO_ENABLED=0` wird Go gezwungen, ein **statisch gelinktes** Binary zu erzeugen: alle Abhängigkeiten sind eingebettet, keine externe C-Bibliothek wird zur Laufzeit benötigt. Das Binary läuft in jedem Linux-Container, unabhängig von der C-Runtime.

### Image-Größe: Multi-Stage vs. Single-Stage

| Build-Ansatz | Basis-Image | Enthält | Größe (ca.) |
|---|---|---|---|
| Single-stage | `golang:1.26-alpine` | Go-Toolchain + Source + Binary | **509 MB** |
| Multi-stage (Stage 2) | `alpine:3.19` | Binary + ca-certificates | **29.3 MB** |

**Faktor ~17× kleiner**

### Docker Compose Setup

`docker-compose.yml` startet zwei Services:

| Service | Image | Port | Zweck |
|---------|-------|------|-------|
| `db` | `postgres:16-alpine` | 5432 | PostgreSQL Datenbank |
| `api` | Build aus `Dockerfile` | 8080 | Product Catalog REST API |

**Persistenz:** Volume `pgdata` mappt `/var/lib/postgresql/data` → Daten überleben `docker compose down`.

**Health Check:** `api` startet erst wenn `db` den Health-Check (`pg_isready`) besteht (`depends_on: condition: service_healthy`).

---

## Part B – CRUD-Tests mit Docker Compose

### Start

```bash
docker compose up --build
```

Output (gekürzt):
```
[+] Building 1.4s (18/18) FINISHED
 => CACHED [builder 4/6] RUN go mod download
 => CACHED [builder 6/6] RUN CGO_ENABLED=0 GOOS=linux go build -o /api-server ./cmd/api
 => CACHED [stage-1 4/4] COPY --from=builder /api-server .
[+] Running 4/4
 ✔ Network ci-cd-mcm-auer_default  Created
 ✔ Container ci-cd-mcm-auer-db-1   Healthy
 ✔ Container ci-cd-mcm-auer-api-1  Started
```

**Tatsächliche Image-Größe:** `29.3 MB` (vs. ~500 MB single-stage)

---

### Create – 3 Produkte anlegen

```bash
curl -s -X POST http://localhost:8080/products \
  -H "Content-Type: application/json" \
  -d '{"name":"Laptop","price":999.99}'
```
```json
{"id":1,"name":"Laptop","price":999.99}
```

```bash
curl -s -X POST http://localhost:8080/products \
  -H "Content-Type: application/json" \
  -d '{"name":"Mouse","price":29.99}'
```
```json
{"id":2,"name":"Mouse","price":29.99}
```

```bash
curl -s -X POST http://localhost:8080/products \
  -H "Content-Type: application/json" \
  -d '{"name":"Keyboard","price":79.99}'
```
```json
{"id":3,"name":"Keyboard","price":79.99}
```

---

### Read – Alle Produkte auflisten

```bash
curl -s http://localhost:8080/products
```
```json
[{"id":1,"name":"Laptop","price":999.99},{"id":2,"name":"Mouse","price":29.99},{"id":3,"name":"Keyboard","price":79.99}]
```

### Read – Einzelnes Produkt abrufen

```bash
curl -s http://localhost:8080/products/1
```
```json
{"id":1,"name":"Laptop","price":999.99}
```

---

### Update – Produkt aktualisieren

```bash
curl -s -X PUT http://localhost:8080/products/1 \
  -H "Content-Type: application/json" \
  -d '{"name":"Gaming Laptop","price":1499.99}'
```
```json
{"id":1,"name":"Gaming Laptop","price":1499.99}
```

---

### Delete – Produkt löschen und 404 prüfen

```bash
curl -s -X DELETE http://localhost:8080/products/3
```
```json
{"result":"success"}
```

```bash
curl -s http://localhost:8080/products/3
```
```json
{"error":"Product not found"}
```

---

### Persistenz-Test

```bash
docker compose down
docker compose up
curl -s http://localhost:8080/products
```
```json
[{"id":1,"name":"Gaming Laptop","price":1499.99},{"id":2,"name":"Mouse","price":29.99}]
```

Die Daten sind nach dem Neustart noch vorhanden — `id:3` (Keyboard) wurde vorher gelöscht, `id:1` zeigt den aktualisierten Namen "Gaming Laptop".  
**Grund:** Das Volume `pgdata` persistiert `/var/lib/postgresql/data` über Container-Neustarts hinweg.
