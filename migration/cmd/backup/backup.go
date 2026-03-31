package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

var (
	client *firestore.Client
	ctx    context.Context
)

func main() {
	startTime := time.Now()
	ctx = context.Background()

	// Conectar a Firestore
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

	fmt.Println("💾 BACKUP DE FIRESTORE")
	fmt.Println("======================")
	fmt.Println()

	// Crear carpeta de backup con timestamp
	backupDir := fmt.Sprintf("backup_%s", time.Now().Format("20060102_150405"))
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		log.Fatalf("❌ Error creando directorio de backup: %v", err)
	}

	fmt.Printf("📁 Directorio de backup: %s\n\n", backupDir)

	// Colecciones a respaldar
	collections := []string{"characters"}

	totalDocs := 0
	for _, collName := range collections {
		fmt.Printf("📦 Respaldando colección '%s'... ", collName)

		count, err := backupCollection(collName, backupDir)
		if err != nil {
			fmt.Printf("❌ Error: %v\n", err)
			continue
		}

		totalDocs += count
		fmt.Printf("✅ (%d documentos)\n", count)
	}

	duration := time.Since(startTime)
	fmt.Printf("\n✅ Backup completado exitosamente\n")
	fmt.Printf("   Total documentos: %d\n", totalDocs)
	fmt.Printf("   Tiempo: %v\n", duration.Round(time.Second))
	fmt.Printf("   Ubicación: %s/\n\n", backupDir)

	// Crear metadata file
	createMetadata(backupDir, collections, totalDocs, duration)

	fmt.Println("💡 Para restaurar este backup:")
	fmt.Printf("   go run restore.go %s\n", backupDir)
}

func backupCollection(collectionName, backupDir string) (int, error) {
	iter := client.Collection(collectionName).Documents(ctx)
	defer iter.Stop()

	documents := make([]map[string]interface{}, 0)

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return 0, err
		}

		var data map[string]interface{}
		if err := doc.DataTo(&data); err != nil {
			return 0, err
		}

		// Agregar metadata del documento
		data["_document_id"] = doc.Ref.ID
		data["_document_path"] = doc.Ref.Path

		documents = append(documents, data)
	}

	// Guardar en archivo JSON
	filename := filepath.Join(backupDir, fmt.Sprintf("%s.json", collectionName))
	data, err := json.MarshalIndent(documents, "", "  ")
	if err != nil {
		return 0, err
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return 0, err
	}

	return len(documents), nil
}

func createMetadata(backupDir string, collections []string, totalDocs int, duration time.Duration) {
	metadata := map[string]interface{}{
		"timestamp":        time.Now().Format(time.RFC3339),
		"collections":      collections,
		"total_documents":  totalDocs,
		"duration_seconds": duration.Seconds(),
		"firebase_project": os.Getenv("FIREBASE_PROJECT_ID"),
	}

	data, _ := json.MarshalIndent(metadata, "", "  ")
	filename := filepath.Join(backupDir, "backup_metadata.json")
	os.WriteFile(filename, data, 0644)
}
