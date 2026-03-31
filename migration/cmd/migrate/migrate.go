package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

// Configuración de migración
const (
	BATCH_SIZE     = 500   // Registros por lote
	NUM_WORKERS    = 5     // Workers paralelos
	CHECKPOINT_INT = 10000 // Guardar checkpoint cada N registros
	SLEEP_BETWEEN  = 2     // Milisegundos entre lotes
)

// ---------------------------------------------------------------------------
// Estructuras VIEJAS (actuales)
// ---------------------------------------------------------------------------

type OldCharacter struct {
	ID              string           `firestore:"id"`
	Name            string           `firestore:"name"`
	Alias           string           `firestore:"alias"`
	Species         string           `firestore:"species"`
	Role            string           `firestore:"role"`
	FirstAppearance string           `firestore:"firstAppearance"`
	DevilFruit      *OldDevilFruit   `firestore:"devilFruit,omitempty"`
	HakiAbilities   []OldHakiAbility `firestore:"hakiAbilities,omitempty"`
	Abilities       []OldAbility     `firestore:"abilities,omitempty"`
	Notes           string           `firestore:"notes,omitempty"`
}

type OldDevilFruit struct {
	Name        string `firestore:"name"`
	Type        string `firestore:"type"`
	Description string `firestore:"description,omitempty"`
}

type OldHakiAbility struct {
	HakiType    string `firestore:"hakiType"`
	Proficiency string `firestore:"proficiency"`
	Awakened    bool   `firestore:"awakened"`
	Notes       string `firestore:"notes,omitempty"`
}

type OldAbility struct {
	Type  string `firestore:"type"`
	Notes string `firestore:"notes,omitempty"`
}

// ---------------------------------------------------------------------------
// Estructuras NUEVAS (normalizadas)
// ---------------------------------------------------------------------------

type NewCharacter struct {
	ID              string    `firestore:"id"`
	Name            string    `firestore:"name"`
	Alias           string    `firestore:"alias"`
	Species         string    `firestore:"species"`
	Role            string    `firestore:"role"`
	FirstAppearance string    `firestore:"firstAppearance"`
	Notes           string    `firestore:"notes,omitempty"`
	MigratedAt      time.Time `firestore:"migratedAt"`
}

type NewDevilFruit struct {
	FruitID     string    `firestore:"fruit_id"`
	CharacterID string    `firestore:"character_id"`
	Name        string    `firestore:"name"`
	Type        string    `firestore:"type"`
	Description string    `firestore:"description,omitempty"`
	CreatedAt   time.Time `firestore:"createdAt"`
}

type HakiType struct {
	ID          string `firestore:"id"`
	Name        string `firestore:"name"`
	Description string `firestore:"description,omitempty"`
}

type CharacterHaki struct {
	ID          string    `firestore:"id"`
	CharacterID string    `firestore:"character_id"`
	HakiTypeID  string    `firestore:"haki_type_id"`
	Proficiency string    `firestore:"proficiency"`
	Awakened    bool      `firestore:"awakened"`
	Notes       string    `firestore:"notes,omitempty"`
	CreatedAt   time.Time `firestore:"createdAt"`
}

type NewAbility struct {
	ID          string    `firestore:"id"`
	CharacterID string    `firestore:"character_id"`
	Type        string    `firestore:"type"`
	Notes       string    `firestore:"notes,omitempty"`
	CreatedAt   time.Time `firestore:"createdAt"`
}

// ---------------------------------------------------------------------------
// Estado de migración (checkpoint)
// ---------------------------------------------------------------------------

type MigrationState struct {
	TotalRecords    int       `json:"total_records"`
	MigratedRecords int       `json:"migrated_records"`
	FailedRecords   int       `json:"failed_records"`
	LastCheckpoint  int       `json:"last_checkpoint"`
	StartTime       time.Time `json:"start_time"`
	LastUpdateTime  time.Time `json:"last_update_time"`
	Status          string    `json:"status"`
	ErrorLog        []string  `json:"error_log,omitempty"`
}

// ---------------------------------------------------------------------------
// Variables globales
// ---------------------------------------------------------------------------

