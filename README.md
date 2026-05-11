## One Piece API 🏴‍☠️

> Manage One Piece characters with a clean, Firestore‑backed Go REST API.

[![Go Version](https://img.shields.io/badge/Go-%3E%3D1.25-blue?logo=go)](https://golang.org/)
[![Firestore](https://img.shields.io/badge/Backend-Firestore-orange?logo=firebase)](https://firebase.google.com/docs/firestore)
[![Status](https://img.shields.io/badge/Status-Active-success)](#)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen)](#contributions)

This project exposes HTTP endpoints to create, read, update, delete, and search characters, including their devil fruits and Haki abilities.

---

## 📖 Table of Contents

- [✨ Features](#-features)
- [🚀 Quick Start](#-quick-start)
- [🔧 Configuration](#-configuration)
- [📚 API Overview](#-api-overview)
- [🧱 Data Model](#-data-model)
- [📁 Project Structure](#-project-structure)
- [🧪 Example Requests](#-example-requests)
- [🔄 Migrations (advanced)](#-migrations-advanced)
- [🧑‍💻 Local Development Tips](#-local-development-tips)
- [🤝 Contributions](#-contributions)

---

## ✨ Features

- **Full REST API** for One Piece characters (CRUD + search + filters).
- **Firestore‑backed** storage using Firebase’s official Go SDK.
- **Rich character model**: devil fruits, Haki abilities, and custom abilities.
- **CORS‑enabled** out of the box (ready to be consumed from frontends).
- **Health & root endpoints** for easy monitoring and discovery.
- **Migration toolkit** to evolve from embedded NoSQL documents to a normalized schema.

---

## 🚀 Quick Start

### 1. Clone and enter the project

```bash
git clone <your-repo-url>
cd onepiece-api
```

### 2. Configure environment variables

Use the example file as a starting point:

```bash
cp .env.example .env
```

Then set the variables (or export them in your shell profile):

```bash
export FIREBASE_PROJECT_ID=conectionwdb
export GOOGLE_APPLICATION_CREDENTIALS=/absolute/path/to/credentials.json
export PORT=8080
export FIRESTORE_COLLECTION=characters
```

- **`GOOGLE_APPLICATION_CREDENTIALS`** must point to your Firebase service account JSON file.
- **`FIREBASE_PROJECT_ID`** must match your Firebase project ID.

### 3. Install dependencies

```bash
go mod tidy
```

### 4. Run the server

```bash
go run main.go
```

By default the server runs on `http://localhost:8080`.

On startup you will see a summary of the available endpoints printed in the console, and by default two sample characters (**Luffy** and **Zoro**) are inserted on the first run if the collection is empty.

---

## 🔧 Configuration

All configuration is done via environment variables. Most of them have sensible defaults:

| Variable                      | Description                           | Default          |
|------------------------------|---------------------------------------|------------------|
| `FIREBASE_PROJECT_ID`        | Firebase project ID                   | `conectionwdb`   |
| `GOOGLE_APPLICATION_CREDENTIALS` | Absolute path to service account JSON | **required**     |
| `PORT`                       | HTTP server port                      | `8080`           |
| `FIRESTORE_COLLECTION`       | Firestore collection for characters   | `characters`     |

If `PORT` does not start with `:`, it will automatically be prefixed (e.g. `8080` → `:8080`).

---

## 📚 API Overview

Base URL (local):

```text
http://localhost:8080
```

### Characters

- **GET `/api/characters`**  
  Returns all characters.

- **GET `/api/characters/{id}`**  
  Returns a character by ID.

- **POST `/api/characters`**  
  Creates a new character.

- **PUT `/api/characters/{id}`**  
  Updates an existing character.

- **DELETE `/api/characters/{id}`**  
  Deletes a character by ID.

- **GET `/api/characters/search?name={name}`**  
  Searches characters by name or alias (case-insensitive).

- **GET `/api/characters/devil-fruits`**  
  Returns only characters that have a devil fruit.

### Routes & Islands (v2 — Dijkstra con 4 modos)

El motor de rutas modela el mundo como un grafo: **islas = nodos**, **rutas = aristas**. Cada arista carga 3 métricas (`distance`, `travelHours`, `danger 1–5`) y cada isla un `logPoseHours` (espera del Log Pose). El endpoint principal acepta 4 **modos** que optimizan distintos objetivos pero retornan **siempre las mismas 4 métricas globales**.

- **GET `/api/islands`** — todas las islas, con `logPoseHours`.
- **GET `/api/routes`** — todas las rutas, con `distance`/`travelHours`/`danger`.
- **GET `/api/routes/shortest?from={id}&to={id}&mode={mode}`**
  - `mode=fastest` (default) — minimiza distancia total.
  - `mode=quickest` — minimiza tiempo total (`travelHours` + `logPoseHours` de islas intermedias).
  - `mode=safest` — bottleneck min: minimiza el peligro **máximo** del camino.
  - `mode=riskiest` — bottleneck max: maximiza el peligro **mínimo** del camino.
- **GET `/api/routes/reachable?from={id}&maxCost={n}`** — islas alcanzables bajo presupuesto de distancia.

Toda respuesta de `shortest` incluye:

```jsonc
{
  "from": "windmill-village", "to": "wano", "mode": "quickest", "found": true, "hops": 14,
  "totalDistance": 5850,        // suma de distance del camino
  "totalTime": 591.67,          // travelHours + logPose de islas intermedias
  "worstDanger": 5,             // max(danger) del camino
  "bestDanger": 2,              // min(danger) del camino
  "totalCost": 591.67,          // legacy: refleja la métrica del modo activo
  "path": [
    { "islandId": "windmill-village", "islandName": "Windmill Village",
      "distanceSoFar": 0, "timeSoFar": 0, "worstDangerSoFar": 0, "bestDangerSoFar": 0, "costSoFar": 0 },
    // … steps con métricas acumuladas
  ]
}
```

**Ejemplos curl:**

```bash
# La distancia más corta
curl 'http://localhost:8080/api/routes/shortest?from=windmill-village&to=wano&mode=fastest' | jq

# El menor tiempo total (penaliza Log Pose)
curl 'http://localhost:8080/api/routes/shortest?from=windmill-village&to=wano&mode=quickest' | jq

# La ruta que evita los peores tramos
curl 'http://localhost:8080/api/routes/shortest?from=windmill-village&to=wano&mode=safest' | jq

# La ruta que evita los tramos más fáciles (turismo de aventura)
curl 'http://localhost:8080/api/routes/shortest?from=windmill-village&to=wano&mode=riskiest' | jq
```

Detalles algorítmicos: el modo `fastest`/`quickest` usa **Dijkstra clásico**; `safest`/`riskiest` usan **Dijkstra de cuello de botella** (bottleneck) reusando el mismo min-heap. Implementación en [pkg/graph/dijkstra.go](pkg/graph/dijkstra.go) y orquestación en [usecase/route_usecase.go](usecase/route_usecase.go).

### Routes — Análisis del grafo (v3)

Endpoint estructural pensado para auditoría y observabilidad del seed. Útil para verificar que el grafo está sano (un solo componente conexo) y que la distribución de `danger` es lo bastante diversa para que `safest` y `riskiest` diverjan de `fastest`.

- **GET `/api/routes/stats`** — métricas globales del grafo. Cacheado 60s (`Cache-Control: public, max-age=60`).

```jsonc
{
  "totalIslands": 32,
  "totalRoutes": 48,
  "bidirectionalCount": 41,         // rutas que se navegan en ambos sentidos
  "islandsWithLogPose": 19,         // islas con logPoseHours > 0
  "connectedComponents": 1,         // BFS no dirigido (bidi en ambos sentidos, uni solo en su sentido)
  "largestComponent": 32,           // tamaño del componente más grande
  "avgDistance": 478.42,
  "avgTravelHours": 29.90,
  "avgDanger": 3.02,
  "dangerHistogram": [5, 12, 14, 11, 6]   // índice i = rutas con danger (i+1)
}
```

```bash
curl -s 'http://localhost:8080/api/routes/stats' | jq
```

Frontend: la vista [`/stats`](frontend/src/pages/StatsPage.tsx) renderiza estos datos como tarjetas + histograma de Danger.

Auditoría adicional offline: `go run ./migration/cmd/audit` muestra distribuciones, componentes BFS y un sample aleatorio (seed 42) con tasa de divergencia entre los 4 modos.

### Health & Root

- **GET `/health`**  
  Simple health check:

  ```json
  {
    "status": "ok",
    "message": "One Piece API is running"
  }
  ```

- **GET `/`**  
  Welcome payload with basic metadata and a list of main endpoints.

---

## 🧱 Data Model

The main domain model is a `Character` stored in Firestore. Simplified structure:

```json
{
  "id": "luffy",
  "name": "Monkey D. Luffy",
  "alias": "Mugiwara",
  "species": "Human",
  "role": "Captain",
  "firstAppearance": "Chapter 1",
  "devilFruit": {
    "name": "Gomu Gomu no Mi (Hito Hito no Mi, Model: Nika)",
    "type": "Zoan",
    "description": "Awakened Zoan that grants rubber properties and imagination-based powers"
  },
  "hakiAbilities": [
    {
      "hakiType": "Armament",
      "proficiency": "Advanced",
      "awakened": true,
      "notes": "Internal destruction"
    }
  ],
  "abilities": [
    {
      "type": "Gear 5",
      "notes": "Awakened form"
    }
  ],
  "notes": "Future Pirate King"
}
```

Key constraints (enforced in service layer):

- `id` and `name` are **required**.
- `name` must be between 2 and 100 characters.
- If `devilFruit.type` is present, it must be one of: `Paramecia`, `Logia`, `Zoan`.
- Each Haki ability `proficiency` must be one of: `Basic`, `Advanced`, `Master`.

When creating a character, the service normalizes:

- `id` → lowercased and spaces replaced by `_`.
- `name` / `alias` → trimmed.

---

## 📁 Project Structure

High-level modules:

- `main.go` – Application entrypoint, server startup and initial sample data.
- `router/` – HTTP routing and CORS configuration.
- `controller/` – HTTP handlers and JSON marshaling/unmarshaling.
- `service/` – Business logic for characters and generic Firestore helpers.
- `repository/` – Firestore data access layer and domain models.
- `config/` – Firebase initialization and configuration helpers.
- `migration/` – Backup, migration, validation and rollback scripts for data structure changes.

There is also a Postman collection file `OnePiece_API.postman_collection.json` you can import to quickly test the endpoints.

---

## 🧪 Example Requests

### Create a character

```bash
curl -X POST http://localhost:8080/api/characters \
  -H "Content-Type: application/json" \
  -d '{
    "id": "sanji",
    "name": "Vinsmoke Sanji",
    "alias": "Black Leg",
    "species": "Human",
    "role": "Cook",
    "firstAppearance": "Chapter 43",
    "devilFruit": null,
    "hakiAbilities": [
      { "hakiType": "Armament", "proficiency": "Advanced", "awakened": true }
    ],
    "abilities": [
      { "type": "Diable Jambe", "notes": "Fire-enhanced kicks" }
    ],
    "notes": "Straw Hat Pirates cook"
  }'
```

### Search characters by name

```bash
curl "http://localhost:8080/api/characters/search?name=luffy"
```

---

## 🔄 Migrations (advanced)

The `migration/` directory contains scripts and documentation to migrate from the current embedded NoSQL structure to a more normalized schema (separate collections for devil fruits, Haki, abilities, etc.).

Refer to:

- `migration/README.md` – How to run backup, migrate, validate and rollback scripts.
- `migration/MIGRATION_STRATEGY.md` – Detailed migration and rollout strategy (phases, risks, and mitigation).

These tools are optional for running the API locally, but are useful if you plan to evolve the Firestore data model in production.

---

## 🧑‍💻 Local Development Tips

- Use the provided **Postman collection** `OnePiece_API.postman_collection.json` to explore and test endpoints quickly.
- The first time you run the API, it will insert **Luffy** and **Zoro** automatically if the collection is empty.
- For debugging Firestore issues, double‑check:
  - `GOOGLE_APPLICATION_CREDENTIALS` path.
  - `FIREBASE_PROJECT_ID` value.
  - Firestore rules permitting your environment.

---

## 🤝 Contributions

- Keep code and comments in **English** whenever possible.
- Follow the existing layering: `controller` → `service` → `repository`.
- When adding new fields or collections, update this README and (if needed) the migration docs.
- Prefer small, focused pull requests with clear descriptions and, if relevant, example requests/responses.

Feel free to open issues or propose improvements to the API, data model, or documentation.