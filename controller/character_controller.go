package controller

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"onepiece-api/domain"
	"onepiece-api/usecase"
)

type CharacterController struct {
	usecase *usecase.CharacterUsecase
}

func NewCharacterController(uc *usecase.CharacterUsecase) *CharacterController {
	return &CharacterController{usecase: uc}
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

type SuccessResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
	Message string      `json:"message,omitempty"`
}

func (c *CharacterController) GetAllCharacters(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		c.sendError(w, http.StatusMethodNotAllowed, "Método no permitido")
		return
	}

	characters, err := c.usecase.GetAllCharacters(r.Context())
	if err != nil {
		log.Printf("[ERROR] GetAllCharacters: %v", err)
		c.sendError(w, http.StatusInternalServerError, "Error interno al obtener personajes")
		return
	}

	c.sendSuccess(w, http.StatusOK, characters, "")
}

func (c *CharacterController) GetCharacterByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		c.sendError(w, http.StatusMethodNotAllowed, "Método no permitido")
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/api/characters/")
	if id == "" || id == "/api/characters" {
		c.sendError(w, http.StatusBadRequest, "ID es requerido")
		return
	}

	character, err := c.usecase.GetCharacterByID(r.Context(), id)
	if err != nil {
		log.Printf("[ERROR] GetCharacterByID id=%s: %v", id, err)
		c.sendError(w, http.StatusNotFound, "Personaje no encontrado")
		return
	}

	c.sendSuccess(w, http.StatusOK, character, "")
}

func (c *CharacterController) CreateCharacter(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		c.sendError(w, http.StatusMethodNotAllowed, "Método no permitido")
		return
	}

	var character domain.Character
	if err := json.NewDecoder(r.Body).Decode(&character); err != nil {
		c.sendError(w, http.StatusBadRequest, "JSON inválido")
		return
	}
	defer r.Body.Close()

	if err := c.usecase.CreateCharacter(r.Context(), &character); err != nil {
		log.Printf("[ERROR] CreateCharacter: %v", err)
		c.sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	c.sendSuccess(w, http.StatusCreated, character, "Personaje creado exitosamente")
}

func (c *CharacterController) UpdateCharacter(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		c.sendError(w, http.StatusMethodNotAllowed, "Método no permitido")
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/api/characters/")
	if id == "" {
		c.sendError(w, http.StatusBadRequest, "ID es requerido")
		return
	}

	var character domain.Character
	if err := json.NewDecoder(r.Body).Decode(&character); err != nil {
		c.sendError(w, http.StatusBadRequest, "JSON inválido")
		return
	}
	defer r.Body.Close()

	character.ID = id

	if err := c.usecase.UpdateCharacter(r.Context(), &character); err != nil {
		log.Printf("[ERROR] UpdateCharacter id=%s: %v", id, err)
		c.sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	c.sendSuccess(w, http.StatusOK, character, "Personaje actualizado exitosamente")
}

func (c *CharacterController) DeleteCharacter(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		c.sendError(w, http.StatusMethodNotAllowed, "Método no permitido")
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/api/characters/")
	if id == "" {
		c.sendError(w, http.StatusBadRequest, "ID es requerido")
		return
	}

	if err := c.usecase.DeleteCharacter(r.Context(), id); err != nil {
		log.Printf("[ERROR] DeleteCharacter id=%s: %v", id, err)
		c.sendError(w, http.StatusNotFound, err.Error())
		return
	}

	c.sendSuccess(w, http.StatusOK, nil, "Personaje eliminado exitosamente")
}

func (c *CharacterController) SearchCharacters(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		c.sendError(w, http.StatusMethodNotAllowed, "Método no permitido")
		return
	}

	name := r.URL.Query().Get("name")
	if name == "" {
		c.sendError(w, http.StatusBadRequest, "Parámetro 'name' es requerido")
		return
	}

	characters, err := c.usecase.SearchByName(r.Context(), name)
	if err != nil {
		log.Printf("[ERROR] SearchByName name=%s: %v", name, err)
		c.sendError(w, http.StatusInternalServerError, "Error interno al buscar personajes")
		return
	}

	c.sendSuccess(w, http.StatusOK, characters, "")
}

func (c *CharacterController) GetCharactersWithDevilFruit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		c.sendError(w, http.StatusMethodNotAllowed, "Método no permitido")
		return
	}

	characters, err := c.usecase.GetCharactersWithDevilFruit(r.Context())
	if err != nil {
		log.Printf("[ERROR] GetCharactersWithDevilFruit: %v", err)
		c.sendError(w, http.StatusInternalServerError, "Error interno al obtener personajes")
		return
	}

	c.sendSuccess(w, http.StatusOK, characters, "")
}

func (c *CharacterController) sendError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(ErrorResponse{
		Error:   http.StatusText(statusCode),
		Message: message,
	})
}

func (c *CharacterController) sendSuccess(w http.ResponseWriter, statusCode int, data interface{}, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(SuccessResponse{
		Success: true,
		Data:    data,
		Message: message,
	})
}
