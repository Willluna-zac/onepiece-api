package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

var (
	client *firestore.Client
	ctx    context.Context
)

func main() {
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

	fmt.Println("🔄 INICIANDO ROLLBACK")
	fmt.Println("====================")
	fmt.Println()
	fmt.Println("⚠️  ADVERTENCIA: Esta operación eliminará TODAS las colecciones nuevas")
	fmt.Println("    y restaurará el sistema a usar las colecciones antiguas.")
	fmt.Println()

	// Confirmar rollback
	fmt.Print("¿Está seguro que desea continuar? (escriba 'SI' para confirmar): ")
	var confirmation string
	fmt.Scanln(&confirmation)

	if confirmation != "SI" {
		fmt.Println("\n❌ Rollback cancelado")
		return
	}

	fmt.Println("\n🗑️  FASE 1: Eliminando colecciones nuevas")

	// Eliminar colecciones nuevas
	collections := []string{
		"characters_new",
		"devilfruits",
		"character_haki",
		"abilities",
		"hakitypes",
	}

	for _, coll := range collections {
		fmt.Printf("   Eliminando '%s'... ", coll)
		if err := deleteCollection(coll); err != nil {
			fmt.Printf("❌ Error: %v\n", err)
		} else {
			fmt.Println("✅")
		}
	}

	fmt.Println("\n✅ ROLLBACK COMPLETADO")
	fmt.Println("\n💡 Acciones recomendadas:")
	fmt.Println("   1. Reiniciar los servicios/aplicación")
	fmt.Println("   2. Verificar que la aplicación usa 'characters' (colección vieja)")
	fmt.Println("   3. Revisar logs de aplicación")
	fmt.Println("   4. Analizar causa del rollback antes de reintentar")
}

func deleteCollection(collectionName string) error {
	ref := client.Collection(collectionName)

	// Eliminar en lotes
	bulkwriter := client.BulkWriter(ctx)

	iter := ref.Documents(ctx)
	defer iter.Stop()

	numDeleted := 0
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return err
		}

		_, err = bulkwriter.Delete(doc.Ref)
		if err != nil {
			return err
		}
		numDeleted++
	}

	bulkwriter.End()

	if numDeleted > 0 {
		fmt.Printf("(%d docs) ", numDeleted)
	}

	return nil
}