var (
	client         *firestore.Client
	ctx            context.Context
	migrationState MigrationState
	stateMutex     sync.Mutex
	hakiTypesMap   map[string]string // nombre -> ID
)

// ---------------------------------------------------------------------------
// main
// ---------------------------------------------------------------------------

func main() {
	ctx = context.Background()

	projectID := os.Getenv("FIREBASE_PROJECT_ID")
	if projectID == "" {
		projectID = "conectionwdb"
	}

	var err error
	client, err = firestore.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("❌ Error al conectar a Firestore: %v", err)
	}
	defer client.Close()

	fmt.Println("🚀 Iniciando Migración de Datos")
	fmt.Println("================================")
	fmt.Println()

	// FASE 1: Configuración inicial
	fmt.Println("📋 FASE 1: Configuración Inicial")
	if err := setupMigration(); err != nil {
		log.Fatalf("❌ Error en setup: %v", err)
	}

	// FASE 2: Poblar catálogo HakiTypes
	fmt.Println("\n📋 FASE 2: Crear Catálogo HakiTypes")
	if err := createHakiTypes(); err != nil {
		log.Fatalf("❌ Error creando HakiTypes: %v", err)
	}

	// FASE 3: Contar registros totales
	fmt.Println("\n📋 FASE 3: Contando Registros Totales")
	totalRecords, err := countTotalRecords()
	if err != nil {
		log.Fatalf("❌ Error contando registros: %v", err)
	}
	migrationState.TotalRecords = totalRecords
	migrationState.StartTime = time.Now()
	fmt.Printf("✅ Total de personajes a migrar: %d\n", totalRecords)

	// FASE 4: Migración por lotes
	fmt.Println("\n📋 FASE 4: Migración por Lotes")
	fmt.Printf("   Batch size: %d\n", BATCH_SIZE)
	fmt.Printf("   Workers: %d\n", NUM_WORKERS)
	fmt.Printf("   Checkpoint cada: %d registros\n\n", CHECKPOINT_INT)
	if err := migrateInBatches(); err != nil {
		log.Fatalf("❌ Error en migración: %v", err)
	}

	// FASE 5: Validación
	fmt.Println("\n📋 FASE 5: Validación")
	if err := validateMigration(); err != nil {
		log.Printf("⚠️  Advertencias en validación: %v", err)
	}

	printFinalReport()
}

// ---------------------------------------------------------------------------
// Setup
// ---------------------------------------------------------------------------

func setupMigration() error {
	stateFile := "migration_state.json"

	if data, err := os.ReadFile(stateFile); err == nil {
		json.Unmarshal(data, &migrationState)
		fmt.Printf("✅ Estado previo cargado: %d/%d registros migrados\n",
			migrationState.MigratedRecords, migrationState.TotalRecords)
	} else {
		migrationState = MigrationState{
			Status:    "initialized",
			ErrorLog:  make([]string, 0),
			StartTime: time.Now(),
		}
		fmt.Println("✅ Nuevo estado de migración inicializado")
	}
	return nil
}

func createHakiTypes() error {
	hakiTypesMap = make(map[string]string)

	hakiTypes := []HakiType{
		{ID: "armament", Name: "Armament", Description: "Haki de armadura"},
		{ID: "observation", Name: "Observation", Description: "Haki de observación"},
		{ID: "conqueror", Name: "Conqueror", Description: "Haki del conquistador"},
	}

	for _, ht := range hakiTypes {
		doc, err := client.Collection("hakitypes").Doc(ht.ID).Get(ctx)
		if err == nil && doc.Exists() {
			fmt.Printf("   ⏭️  HakiType '%s' ya existe\n", ht.Name)
		} else {
			_, err := client.Collection("hakitypes").Doc(ht.ID).Set(ctx, ht)
			if err != nil {
				return fmt.Errorf("error creando HakiType %s: %v", ht.Name, err)
			}
			fmt.Printf("   ✅ HakiType '%s' creado\n", ht.Name)
		}
		hakiTypesMap[ht.Name] = ht.ID
	}
	return nil
}

// ---------------------------------------------------------------------------
// Conteo
// ---------------------------------------------------------------------------

