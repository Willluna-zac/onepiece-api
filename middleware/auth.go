package middleware

import (
	"encoding/json"
	"net/http"
	"os"
)

// APIKeyMiddleware protege un handler verificando el header X-API-Key.
// La clave esperada se lee de la variable de entorno API_KEY.
// Si API_KEY no está configurada, las rutas de escritura están deshabilitadas
// por seguridad (fail-closed: mejor rechazar que aceptar sin autenticación).
func APIKeyMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		apiKey := os.Getenv("API_KEY")
		if apiKey == "" {
			writeUnauthorized(w, "Escrituras deshabilitadas: API_KEY no configurada en el servidor")
			return
		}
		if r.Header.Get("X-API-Key") != apiKey {
			writeUnauthorized(w, "API Key inválida o ausente")
			return
		}
		next(w, r)
	}
}

func writeUnauthorized(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	json.NewEncoder(w).Encode(map[string]string{
		"error":   "Unauthorized",
		"message": msg,
	})
}
