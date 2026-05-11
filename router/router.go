package router

import (
	"net/http"
	"strings"

	"onepiece-api/controller"
	"onepiece-api/middleware"

	httpSwagger "github.com/swaggo/http-swagger"
)

func SetupRoutes(characterController *controller.CharacterController, islandController *controller.IslandController, routeController *controller.RouteController) http.Handler {
	mux := http.NewServeMux()
	imageProxy := controller.NewImageProxyController()

	// Image proxy — permite cargar imágenes externas sin problemas de CORS/hotlink
	mux.HandleFunc("/api/proxy/image", imageProxy.ProxyImage)

	mux.HandleFunc("/api/characters", func(w http.ResponseWriter, r *http.Request) {
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
		characterController.SearchCharacters(w, r)
	})

	mux.HandleFunc("/api/characters/devil-fruits", func(w http.ResponseWriter, r *http.Request) {
		characterController.GetCharactersWithDevilFruit(w, r)
	})

	mux.HandleFunc("/api/characters/", func(w http.ResponseWriter, r *http.Request) {
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

	// Swagger UI
	mux.Handle("/docs/", httpSwagger.WrapHandler)

	// Islands — rutas específicas antes del wildcard /{id}
	mux.HandleFunc("/api/islands", func(w http.ResponseWriter, r *http.Request) {
		islandController.GetAllIslands(w, r)
	})

	mux.HandleFunc("/api/islands/nearest", func(w http.ResponseWriter, r *http.Request) {
		islandController.GetNearestIsland(w, r)
	})

	mux.HandleFunc("/api/islands/region/", func(w http.ResponseWriter, r *http.Request) {
		islandController.GetIslandsByRegion(w, r)
	})

	mux.HandleFunc("/api/islands/", func(w http.ResponseWriter, r *http.Request) {
		islandController.GetIslandByID(w, r)
	})

	// Routes — rutas marítimas y navegación con Dijkstra
	// Orden: específicas antes del wildcard /from/
	mux.HandleFunc("/api/routes", func(w http.ResponseWriter, r *http.Request) {
		routeController.GetAllRoutes(w, r)
	})

	mux.HandleFunc("/api/routes/shortest", func(w http.ResponseWriter, r *http.Request) {
		routeController.GetShortestPath(w, r)
	})

	mux.HandleFunc("/api/routes/reachable", func(w http.ResponseWriter, r *http.Request) {
		routeController.GetReachableIslands(w, r)
	})

	mux.HandleFunc("/api/routes/stats", func(w http.ResponseWriter, r *http.Request) {
		routeController.GetGraphStats(w, r)
	})

	mux.HandleFunc("/api/routes/from/", func(w http.ResponseWriter, r *http.Request) {
		routeController.GetRoutesFromIsland(w, r)
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message":"Welcome to One Piece API","version":"1.0.0"}`))
	})

	// CORSMiddleware envuelve todo el mux — un único lugar para la política CORS
	return middleware.CORSMiddleware(mux)
}