func countTotalRecords() (int, error) {
	return countCollection("characters")
}

func countCollection(name string) (int, error) {
	iter := client.Collection(name).Documents(ctx)
	defer iter.Stop()

	count := 0
	for {
		_, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return 0, err
		}
		count++
	}
	return count, nil
}

// ---------------------------------------------------------------------------
// Migración por lotes con workers
// ---------------------------------------------------------------------------

func migrateInBatches() error {
	jobs := make(chan *OldCharacter, BATCH_SIZE)
	results := make(chan error, BATCH_SIZE)

	var wg sync.WaitGroup
	for i := 0; i < NUM_WORKERS; i++ {
		wg.Add(1)
		go worker(i, jobs, results, &wg)
	}

	// Productor
	go func() {
		iter := client.Collection("characters").Documents(ctx)
		defer iter.Stop()

		for {
			doc, err := iter.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				log.Printf("❌ Error leyendo documento: %v", err)
				continue
			}

			var oldChar OldCharacter
			if err := doc.DataTo(&oldChar); err != nil {
				log.Printf("❌ Error parseando documento %s: %v", doc.Ref.ID, err)
				continue
			}
			jobs <- &oldChar
		}
		close(jobs)
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	for err := range results {
		stateMutex.Lock()
		if err != nil {
			migrationState.FailedRecords++
			migrationState.ErrorLog = append(migrationState.ErrorLog, err.Error())
		} else {
			migrationState.MigratedRecords++
			if migrationState.MigratedRecords%CHECKPOINT_INT == 0 {
				saveCheckpoint()
				printProgress()
			}
		}
		stateMutex.Unlock()
	}

	return nil
}

func worker(id int, jobs <-chan *OldCharacter, results chan<- error, wg *sync.WaitGroup) {
	defer wg.Done()
	for oldChar := range jobs {
		results <- migrateCharacter(oldChar)
		time.Sleep(time.Duration(SLEEP_BETWEEN) * time.Millisecond)
	}
}

func migrateCharacter(old *OldCharacter) error {
	batch := client.Batch()

	// 1. Character base
	newChar := NewCharacter{
		ID:              old.ID,
		Name:            old.Name,
		Alias:           old.Alias,
		Species:         old.Species,
		Role:            old.Role,
		FirstAppearance: old.FirstAppearance,
		Notes:           old.Notes,
		MigratedAt:      time.Now(),
	}
	batch.Set(client.Collection("characters_new").Doc(old.ID), newChar)

	// 2. DevilFruit
	if old.DevilFruit != nil {
		fruitID := fmt.Sprintf("%s_fruit", old.ID)
		newFruit := NewDevilFruit{
			FruitID:     fruitID,
			CharacterID: old.ID,
			Name:        old.DevilFruit.Name,
			Type:        old.DevilFruit.Type,
			Description: old.DevilFruit.Description,
			CreatedAt:   time.Now(),
		}
		batch.Set(client.Collection("devilfruits").Doc(fruitID), newFruit)
	}

	// 3. HakiAbilities
	for i, haki := range old.HakiAbilities {
		charHakiID := fmt.Sprintf("%s_haki_%d", old.ID, i)
		hakiTypeID, ok := hakiTypesMap[haki.HakiType]
		if !ok {
			hakiTypeID = "unknown"
		}
		charHaki := CharacterHaki{
			ID:          charHakiID,
			CharacterID: old.ID,
			HakiTypeID:  hakiTypeID,
			Proficiency: haki.Proficiency,
			Awakened:    haki.Awakened,
			Notes:       haki.Notes,
			CreatedAt:   time.Now(),
		}
		batch.Set(client.Collection("character_haki").Doc(charHakiID), charHaki)
	}

	// 4. Abilities
	for i, ability := range old.Abilities {
		abilityID := fmt.Sprintf("%s_ability_%d", old.ID, i)
		newAbility := NewAbility{
			ID:          abilityID,
			CharacterID: old.ID,
			Type:        ability.Type,
			Notes:       ability.Notes,
			CreatedAt:   time.Now(),
		}
		batch.Set(client.Collection("abilities").Doc(abilityID), newAbility)
	}

	_, err := batch.Commit(ctx)
	if err != nil {
		return fmt.Errorf("error migrando %s: %v", old.ID, err)
	}
	return nil
}

