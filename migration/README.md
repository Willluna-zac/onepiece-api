# 🔄 Scripts de Migración

Este directorio contiene todos los scripts necesarios para migrar de una estructura NoSQL (documentos embebidos) a una estructura relacional normalizada.

## 📁 Archivos

### Documentación
- **MIGRATION_STRATEGY.md** - Estrategia completa de migración, fases, riesgos y mitigaciones
- **ANALYSIS.md** - Comparación detallada entre estructura actual vs propuesta

### Scripts Ejecutables (cmd/)

Cada script está en su propio subdirectorio para evitar conflictos de compilación.

| Script | Propósito | Cuándo usar |
|--------|-----------|-------------|
| `cmd/backup/backup.go` | Backup completo de Firestore | ANTES de cualquier migración |
| `cmd/migrate/migrate.go` | Migración por lotes con checkpoints | Durante la migración |  
| `cmd/validate/validate.go` | Validación exhaustiva de datos migrados | Después de la migración |
| `cmd/rollback/rollback.go` | Reversar migración (eliminar estructura nueva) | Si algo sale mal |

---

## 🚀 Guía de Uso

### 1️⃣ PREPARACIÓN (Pre-Migración)

#### Paso 1: Backup Completo
```bash
# Crear backup de seguridad
export GOOGLE_APPLICATION_CREDENTIALS=/ruta/a/credentials.json
export FIREBASE_PROJECT_ID=conectionwdb

cd migration
go run cmd/backup/backup.go
```

**Salida esperada:**
```
💾 BACKUP DE FIRESTORE
======================

📁 Directorio de backup: backup_20260223_143052

📦 Respaldando colección 'characters'... ✅ (3 documentos)

✅ Backup completado exitosamente
   Total documentos: 3
   Tiempo: 2s
   Ubicación: backup_20260223_143052/
```

**Resultado:** Se crea carpeta `backup_YYYYMMDD_HHMMSS/` con todos los datos.

---

### 2️⃣ MIGRACIÓN (Datos Históricos)

#### Paso 2: Ejecutar Migración
```bash
go run cmd/migrate/migrate.go
```

**Características:**
- ✅ Migración por lotes (500 registros)
- ✅ Procesamiento paralelo (5 workers)
- ✅ Checkpoints automáticos cada 10K
- ✅ Reiniciable si falla
- ✅ Rate limiting (no satura Firestore)
- ✅ Logs detallados

**Salida esperada:**
```
🚀 Iniciando Migración de Datos
================================

📋 FASE 1: Configuración Inicial
✅ Nuevo estado de migración inicializado

📋 FASE 2: Crear Catálogo HakiTypes
   ✅ HakiType 'Armament' creado
   ✅ HakiType 'Observation' creado
   ✅ HakiType 'Conqueror' creado

📋 FASE 3: Contando Registros Totales
✅ Total de personajes a migrar: 100000

📋 FASE 4: Migración por Lotes
   Batch size: 500
   Workers: 5
   Checkpoint cada: 10000 registros

📊 Progreso: 10.0% (10000/100000)
   ⚡ Velocidad: 125.3 registros/seg
   ⏱️  ETA: 12m5s
   ❌ Errores: 0

[... continúa ...]

==================================================
🎉 MIGRACIÓN COMPLETADA
==================================================

📊 Estadísticas Finales:
   Total registros: 100000
   Migrados exitosamente: 99998 ✅
   Fallidos: 2 ❌
   Tiempo total: 13m24s
   Velocidad promedio: 124.5 registros/seg

💾 Estado guardado en: migration_state.json
```

**Resultado:** 
- Colecciones nuevas pobladas
- Archivo `migration_state.json` con el progreso
- Si falla, se puede reiniciar y continúa desde el último checkpoint

---

### 3️⃣ VALIDACIÓN

#### Paso 3: Validar Datos Migrados
```bash
go run cmd/validate/validate.go
```

**Qué valida:**
- ✅ Counts de registros (characters vs characters_new)
- ✅ Integridad referencial (no FKs rotas)
- ✅ Datos críticos (campos no vacíos)
- ✅ DevilFruits migradas correctamente
- ✅ HakiAbilities migradas correctamente
- ✅ Abilities migradas correctamente
- ✅ No hay registros huérfanos

**Salida esperada:**
```
🔍 VALIDACIÓN DE MIGRACIÓN
==========================

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
📋 REPORTE DE VALIDACIÓN
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

✅ Conteo de Registros
   Characters: 100000 → 100000 ✅

✅ Integridad Referencial
   Todas las referencias son válidas ✅

✅ Datos Críticos
   Todos los campos críticos están poblados ✅

✅ DevilFruits
   DevilFruits: 5234 → 5234 ✅

⚠️  HakiAbilities
   Viejo: 8521, Nuevo: 8520

✅ Abilities
   Abilities: 15432 → 15432 ✅

✅ Registros Huérfanos
   No se encontraron registros huérfanos ✅

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
📊 Resumen:
   Total checks: 7
   ✅ Passed: 6
   ❌ Failed: 0
   ⚠️  Warnings: 1
   ⏱️  Tiempo: 8s
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

⚠️  VALIDACIÓN CON ADVERTENCIAS - Revisar warnings
```

