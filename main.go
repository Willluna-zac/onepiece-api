// @title           One Piece API
// @version         1.0
// @description     API REST para gestionar personajes e islas del mundo de One Piece. Proyecto de estudio con Clean Architecture en Go + Firestore.
// @contact.name    William Luna
// @license.name    MIT
// @host            localhost:8080
// @BasePath        /
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name X-API-Key
// @description Requerido para POST, PUT y DELETE. Configura la variable de entorno API_KEY en el servidor.
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"onepiece-api/config"
	"onepiece-api/controller"
	"onepiece-api/domain"
	_ "onepiece-api/docs"
	"onepiece-api/repository"
	"onepiece-api/router"
	"onepiece-api/usecase"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	// Cargar variables de entorno desde .env (si existe)
	_ = godotenv.Load()

	// Inicializar Firebase
	config.InitFirebase()
	defer config.CloseFirebase()

	// Crear instancias (bottom-up: Repository → UseCase → Controller)
	characterRepo := repository.NewCharacterRepository(config.FirestoreClient)
	characterUsecase := usecase.NewCharacterUsecase(characterRepo)
	characterController := controller.NewCharacterController(characterUsecase)

	islandRepo := repository.NewIslandRepository()
	islandUsecase := usecase.NewIslandUsecase(islandRepo)
	islandController := controller.NewIslandController(islandUsecase)

	handler := router.SetupRoutes(characterController, islandController)

	createExampleData(characterUsecase)

	// Iniciar servidor HTTP (puede configurarse con variable de entorno)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Valor por defecto
	}
	if port[0] != ':' {
		port = ":" + port
	}
	fmt.Printf("\n🚀 Servidor iniciado en http://localhost%s\n", port)
	fmt.Println("📚 Endpoints disponibles:")
	fmt.Println("  GET    /api/characters              - Obtener todos los personajes")
	fmt.Println("  GET    /api/characters/{id}         - Obtener un personaje por ID")
	fmt.Println("  POST   /api/characters              - Crear un personaje")
	fmt.Println("  PUT    /api/characters/{id}         - Actualizar un personaje")
	fmt.Println("  DELETE /api/characters/{id}         - Eliminar un personaje")
	fmt.Println("  GET    /api/characters/search?name= - Buscar personajes")
	fmt.Println("  GET    /api/characters/devil-fruits - Personajes con fruta del diablo")
	fmt.Println("  GET    /api/islands                       - Todas las islas del mundo")
	fmt.Println("  GET    /api/islands/{id}                  - Detalle de una isla")
	fmt.Println("  GET    /api/islands/nearest?x=&y=         - Isla más cercana (Quadtree)")
	fmt.Println("  GET    /api/islands/region/{region}       - Islas por región")
	fmt.Println("\n💡 Prueba con: curl http://localhost:8080/api/characters")
	fmt.Println()

	if err := http.ListenAndServe(port, handler); err != nil {
		log.Fatalf("❌ Error al iniciar servidor: %v", err)
	}
}

func createExampleData(uc *usecase.CharacterUsecase) {
	ctx := context.Background()

	existing, _ := uc.GetAllCharacters(ctx)
	if len(existing) > 0 {
		fmt.Println("✅ Ya existen personajes en la base de datos")
		return
	}

	fmt.Println("📝 Creando datos de ejemplo...")

	luffy := &domain.Character{
		Name:            "Monkey D. Luffy",
		Alias:           "Mugiwara",
		Species:         "Human",
		Role:            "Captain",
		FirstAppearance: "Chapter 1",
		DevilFruit: &domain.DevilFruit{
			Name:        "Hito Hito no Mi, Model: Nika",
			Type:        "Mythical Zoan",
			Description: "Awakened Zoan that grants rubber properties",
		},
		HakiAbilities: []domain.HakiAbility{
			{HakiType: "Armament", Proficiency: "Advanced", Awakened: true},
			{HakiType: "Observation", Proficiency: "Advanced", Awakened: true},
			{HakiType: "Conqueror", Proficiency: "Master", Awakened: true},
		},
		Abilities: []domain.Ability{
			{Type: "Gear 5", Notes: "Awakened form"},
		},
		Notes: "Future Pirate King",
	}

	uc.CreateCharacter(ctx, luffy)
	fmt.Println("✅ Datos de ejemplo creados")
}
