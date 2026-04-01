package router

import (
	"net/http"
	"strings"

	"onepiece-api/controller"
	"onepiece-api/middleware"

	httpSwagger "github.com/swaggo/http-swagger"
)

func SetupRoutes(characterController *controller.CharacterController, islandController *controller.IslandController) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/characters", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-API-Key")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		switch r.Method {
		case http.MethodGet:
			characterController.GetAllCharacters(w, r)
		case http.MethodPost:
			middleware.APIKeyMiddleware(characterController.CreateCharacter)(w, r)
		default:
			http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/characters/search", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		characterController.SearchCharacters(w, r)
	})

	mux.HandleFunc("/api/characters/devil-fruits", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		characterController.GetCharactersWithDevilFruit(w, r)
	})

	mux.HandleFunc("/api/characters/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-API-Key")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		id := strings.TrimPrefix(r.URL.Path, "/api/characters/")
		if id == "" {
			http.Error(w, "ID es requerido", http.StatusBadRequest)
			return
		}

		switch r.Method {
		case http.MethodGet:
			characterController.GetCharacterByID(w, r)
		case http.MethodPut:
			middleware.APIKeyMiddleware(characterController.UpdateCharacter)(w, r)
		case http.MethodDelete:
			middleware.APIKeyMiddleware(characterController.DeleteCharacter)(w, r)
		default:
			http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","message":"One Piece API is running"}`))
	})

	// Swagger UI — documentación interactiva en /docs/index.html
	mux.Handle("/docs/", httpSwagger.WrapHandler)

	// -----------------------------------------------------------------------
	// Islands — orden importante: rutas específicas antes que el wildcard /{id}
	// -----------------------------------------------------------------------
	mux.HandleFunc("/api/islands", func(w http.ResponseWriter, r *http.Request) {
		setCORS(w, "GET, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		if r.Method != http.MethodGet {
			http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
			return
		}
		islandController.GetAllIslands(w, r)
	})

	mux.HandleFunc("/api/islands/nearest", func(w http.ResponseWriter, r *http.Request) {
		setCORS(w, "GET, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		if r.Method != http.MethodGet {
			http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
			return
		}
		islandController.GetNearestIsland(w, r)
	})

	mux.HandleFunc("/api/islands/region/", func(w http.ResponseWriter, r *http.Request) {
		setCORS(w, "GET, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		if r.Method != http.MethodGet {
			http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
			return
		}
		islandController.GetIslandsByRegion(w, r)
	})

	mux.HandleFunc("/api/islands/", func(w http.ResponseWriter, r *http.Request) {
		setCORS(w, "GET, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		if r.Method != http.MethodGet {
			http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
			return
		}
		islandController.GetIslandByID(w, r)
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message":"Welcome to One Piece API","version":"1.0.0","endpoints":["/api/characters","/api/characters/{id}","/api/characters/search?name=xxx","/api/characters/devil-fruits","/health"]}`))
	})

	return mux
}

// setCORS agrega los headers CORS estándar a la respuesta.
func setCORS(w http.ResponseWriter, methods string) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", methods)
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-API-Key")
}