// ---------------------------------------------------------------------------
// Checkpoint y progreso
// ---------------------------------------------------------------------------

func saveCheckpoint() {
	migrationState.LastCheckpoint = migrationState.MigratedRecords
	migrationState.LastUpdateTime = time.Now()
	migrationState.Status = "in_progress"

	data, _ := json.MarshalIndent(migrationState, "", "  ")
	os.WriteFile("migration_state.json", data, 0644)
}

func printProgress() {
	elapsed := time.Since(migrationState.StartTime)
	if elapsed.Seconds() == 0 {
		return
	}
	rate := float64(migrationState.MigratedRecords) / elapsed.Seconds()
	remaining := migrationState.TotalRecords - migrationState.MigratedRecords
	eta := time.Duration(float64(remaining)/rate) * time.Second
	progress := float64(migrationState.MigratedRecords) / float64(migrationState.TotalRecords) * 100

	fmt.Printf("\n📊 Progreso: %.1f%% (%d/%d)\n",
		progress, migrationState.MigratedRecords, migrationState.TotalRecords)
	fmt.Printf("   ⚡ Velocidad: %.1f registros/seg\n", rate)
	fmt.Printf("   ⏱️  ETA: %v\n", eta.Round(time.Second))
	fmt.Printf("   ❌ Errores: %d\n", migrationState.FailedRecords)
}

// ---------------------------------------------------------------------------
// Validación post-migración
// ---------------------------------------------------------------------------

func validateMigration() error {
	fmt.Println("\n🔍 Validando Migración...")

	oldCount, _ := countCollection("characters")
	newCount, _ := countCollection("characters_new")
	fmt.Printf("   Characters: %d → %d ", oldCount, newCount)
	if oldCount == newCount {
		fmt.Println("✅")
	} else {
		fmt.Println("⚠️  DIFERENCIA!")
	}

	devilFruitsCount, _ := countCollection("devilfruits")
	hakiRelCount, _ := countCollection("character_haki")
	abilitiesCount, _ := countCollection("abilities")

	fmt.Printf("   DevilFruits: %d ✅\n", devilFruitsCount)
	fmt.Printf("   Character_Haki: %d ✅\n", hakiRelCount)
	fmt.Printf("   Abilities: %d ✅\n", abilitiesCount)

	return nil
}

// ---------------------------------------------------------------------------
// Reporte final
// ---------------------------------------------------------------------------

func printFinalReport() {
	elapsed := time.Since(migrationState.StartTime)

	fmt.Println("\n" + repeatStr("=", 50))
	fmt.Println("🎉 MIGRACIÓN COMPLETADA")
	fmt.Println(repeatStr("=", 50))
	fmt.Printf("\n📊 Estadísticas Finales:\n")
	fmt.Printf("   Total registros: %d\n", migrationState.TotalRecords)
	fmt.Printf("   Migrados exitosamente: %d ✅\n", migrationState.MigratedRecords)
	fmt.Printf("   Fallidos: %d ❌\n", migrationState.FailedRecords)
	fmt.Printf("   Tiempo total: %v\n", elapsed.Round(time.Second))
	if elapsed.Seconds() > 0 {
		fmt.Printf("   Velocidad promedio: %.1f registros/seg\n\n",
			float64(migrationState.MigratedRecords)/elapsed.Seconds())
	}

	if len(migrationState.ErrorLog) > 0 {
		fmt.Println("⚠️  Errores encontrados:")
		for i, errMsg := range migrationState.ErrorLog {
			if i < 10 {
				fmt.Printf("   - %s\n", errMsg)
			}
		}
		if len(migrationState.ErrorLog) > 10 {
			fmt.Printf("   ... y %d errores más\n", len(migrationState.ErrorLog)-10)
		}
	}

	saveCheckpoint()
	fmt.Println("\n💾 Estado guardado en: migration_state.json")
}

func repeatStr(s string, n int) string {
	result := ""
	for i := 0; i < n; i++ {
		result += s
	}
	return result
}
