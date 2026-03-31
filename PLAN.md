# 🏴‍☠️ Onepiece API — Plan de Mejoras y Nuevas Features

> **Proyecto de estudio:** Go stdlib + Firestore + Clean Architecture  
> Este documento centraliza todos los cambios pendientes, organizados por categoría y prioridad.

---

## 📋 Índice

1. [🐛 Bugs](#-bugs)
2. [🏛️ Arquitectura](#-arquitectura)
3. [⚡ Performance](#-performance)
4. [🔒 Seguridad](#-seguridad)
5. [🧪 Testing](#-testing)
6. [🗺️ Feature: Quadtree — Mapa de Islas](#-feature-quadtree--mapa-de-islas)
7. [🖥️ Feature: Frontend React](#-feature-frontend-react)
8. [📄 Feature: FastAPI Docs](#-feature-fastapi-docs)
9. [🔧 Datos: Limpieza de BD](#-datos-limpieza-de-bd)

---

## 🐛 Bugs

### BUG-1 (HIGH) — Batch limit de 500 en `deleteNormalized`
**Archivo:** `repository/characters.go`

**Problema:** Firestore tiene un límite estricto de 500 operaciones por batch commit. En `deleteNormalized`, si un personaje tiene muchos registros en `character_haki` y `abilities`, el batch puede superar ese límite y **falla silenciosamente** (el error se ignora o el SDK hace panic).

**Por qué importa:** Un personaje muy complejo (ej. Luffy con múltiples estilos de Gear + haki) podría corromper datos en un DELETE.

**Fix:** Dividir las operaciones en chunks de máximo 499 ops cada uno:
```go
func commitInBatches(ctx context.Context, client *firestore.Client, ops []func(*firestore.WriteBatch)) error {
    const maxBatchSize = 499
    for i := 0; i < len(ops); i += maxBatchSize {
        end := min(i+maxBatchSize, len(ops))
        batch := client.Batch()
        for _, op := range ops[i:end] {
            op(batch)
        }
        if _, err := batch.Commit(ctx); err != nil {
            return err
        }
    }
    return nil
}
```

---

### BUG-2 (HIGH) — Haki/abilities huérfanos en UPDATE
**Archivo:** `repository/characters.go` → función `writeNormalized`

**Problema:** Cuando se actualiza un personaje cambiando su haki de 3 tipos a 2 tipos, `writeNormalized` escribe los 2 nuevos docs pero **no elimina** el tercero viejo. Los IDs son posicionales: `{charID}_haki_0`, `{charID}_haki_1`, `{charID}_haki_2`. El `_haki_2` queda huérfano en Firestore para siempre.

**Por qué importa:** Con cada update, la BD acumula datos basura que aparecen en reads posteriores.

**Fix:** Antes del batch de escritura en un update, hacer delete de todos los docs normalizados previos:
```go
// 1. Borrar todos los haki/abilities del personaje anterior
if err := r.deleteNormalized(ctx, charID); err != nil {
    return err
}
// 2. Escribir los nuevos
batch := r.client.Batch()
// ... writeNormalized como siempre
```

---

### BUG-3 (MEDIUM) — `go.mod` declara versión de Go inexistente
**Archivo:** `go.mod`, `migration/cmd/migrate/go.mod`

**Problema:** `go 1.25.1` no existe. Go va en 1.24.x al momento de escribir esto. El toolchain puede comportarse de forma inesperada.

**Fix:** Cambiar a `go 1.24.0`

---

### BUG-4 (MEDIUM) — Errores internos de Firestore expuestos al cliente HTTP
**Archivos:** `controller/character_controller.go`

**Problema:** Los errores de Firestore se devuelven directamente al cliente con `http.Error(w, err.Error(), ...)`. Esto expone detalles internos de la infraestructura (nombres de colecciones, rutas de documentos, etc.) que son un riesgo de seguridad y mala práctica.

**Por qué importa:** Principio de "never leak internals". El cliente solo necesita saber si falló, no por qué falló internamente.

**Fix:**
```go
// MAL:
http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)

// BIEN:
log.Printf("[ERROR] GetCharacter id=%s: %v", id, err)
http.Error(w, "internal server error", http.StatusInternalServerError)
```

---

### BUG-5 (LOW) — Postman collection usa IDs legacy (`luffy`, `nami`)
**Archivo:** `postman/Onepiece API.postman_collection.json`

**Problema:** Después de la migración a UUIDs, todos los IDs son del tipo `uuid-v4`. Las requests en la colección Postman usan los IDs viejos y retornan 404.

**Fix:** Reemplazar IDs hardcodeados por variables de entorno de Postman:
- Crear un Postman Environment con variable `CHARACTER_ID`
- Usar `{{CHARACTER_ID}}` en todas las requests de detalle/update/delete
- Documentar cómo obtener un ID válido (GET /api/characters → copiar un ID)

---

## 🏛️ Arquitectura

### ARCH-1 (HIGH) — Agregar interfaces para Repository y UseCase
**Archivos:** `usecase/character_usecase.go`, `repository/characters.go`

**Problema:** El `UseCase` depende directamente de `*repository.CharacterRepository` (struct concreta). Esto viola el principio de Dependency Inversion y hace imposible hacer mocking en tests.

**Por qué importa:** Sin interfaz, no hay unit tests reales. Con interfaz, el usecase puede probarse sin Firestore.

**Fix — Crear `domain/ports.go`:**
```go
package domain

type CharacterRepository interface {
    GetAll(ctx context.Context) ([]Character, error)
    GetByID(ctx context.Context, id string) (*Character, error)
    Create(ctx context.Context, c Character) (*Character, error)
    Update(ctx context.Context, id string, c Character) (*Character, error)
    Delete(ctx context.Context, id string) error
    SearchByName(ctx context.Context, name string) ([]Character, error)
    GetWithDevilFruit(ctx context.Context) ([]Character, error)
}

type CharacterUseCase interface {
    GetAllCharacters(ctx context.Context) ([]Character, error)
    // ... mismos métodos
}
```

---

### ARCH-2 (LOW) — CORS duplicado en cada handler
**Archivo:** `router/router.go`

**Problema:** Los headers CORS se configuran dentro de cada case del switch, duplicando ~4 líneas en cada handler. Si cambia el dominio permitido, hay que editar N lugares.

**Fix:** Crear un middleware CORS que envuelva el mux completo:
```go
func corsMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
        if r.Method == http.MethodOptions {
            w.WriteHeader(http.StatusNoContent)
            return
        }
        next.ServeHTTP(w, r)
    })
}
// En main: http.ListenAndServe(addr, corsMiddleware(mux))
```

---

## ⚡ Performance

### PERF-1 (HIGH) — N+1 Queries en `GetAllCharacters`
**Archivo:** `repository/characters.go` → `assembleCharacter`

**Problema:** Para cargar N personajes, se hace:
- 1 query `GetAll` a `characters_new` → retorna N docs
- Por cada personaje: 3 queries más (devilfruits, character_haki, abilities)
- **Total: 1 + 3N queries**

Con 100 personajes = 301 queries. Con 1000 = 3001 queries.

**Fix (corto plazo):** Paralelizar las 3 sub-queries con `errgroup`:
```go
var g errgroup.Group
g.Go(func() error { /* fetch devil fruit */ return nil })
g.Go(func() error { /* fetch haki */ return nil })
g.Go(func() error { /* fetch abilities */ return nil })
if err := g.Wait(); err != nil { return nil, err }
```
Esto reduce el tiempo de 3N queries secuenciales a N queries paralelas.

**Fix (largo plazo):** Desnormalizar los datos en `characters_new` para que un solo doc contenga todo (devil fruit embebida, haki array, abilities array). Elimina el N+1 completamente.

---

### PERF-2 (MEDIUM) — `SearchByName` y `GetCharactersWithDevilFruit` hacen full scan
**Archivo:** `repository/characters.go`

**Problema:**
- `SearchByName`: carga TODOS los personajes en memoria y filtra en Go con `strings.Contains`. Si hay 10,000 personajes, trae 10,000 docs para filtrar 2.
- `GetCharactersWithDevilFruit`: mismo patrón, carga todo para filtrar los que tienen `devil_fruit != nil`.

**Fix:**
```go
// SearchByName — usar Firestore range query (requiere índice compuesto):
q := r.client.Collection("characters_new").
    Where("name", ">=", name).
    Where("name", "<=", name+"\uf8ff")

// GetWithDevilFruit — query directa a la colección devilfruits:
docs, err := r.client.Collection("devilfruits").Documents(ctx).GetAll()
// Luego fetch de characters por los IDs obtenidos
```

---

## 🔒 Seguridad

### SEC-1 (MEDIUM) — Sin autenticación en endpoints de escritura
**Archivo:** `router/router.go`

**Problema:** `POST /api/characters`, `PUT /api/characters/{id}`, y `DELETE /api/characters/{id}` son completamente públicos. Cualquiera puede crear/modificar/borrar datos.

**Por qué importa:** Aunque es un proyecto de estudio, practicar auth desde el inicio es fundamental.

**Fix — API Key middleware simple:**
```go
func apiKeyMiddleware(apiKey string, next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        if r.Header.Get("X-API-Key") != apiKey {
            http.Error(w, "unauthorized", http.StatusUnauthorized)
            return
        }
        next(w, r)
    }
}
// En el router, envolver solo las rutas de escritura
```

---

## 🧪 Testing

### TEST-1 (HIGH) — Zero unit tests
**Depende de:** ARCH-1 (interfaces)

**Problema:** No existe ningún archivo `*_test.go`. Esto hace imposible validar cambios sin correr el servidor completo contra Firestore real.

**Plan de tests:**
```
usecase/
  character_usecase_test.go   ← mock del repository (tabla driven tests)
repository/
  characters_integration_test.go  ← tests contra Firestore emulador
controller/
  character_controller_test.go    ← httptest.NewRecorder
```

**Mocking con interfaz:**
```go
type mockRepo struct {
    characters []domain.Character
}
func (m *mockRepo) GetAll(ctx context.Context) ([]domain.Character, error) {
    return m.characters, nil
}
// ... implementar la interfaz con datos fake
```

---

## 🗺️ Feature: Quadtree — Mapa de Islas

### Descripción
Implementar un **Quadtree** para indexación espacial 2D del mundo de One Piece. Dado un punto (x, y) del mapa, encontrar la isla más cercana.

### ¿Qué es un Quadtree?
Un Quadtree divide el espacio 2D en 4 cuadrantes recursivamente. Cada nodo puede tener hasta N puntos; si se supera, se subdivide. Ideal para búsqueda de vecinos más cercanos (KNN) en 2D.

```
+--+--+
|NW|NE|
+--+--+
|SW|SE|
+--+--+
       → subdivide cada cuadrante si hay más de N puntos
```

### Coordenadas del mundo One Piece
Sistema: X (0-10000, Oeste→Este), Y (0-5000, Sur→Norte)
- East Blue: X 0-2000
- Grand Line (primera mitad): X 2000-5000  
- New World (segunda mitad): X 5000-9000
- Grand Line corre horizontalmente en Y≈2500

### Islas a implementar (30 locaciones)

| ID | Nombre | Región | X | Y |
|----|--------|--------|---|---|
| 1 | Windmill Village | East Blue | 500 | 1800 |
| 2 | Shells Town | East Blue | 700 | 2200 |
| 3 | Orange Town | East Blue | 900 | 2100 |
| 4 | Syrup Village | East Blue | 1100 | 2000 |
| 5 | Baratie | East Blue | 1400 | 2300 |
| 6 | Loguetown | East Blue | 1900 | 2500 |
| 7 | Reverse Mountain | Grand Line | 2100 | 2500 |
| 8 | Whiskey Peak | Grand Line | 2500 | 2100 |
| 9 | Little Garden | Grand Line | 2800 | 2500 |
| 10 | Alabasta | Grand Line | 3200 | 2300 |
| 11 | Jaya | Grand Line | 3600 | 2200 |
| 12 | Skypiea | Sky Islands | 3800 | 4800 |
| 13 | Long Ring Long Land | Grand Line | 4000 | 2400 |
| 14 | Water 7 | Grand Line | 4300 | 2700 |
| 15 | Enies Lobby | Grand Line | 4600 | 2600 |
| 16 | Thriller Bark | Grand Line | 4800 | 2800 |
| 17 | Sabaody Archipelago | Grand Line | 4900 | 2500 |
| 18 | Amazon Lily | New World | 5300 | 1200 |
| 19 | Impel Down | Red Line | 5000 | 1800 |
| 20 | Marineford | Red Line | 5000 | 2400 |
| 21 | Fish-Man Island | New World | 5500 | 2300 |
| 22 | Punk Hazard | New World | 5800 | 2600 |
| 23 | Dressrosa | New World | 6200 | 2200 |
| 24 | Zou | New World | 6600 | 2900 |
| 25 | Whole Cake Island | New World | 6800 | 3100 |
| 26 | Wano Country | New World | 7200 | 2700 |
| 27 | Elbaf | New World | 7600 | 4200 |
| 28 | Laugh Tale | New World | 8800 | 2500 |
| 29 | Kano Country | North Blue | 1200 | 4200 |
| 30 | Banaro Island | South Blue | 3000 | 800 |

### Estructura de archivos

```
onepiece-api/
├── pkg/
│   └── quadtree/
│       ├── quadtree.go      ← implementación genérica del quadtree
│       └── quadtree_test.go ← tests de inserción y búsqueda KNN
├── domain/
│   └── island.go            ← entidad Island con coordenadas
├── repository/
│   └── island_repository.go ← datos en memoria (o Firestore)
├── usecase/
│   └── island_usecase.go    ← lógica: nearest island, get all
└── controller/
    └── island_controller.go ← HTTP handlers
```

### Nuevos Endpoints

| Método | URL | Descripción |
|--------|-----|-------------|
| GET | `/api/islands` | Lista todas las islas |
| GET | `/api/islands/{id}` | Detalle de una isla |
| GET | `/api/islands/nearest?x=4300&y=2500` | Isla más cercana al punto dado |
| GET | `/api/islands/region/{region}` | Islas filtradas por región |

### Implementación del Quadtree

```go
// pkg/quadtree/quadtree.go
type Point struct {
    X, Y float64
    Data interface{} // la isla
}

type Bounds struct {
    MinX, MinY, MaxX, MaxY float64
}

type Quadtree struct {
    bounds   Bounds
    capacity int
    points   []Point
    divided  bool
    children [4]*Quadtree // NW, NE, SW, SE
}

func (qt *Quadtree) Insert(p Point) bool
func (qt *Quadtree) QueryNearest(x, y float64) *Point
func (qt *Quadtree) QueryRange(bounds Bounds) []Point
```

---

## 🖥️ Feature: Frontend React

### Stack recomendado
- **React 18 + TypeScript** — tipado fuerte, componentes reutilizables
- **Vite** — build tool ultrarrápido para desarrollo
- **Tailwind CSS** — utilidades CSS, perfecto para tema custom
- **React Query (TanStack Query)** — manejo de estado del servidor, caché automática
- **React Router v6** — navegación SPA
- **Leaflet.js o Canvas API** — para el mapa interactivo del mundo One Piece

### Tema visual One Piece
```
Paleta de colores:
  - Fondo: #0a1628 (azul océano profundo)
  - Primario: #d4a017 (dorado tesoro)
  - Acento: #8b1a1a (rojo pirata)
  - Texto: #f5e6c8 (pergamino)
  - Logia (DF): #3b82f6 (azul)
  - Zoan (DF): #f97316 (naranja)
  - Paramecia (DF): #a855f7 (morado)

Tipografía:
  - Títulos: "Pirata One" (Google Fonts) o "Cinzel"
  - Cuerpo: "Crimson Text" o "Lora" (serif legible)
```

### Estructura de páginas

```
/                      → Home (logo animado + navegación estilo mapa del tesoro)
/characters            → Grid de personajes con filtros (región, tipo DF, haki)
/characters/{id}       → Detalle completo del personaje
/map                   → Mapa interactivo del mundo One Piece + quadtree nearest island
/islands               → Lista de islas con filtro por región
```

### Componentes principales

```
components/
├── CharacterCard/
│   ├── CharacterCard.tsx       ← Card con imagen, nombre, tipo DF
│   └── DevilFruitBadge.tsx     ← Badge con color por tipo (Logia/Zoan/Paramecia)
├── HakiBar/
│   └── HakiBar.tsx             ← Visualización de tipos de haki como barras
├── Map/
│   ├── WorldMap.tsx            ← Canvas o SVG del mundo OP
│   ├── IslandPin.tsx           ← Punto interactivo en el mapa
│   └── NearestIslandFinder.tsx ← Click en mapa → llama /api/islands/nearest
├── Layout/
│   ├── Navbar.tsx              ← Navegación pirata con skull logo
│   └── Footer.tsx
└── Search/
    └── SearchBar.tsx           ← Búsqueda con debounce 300ms
```

### Estructura de archivos

```
frontend/               ← en la raíz del proyecto
├── src/
│   ├── api/            ← funciones fetch hacia el Go API
│   │   ├── characters.ts
│   │   └── islands.ts
│   ├── components/
│   ├── pages/
│   ├── hooks/          ← useCharacters, useNearestIsland
│   ├── types/          ← tipos TS que espejean los domain structs de Go
│   └── App.tsx
├── public/
│   └── world-map.svg   ← imagen base del mapa de One Piece
├── index.html
├── vite.config.ts
└── tailwind.config.ts
```

### Inicializar el proyecto

```bash
cd /path/to/onepiece-api
npm create vite@latest frontend -- --template react-ts
cd frontend
npm install
npm install -D tailwindcss postcss autoprefixer
npx tailwindcss init -p
npm install @tanstack/react-query react-router-dom leaflet @types/leaflet
```

---

## 📄 Feature: FastAPI Docs

### Propósito
Crear un microservicio Python con **FastAPI** que:
1. **Documenta** todos los endpoints del Go API con modelos Pydantic completos
2. **Auto-genera** Swagger UI (`/docs`) y ReDoc (`/redoc`) con OpenAPI 3.0
3. **Proxea** opcionalmente requests al backend Go (útil para dev local)

### Por qué FastAPI para documentar
- FastAPI genera documentación interactiva automáticamente desde los tipos Python
- Permite probar endpoints directamente desde el browser
- Sirve como "contrato" visual del API para el frontend

### Estructura de archivos

```
docs-api/               ← carpeta en la raíz del proyecto
├── main.py             ← app FastAPI + router
├── models/
│   ├── character.py    ← Pydantic models (espejo de domain/character.go)
│   └── island.py
├── routers/
│   ├── characters.py   ← endpoints /api/characters/*
│   └── islands.py      ← endpoints /api/islands/*
├── proxy.py            ← httpx client que proxea al Go backend
├── requirements.txt
└── README.md
```

### Modelos Pydantic (espejo de Go structs)

```python
# models/character.py
from pydantic import BaseModel
from typing import Optional, List
from enum import Enum

class DevilFruitType(str, Enum):
    LOGIA = "Logia"
    ZOAN = "Zoan"
    PARAMECIA = "Paramecia"

class DevilFruit(BaseModel):
    name: str
    type: DevilFruitType
    ability: str

class Haki(BaseModel):
    type: str      # Observation, Armament, Conqueror
    level: str     # Basic, Advanced, Supreme

class Character(BaseModel):
    id: str
    name: str
    age: Optional[int]
    origin: Optional[str]
    bounty: Optional[int]
    devil_fruit: Optional[DevilFruit]
    haki: Optional[List[Haki]]
    abilities: Optional[List[str]]
    crew: Optional[str]
    role: Optional[str]

class CreateCharacterRequest(BaseModel):
    name: str
    age: Optional[int] = None
    # ...
```

### Endpoints documentados

```python
# routers/characters.py
@router.get("/api/characters", response_model=List[Character], tags=["Characters"])
async def get_all_characters():
    """Retorna todos los personajes de One Piece almacenados en Firestore."""

@router.get("/api/characters/search", response_model=List[Character], tags=["Characters"])
async def search_by_name(name: str = Query(..., min_length=1, description="Nombre a buscar")):
    """Búsqueda de personajes por nombre (case-insensitive, partial match)."""

@router.get("/api/islands/nearest", response_model=Island, tags=["Islands"])
async def nearest_island(
    x: float = Query(..., ge=0, le=10000, description="Coordenada X en el mapa (0=Oeste, 10000=Este)"),
    y: float = Query(..., ge=0, le=5000, description="Coordenada Y en el mapa (0=Sur, 5000=Norte)")
):
    """Encuentra la isla más cercana al punto dado usando el índice Quadtree."""
```

### Inicializar el proyecto

```bash
cd /path/to/onepiece-api
mkdir docs-api && cd docs-api
python3 -m venv venv
source venv/bin/activate
pip install fastapi uvicorn httpx pydantic

# Correr en dev:
uvicorn main:app --reload --port 8000
# Swagger UI: http://localhost:8000/docs
# ReDoc:      http://localhost:8000/redoc
```

---

## 🔧 Datos: Limpieza de BD

### DATA-1 — Ace duplicado en Firestore
**Problema:** Hay dos documentos de `Portgas D. Ace` en la colección `characters_new` con distintos UUIDs. Esto causa duplicados en `GET /api/characters`.

**Fix:** Via script Go o Firestore Console:
1. Obtener todos los docs de `Portgas D. Ace`
2. Conservar el que tiene más datos completos
3. Eliminar el duplicado y sus normalized docs (`_haki_*`, `_ability_*`)

---

## 📊 Resumen de Prioridades

| Prioridad | Items |
|-----------|-------|
| 🔴 HIGH | BUG-1, BUG-2, ARCH-1, PERF-1, TEST-1 |
| 🟡 MEDIUM | BUG-3, BUG-4, PERF-2, SEC-1 |
| 🟢 LOW | BUG-5, ARCH-2, DATA-1 |
| 🆕 FEATURES | Quadtree, Frontend, FastAPI Docs |

## 🔄 Orden de implementación sugerido

```
1. BUG-3 (go.mod)           ← 2 min, evita confusión de toolchain
2. BUG-4 (errores expuestos) ← antes de exponer a FastAPI
3. ARCH-1 (interfaces)       ← desbloquea TEST-1 y PERF-1
4. BUG-1 + BUG-2 (batches)  ← corrigen corrupción de datos
5. TEST-1 (unit tests)       ← con las interfaces ya listas
6. PERF-1 (N+1)             ← con errgroup, después de tener tests
7. feat-quadtree             ← nueva feature Go pura
8. feat-fastapi-docs         ← Python, documenta Go + quadtree
9. feat-frontend             ← consume Go API + llama nearest island
10. SEC-1 (auth)             ← antes de desplegar el frontend
```
