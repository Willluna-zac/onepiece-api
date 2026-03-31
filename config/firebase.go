package config

import (
	"context"
	"log"
	"os"

	"cloud.google.com/go/firestore"
)

var FirestoreClient *firestore.Client

// InitFirebase inicializa la conexión con Firestore
func InitFirebase() {
	ctx := context.Background()

	// ID de tu proyecto de Firebase (puede configurarse con variable de entorno)
	projectID := os.Getenv("FIREBASE_PROJECT_ID")
	if projectID == "" {
		projectID = "conectionwdb" // Valor por defecto
	}

	// Crear cliente de Firestore
	// El SDK buscará automáticamente GOOGLE_APPLICATION_CREDENTIALS
	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("❌ Error al crear el cliente de Firestore: %v", err)
	}

	FirestoreClient = client
	log.Println("✅ Conexión exitosa a Firestore!")
}

// CloseFirebase cierra la conexión con Firestore
func CloseFirebase() {
	if FirestoreClient != nil {
		FirestoreClient.Close()
		log.Println("🔒 Conexión a Firestore cerrada")
	}
}
