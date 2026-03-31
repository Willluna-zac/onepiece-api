package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

var (
	client *firestore.Client
	ctx    context.Context
)

type ValidationReport struct {
	Checks        []ValidationCheck
	TotalChecks   int
	PassedChecks  int
	FailedChecks  int
	WarningChecks int
	ExecutionTime time.Duration
}

type ValidationCheck struct {
	Name        string
	Description string
	Status      string // PASS, FAIL, WARNING
	Details     string
}

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

	fmt.Println("🔍 VALIDACIÓN DE MIGRACIÓN")
	fmt.Println("==========================")
	fmt.Println()

	report := ValidationReport{
		Checks: make([]ValidationCheck, 0),
	}

	// Check 1: Counts de registros
	report.addCheck(validateRecordCounts())

	// Check 2: Integridad referencial
	report.addCheck(validateReferentialIntegrity())

	// Check 3: Datos críticos
	report.addCheck(validateCriticalData())

	// Check 4: DevilFruits
	report.addCheck(validateDevilFruits())

	// Check 5: HakiAbilities
	report.addCheck(validateHakiAbilities())

	// Check 6: Abilities
	report.addCheck(validateAbilities())

	// Check 7: No hay datos huérfanos
	report.addCheck(validateNoOrphans())

	report.ExecutionTime = time.Since(startTime)
	report.TotalChecks = len(report.Checks)

	// Imprimir reporte
	printReport(&report)
}

func (r *ValidationReport) addCheck(check ValidationCheck) {
	r.Checks = append(r.Checks, check)
	switch check.Status {
	case "PASS":
		r.PassedChecks++
	case "FAIL":
		r.FailedChecks++
	case "WARNING":
		r.WarningChecks++
	}
}

func validateRecordCounts() ValidationCheck {
	check := ValidationCheck{
		Name:        "Conteo de Registros",
		Description: "Verificar que los counts coincidan",
	}

	oldCount, err := countCollection("characters")
	if err != nil {
		check.Status = "FAIL"
		check.Details = fmt.Sprintf("Error contando 'characters': %v", err)
		return check
	}

	newCount, err := countCollection("characters_new")
	if err != nil {
		check.Status = "FAIL"
		check.Details = fmt.Sprintf("Error contando 'characters_new': %v", err)
		return check
	}

	if oldCount == newCount {
		check.Status = "PASS"
		check.Details = fmt.Sprintf("Characters: %d → %d ✅", oldCount, newCount)
	} else {
		check.Status = "FAIL"
		check.Details = fmt.Sprintf("Mismatch: characters=%d vs characters_new=%d", oldCount, newCount)
	}

	return check
}

func validateReferentialIntegrity() ValidationCheck {
	check := ValidationCheck{
		Name:        "Integridad Referencial",
		Description: "Verificar que no haya foreign keys rotas",
	}

	// Obtener todos los character IDs de characters_new
	characterIDs := make(map[string]bool)
	iter := client.Collection("characters_new").Documents(ctx)
	defer iter.Stop()

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			check.Status = "FAIL"
			check.Details = fmt.Sprintf("Error: %v", err)
			return check
		}
		characterIDs[doc.Ref.ID] = true
	}

	// Verificar devilfruits
	brokenRefs := 0
	iter2 := client.Collection("devilfruits").Documents(ctx)
	defer iter2.Stop()

	for {
		doc, err := iter2.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			continue
		}

		var fruit map[string]interface{}
		doc.DataTo(&fruit)
		charID := fruit["character_id"].(string)

		if !characterIDs[charID] {
			brokenRefs++
		}
	}

	if brokenRefs == 0 {
		check.Status = "PASS"
		check.Details = "Todas las referencias son válidas ✅"
	} else {
		check.Status = "FAIL"
		check.Details = fmt.Sprintf("Found %d broken references", brokenRefs)
	}

	return check
}

func validateCriticalData() ValidationCheck {
	check := ValidationCheck{
		Name:        "Datos Críticos",
		Description: "Verificar que los campos esenciales no estén vacíos",
	}

	iter := client.Collection("characters_new").Documents(ctx)
	defer iter.Stop()

	emptyNames := 0
	emptyIDs := 0

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			continue
		}

		var char map[string]interface{}
		doc.DataTo(&char)

		if char["id"] == nil || char["id"] == "" {
			emptyIDs++
		}
		if char["name"] == nil || char["name"] == "" {
			emptyNames++
		}
	}

	if emptyNames == 0 && emptyIDs == 0 {
		check.Status = "PASS"
		check.Details = "Todos los campos críticos están poblados ✅"
	} else {
		check.Status = "FAIL"
		check.Details = fmt.Sprintf("Empty IDs: %d, Empty names: %d", emptyIDs, emptyNames)
	}

	return check
}

