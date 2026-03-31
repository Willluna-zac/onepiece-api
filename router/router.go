package router

import (
	"net/http"
	"strings"

	"onepiece-api/controller"
)

func SetupRoutes(characterController *controller.CharacterController) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/characters", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		switch r.Method {
		case http.MethodGet:
			characterController.GetAllCharacters(w, r)
		case http.MethodPost:
			characterController.CreateCharacter(w, r)
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
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

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
			characterController.UpdateCharacter(w, r)
		case http.MethodDelete:
			characterController.DeleteCharacter(w, r)
		default:
			http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","message":"One Piece API is running"}`))
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
