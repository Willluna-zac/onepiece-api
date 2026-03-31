package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/google/uuid"
	"google.golang.org/api/iterator"
)

type migration struct {
	oldID string
	newID string
	data  map[string]interface{}
}

func isUUID(s string) bool {
	_, err := uuid.Parse(s)
	return err == nil
}

func main() {
	ctx := context.Background()

	projectID := os.Getenv("FIREBASE_PROJECT_ID")
	if projectID == "" {
		projectID = "conectionwdb"
	}

	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Error conectando a Firestore: %v", err)
	}
	defer client.Close()

	fmt.Println("Iniciando migracion de IDs a UUID...")

	var migrations []migration

	iter := client.Collection("characters_new").Documents(ctx)
	defer iter.Stop()

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatalf("Error leyendo characters_new: %v", err)
		}

		oldID := doc.Ref.ID
		if isUUID(oldID) {
			fmt.Printf("  SKIP %s (ya es UUID)\n", oldID)
			continue
		}

		newID := uuid.New().String()
		data := doc.Data()
		migrations = append(migrations, migration{oldID: oldID, newID: newID, data: data})
		fmt.Printf("  %s -> %s\n", oldID, newID)
	}

	if len(migrations) == 0 {
		fmt.Println("Todos los personajes ya tienen UUID.")
		return
	}

	fmt.Printf("\nMigrando %d personajes...\n", len(migrations))

	for _, m := range migrations {
		if err := migrateCharacter(ctx, client, m.oldID, m.newID, m.data); err != nil {
			log.Fatalf("Error migrando %s: %v", m.oldID, err)
		}
		fmt.Printf("  OK: %s -> %s\n", m.oldID, m.newID)
	}

	fmt.Printf("\nMigracion completada: %d personajes actualizados a UUID\n", len(migrations))
}

func migrateCharacter(ctx context.Context, client *firestore.Client, oldID, newID string, charData map[string]interface{}) error {
	now := time.Now()
	charData["id"] = newID
	charData["migratedAt"] = now

	// characters_new y characters
	batch := client.Batch()
	batch.Set(client.Collection("characters_new").Doc(newID), charData)
	batch.Delete(client.Collection("characters_new").Doc(oldID))
	batch.Set(client.Collection("characters").Doc(newID), charData)
	batch.Delete(client.Collection("characters").Doc(oldID))
	if _, err := batch.Commit(ctx); err != nil {
		return fmt.Errorf("batch characters: %v", err)
	}

	// devilfruits
	fruitIter := client.Collection("devilfruits").Where("character_id", "==", oldID).Documents(ctx)
	defer fruitIter.Stop()
	for {
		doc, err := fruitIter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return fmt.Errorf("leyendo devilfruits: %v", err)
		}
		fd := doc.Data()
		fd["character_id"] = newID
		fd["fruit_id"] = newID + "_fruit"
		fd["updatedAt"] = now
		b := client.Batch()
		b.Set(client.Collection("devilfruits").Doc(newID+"_fruit"), fd)
		b.Delete(doc.Ref)
		if _, err := b.Commit(ctx); err != nil {
			return fmt.Errorf("batch devilfruits: %v", err)
		}
	}

	// character_haki
	hakiIter := client.Collection("character_haki").Where("character_id", "==", oldID).Documents(ctx)
	defer hakiIter.Stop()
	i := 0
	for {
		doc, err := hakiIter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return fmt.Errorf("leyendo character_haki: %v", err)
		}
		hd := doc.Data()
		newHakiID := fmt.Sprintf("%s_haki_%d", newID, i)
		hd["character_id"] = newID
		hd["id"] = newHakiID
		hd["updatedAt"] = now
		b := client.Batch()
		b.Set(client.Collection("character_haki").Doc(newHakiID), hd)
		b.Delete(doc.Ref)
		if _, err := b.Commit(ctx); err != nil {
			return fmt.Errorf("batch character_haki: %v", err)
		}
		i++
	}

	// abilities
	abilityIter := client.Collection("abilities").Where("character_id", "==", oldID).Documents(ctx)
	defer abilityIter.Stop()
	j := 0
	for {
		doc, err := abilityIter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return fmt.Errorf("leyendo abilities: %v", err)
		}
		ad := doc.Data()
		newAbilityID := fmt.Sprintf("%s_ability_%d", newID, j)
		ad["character_id"] = newID
		ad["id"] = newAbilityID
		ad["updatedAt"] = now
		b := client.Batch()
		b.Set(client.Collection("abilities").Doc(newAbilityID), ad)
		b.Delete(doc.Ref)
		if _, err := b.Commit(ctx); err != nil {
			return fmt.Errorf("batch abilities: %v", err)
		}
		j++
	}

	return nil
}