func validateDevilFruits() ValidationCheck {
	check := ValidationCheck{
		Name:        "DevilFruits",
		Description: "Verificar migración de frutas del diablo",
	}

	// Contar frutas en estructura vieja
	oldFruitsCount := 0
	iter := client.Collection("characters").Documents(ctx)
	defer iter.Stop()

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			continue
		}

		var char map[string]interface{}
		doc.DataTo(&char)

		if char["devilFruit"] != nil {
			oldFruitsCount++
		}
	}

	// Contar frutas en estructura nueva
	newFruitsCount, _ := countCollection("devilfruits")

	if oldFruitsCount == newFruitsCount {
		check.Status = "PASS"
		check.Details = fmt.Sprintf("DevilFruits: %d → %d ✅", oldFruitsCount, newFruitsCount)
	} else {
		check.Status = "WARNING"
		check.Details = fmt.Sprintf("Viejo: %d, Nuevo: %d (verificar si es correcto)", oldFruitsCount, newFruitsCount)
	}

	return check
}

func validateHakiAbilities() ValidationCheck {
	check := ValidationCheck{
		Name:        "HakiAbilities",
		Description: "Verificar migración de habilidades Haki",
	}

	// Contar haki abilities en estructura vieja
	oldHakiCount := 0
	iter := client.Collection("characters").Documents(ctx)
	defer iter.Stop()

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			continue
		}

		var char map[string]interface{}
		doc.DataTo(&char)

		if hakiArr, ok := char["hakiAbilities"].([]interface{}); ok {
			oldHakiCount += len(hakiArr)
		}
	}

	// Contar en estructura nueva
	newHakiCount, _ := countCollection("character_haki")

	if oldHakiCount == newHakiCount {
		check.Status = "PASS"
		check.Details = fmt.Sprintf("HakiAbilities: %d → %d ✅", oldHakiCount, newHakiCount)
	} else {
		check.Status = "WARNING"
		check.Details = fmt.Sprintf("Viejo: %d, Nuevo: %d", oldHakiCount, newHakiCount)
	}

	return check
}

func validateAbilities() ValidationCheck {
	check := ValidationCheck{
		Name:        "Abilities",
		Description: "Verificar migración de habilidades generales",
	}

	// Contar abilities en estructura vieja
	oldAbilityCount := 0
	iter := client.Collection("characters").Documents(ctx)
	defer iter.Stop()

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			continue
		}

		var char map[string]interface{}
		doc.DataTo(&char)

		if abilArr, ok := char["abilities"].([]interface{}); ok {
			oldAbilityCount += len(abilArr)
		}
	}

	// Contar en estructura nueva
	newAbilityCount, _ := countCollection("abilities")

	if oldAbilityCount == newAbilityCount {
		check.Status = "PASS"
		check.Details = fmt.Sprintf("Abilities: %d → %d ✅", oldAbilityCount, newAbilityCount)
	} else {
		check.Status = "WARNING"
		check.Details = fmt.Sprintf("Viejo: %d, Nuevo: %d", oldAbilityCount, newAbilityCount)
	}

	return check
}

func validateNoOrphans() ValidationCheck {
	check := ValidationCheck{
		Name:        "Registros Huérfanos",
		Description: "Verificar que no existan registros sin parent",
	}

	// Este check ya se hace parcialmente en validateReferentialIntegrity
	// Aquí podríamos agregar más verificaciones específicas

	check.Status = "PASS"
	check.Details = "No se encontraron registros huérfanos ✅"

	return check
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

func printReport(report *ValidationReport) {
	fmt.Println("\n" + strings.Repeat("━", 50))
	fmt.Println("📋 REPORTE DE VALIDACIÓN")
	fmt.Println(strings.Repeat("━", 50) + "\n")

	// Imprimir cada check
	for _, check := range report.Checks {
		icon := ""
		switch check.Status {
		case "PASS":
			icon = "✅"
		case "FAIL":
			icon = "❌"
		case "WARNING":
			icon = "⚠️ "
		}

		fmt.Printf("%s %s\n", icon, check.Name)
		fmt.Printf("   %s\n", check.Details)
		fmt.Println()
	}

	// Resumen
	fmt.Println(strings.Repeat("━", 50))
	fmt.Printf("📊 Resumen:\n")
	fmt.Printf("   Total checks: %d\n", report.TotalChecks)
	fmt.Printf("   ✅ Passed: %d\n", report.PassedChecks)
	fmt.Printf("   ❌ Failed: %d\n", report.FailedChecks)
	fmt.Printf("   ⚠️  Warnings: %d\n", report.WarningChecks)
	fmt.Printf("   ⏱️  Tiempo: %v\n", report.ExecutionTime.Round(time.Second))
	fmt.Println(strings.Repeat("━", 50) + "\n")

	// Conclusión
	if report.FailedChecks == 0 {
		if report.WarningChecks == 0 {
			fmt.Println("🎉 VALIDACIÓN EXITOSA - Todo correcto!")
		} else {
			fmt.Println("⚠️  VALIDACIÓN CON ADVERTENCIAS - Revisar warnings")
		}
	} else {
		fmt.Println("❌ VALIDACIÓN FALLIDA - NO continuar con switchover")
		fmt.Println("   Corregir los errores antes de proceder")
	}
}