**Decisión:**
- ✅ **0 errores** → Continuar a switchover
- ⚠️ **Warnings** → Investigar pero posiblemente OK
- ❌ **Errores** → NO continuar, corregir primero

---

### 4️⃣ ROLLBACK (Si es necesario)

#### Paso 4: Revertir Migración
```bash
go run cmd/rollback/rollback.go
```

**⚠️ ADVERTENCIA:** Elimina TODAS las colecciones nuevas.

```
🔄 INICIANDO ROLLBACK
====================

⚠️  ADVERTENCIA: Esta operación eliminará TODAS las colecciones nuevas
    y restaurará el sistema a usar las colecciones antiguas.

¿Está seguro que desea continuar? (escriba 'SI' para confirmar): SI

🗑️  FASE 1: Eliminando colecciones nuevas
   Eliminando 'characters_new'... (100000 docs) ✅
   Eliminando 'devilfruits'... (5234 docs) ✅
   Eliminando 'character_haki'... (8521 docs) ✅
   Eliminando 'abilities'... (15432 docs) ✅
   Eliminando 'hakitypes'... (3 docs) ✅

✅ ROLLBACK COMPLETADO

💡 Acciones recomendadas:
   1. Reiniciar los servicios/aplicación
   2. Verificar que la aplicación usa 'characters' (colección vieja)
   3. Revisar logs de aplicación
   4. Analizar causa del rollback antes de reintentar
```

---

## 📊 Estructura de Datos

### ANTES (NoSQL - Embebido)
```json
{
  "id": "luffy",
  "name": "Monkey D. Luffy",
  "devilFruit": {
    "name": "Gomu Gomu no Mi",
    "type": "Zoan"
  },
  "hakiAbilities": [
    {"hakiType": "Armament", "proficiency": "Advanced"}
  ],
  "abilities": [
    {"type": "Gear 5", "notes": "Awakened form"}
  ]
}
```

### DESPUÉS (SQL - Normalizado)

**characters_new:**
```json
{
  "id": "luffy",
  "name": "Monkey D. Luffy",
  "alias": "Mugiwara"
}
```

**devilfruits:**
```json
{
  "fruit_id": "luffy_fruit",
  "character_id": "luffy",
  "name": "Gomu Gomu no Mi",
  "type": "Zoan"
}
```

**character_haki:**
```json
{
  "id": "luffy_haki_0",
  "character_id": "luffy",
  "haki_type_id": "armament",
  "proficiency": "Advanced"
}
```

**abilities:**
```json
{
  "id": "luffy_ability_0",
  "character_id": "luffy",
  "type": "Gear 5",
  "notes": "Awakened form"
}
```

---

## 🔧 Configuración

### Variables de Entorno Requeridas
```bash
export GOOGLE_APPLICATION_CREDENTIALS=/ruta/a/credentials.json
export FIREBASE_PROJECT_ID=conectionwdb
```

### Parámetros de Migración (migrate.go)
```go
BATCH_SIZE     = 500   // Registros por lote
NUM_WORKERS    = 5     // Workers paralelos
CHECKPOINT_INT = 10000 // Checkpoint cada N registros
SLEEP_BETWEEN  = 2     // Milisegundos entre lotes
```

**Ajustar según:**
- Poder de cómputo disponible
- Límites de Firestore (500 writes/sec)
- Tamaño del dataset

---

## ⚡ Troubleshooting

### Problema: Migración muy lenta
**Solución:** Aumentar `NUM_WORKERS` (máx 10) o `BATCH_SIZE` (máx 1000)

### Problema: Errores de rate limiting
**Solución:** Aumentar `SLEEP_BETWEEN` o reducir `NUM_WORKERS`

### Problema: Migración se detiene
**Solución:** Simplemente vuelve a ejecutar `go run migrate.go` - continúa desde el checkpoint

### Problema: Validación falla
**Solución:** Revisar `migration_state.json` para ver errores específicos, corregir y re-migrar esos registros

---

## 📝 Notas Importantes

1. **Siempre hacer backup antes de migrar**
2. **Validar exhaustivamente antes de switchover**
3. **La migración es reiniciable** - usa checkpoints
4. **Los datos viejos NO se modifican** - se crean nuevas colecciones
5. **Rollback es rápido** - solo elimina colecciones nuevas
6. **Para producción:** Probar primero en staging con datos realistas

---

## 🎯 Checklist de Migración

- [ ] Backup completo realizado y verificado
- [ ] Script de migración probado en staging
- [ ] Equipo notificado y disponible
- [ ] Ventana de mantenimiento programada (opcional)
- [ ] Migración ejecutada exitosamente
- [ ] Validación pasada con 0 errores críticos
- [ ] Plan de rollback documentado y probado
- [ ] Monitoreo configurado
- [ ] Switchover gradual planificado (Canary: 5% → 25% → 50% → 100%)

---

## 🆘 Soporte

En caso de problemas críticos durante la migración:

1. **NO entrar en pánico**
2. Detener la migración (Ctrl+C)
3. Revisar `migration_state.json` para ver el estado
4. Si es necesario, ejecutar `go run rollback.go`
5. Analizar logs y errores
6. Corregir el problema
7. Reintentar desde el checkpoint

---

**Última actualización:** Febrero 2026
