package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"onepiece-api/config"
	"onepiece-api/controller"
	"onepiece-api/domain"
	"onepiece-api/repository"
	"onepiece-api/router"
	"onepiece-api/usecase"
	"os"
)

func main() {
	// Inicializar Firebase
	config.InitFirebase()
	defer config.CloseFirebase()

	// Crear instancias (bottom-up: Repository → Service → Controller)
	characterRepo := repository.NewCharacterRepository(config.FirestoreClient)
	characterUsecase := usecase.NewCharacterUsecase(characterRepo)
	characterController := controller.NewCharacterController(characterUsecase)

	handler := router.SetupRoutes(characterController)

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
	fmt.Println("  GET    /health                      - Estado del servidor")
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
