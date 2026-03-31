# 🔄 Estrategia de Migración: NoSQL → Normalizada

## 📋 Resumen Ejecutivo

**Objetivo**: Migrar de estructura NoSQL (documentos embebidos) a estructura relacional normalizada

**Impacto**: 
- Cambio en 4 colecciones
- Refactorización completa del modelo de datos
- Sin downtime para usuarios finales

**Tiempo estimado**: 4-6 horas en producción con 100K+ registros

---

## 🎯 Fases de Migración

### **FASE 1: Preparación (Pre-migración)**

#### 1.1 Backup Completo
```bash
# Exportar toda la base de datos actual
# Tiempo: ~30 min para 100K registros
```
- ✅ Backup automático de Firestore
- ✅ Backup manual adicional
- ✅ Verificar integridad del backup

#### 1.2 Crear Nuevas Colecciones
```
- devilfruits (nueva)
- hakitypes (nueva - catálogo)
- character_haki (nueva - relación)
- abilities (nueva)
- characters_new (temporal para validación)
```

#### 1.3 Poblar Catálogo HakiTypes
```
- Armament Haki
- Observation Haki  
- Conqueror Haki
```

---

### **FASE 2: Migración Dual-Write (Escritura Dual)**

⚠️ **Clave para zero-downtime**

#### 2.1 Implementar Adaptador de Datos
```
Aplicación escribe en:
├── Estructura VIEJA (backward compatibility)
└── Estructura NUEVA (forward compatibility)

Aplicación lee de:
└── Estructura VIEJA (por ahora)
```

#### 2.2 Desplegar Código con Dual-Write
- Sin cambios en lectura aún
- Todas las escrituras van a ambas estructuras
- Logs de auditoría activados

**Duración**: 24-48 horas mínimo
**Por qué**: Para acumular datos en la nueva estructura y verificar

---

### **FASE 3: Migración de Datos Históricos**

#### 3.1 Migración por Lotes (Batch Migration)

**Estrategia: Blue-Green con validación incremental**

```
Para cada lote de N personajes (N = 500):
  1. Leer documentos de 'characters' 
  2. Transformar estructura
  3. Escribir en nuevas colecciones
  4. Validar datos migrados
  5. Marcar como migrados (flag)
  6. Log de progreso
  7. Sleep 2s (para no saturar)
```

#### 3.2 Procesamiento Paralelo
- **Workers**: 5-10 workers paralelos
- **Rate limiting**: 500 writes/second (Firestore limit)
- **Checkpoints**: Cada 10K registros guardamos progreso
- **Reiniciable**: Si falla, continúa desde último checkpoint

#### 3.3 Manejo de Errores
```go
- Retry automático (3 intentos)
- Dead Letter Queue para registros problemáticos
- Alertas si tasa de error > 1%
- Logs detallados de cada error
```

**Duración**: 2-4 horas para 100K registros

---

### **FASE 4: Validación Exhaustiva**

#### 4.1 Validación de Integridad
```sql
✓ Verificar counts totales por colección
✓ Verificar relaciones (no FKs rotas)
✓ Verificar datos críticos (nombres, IDs)
✓ Comparar checksums entre viejo y nuevo
```

#### 4.2 Testing en Producción
- Pruebas de lectura con shadow traffic
- Comparar resultados viejo vs nuevo
- Métricas de performance
- **Criterio de éxito**: 99.99% coincidencia

**Duración**: 2-4 horas

---

### **FASE 5: Switchover (Cambio de Lectura)**

⚠️ **Momento crítico - reversible**

#### 5.1 Cambio Gradual (Canary Release)
```
Tráfico de lectura:
  5%  → Nueva estructura (15 min)
  25% → Nueva estructura (30 min)
  50% → Nueva estructura (1 hora)
  100% → Nueva estructura (2 horas después)
```

#### 5.2 Monitoreo Intensivo
- Latencia de queries
- Tasa de errores
- Alertas en tiempo real
- Dashboards activos

#### 5.3 Rollback Plan (si es necesario)
```
Si errores > 0.1% O latencia > +50%:
  → Revertir a estructura vieja INMEDIATAMENTE
  → Investigar
  → Fix
  → Reintentar
```

---

### **FASE 6: Limpieza Post-Migración**

#### 6.1 Periodo de Gracia (7-14 días)
- Mantener datos viejos como backup
- Dual-write continúa (opcional)
- Monitoreo activo

#### 6.2 Desactivar Estructura Vieja
```
1. Detener dual-write
2. Marcar colección vieja como deprecated
3. Crear backup final
4. Archivar datos viejos
```

#### 6.3 Eliminar Código Legacy
- Remover adaptador de dual-write
- Limpiar código de compatibilidad
- Actualizar documentación

---

## 📊 Diagrama de Flujo

```
┌─────────────────────────────────────────────────┐
│  FASE 1: Backup + Setup (30 min)               │
└─────────────────┬───────────────────────────────┘
                  │
                  ▼
┌─────────────────────────────────────────────────┐
│  FASE 2: Deploy Dual-Write (24-48h)            │
│  ✓ App escribe en ambas estructuras             │
│  ✓ App lee de estructura vieja                  │
└─────────────────┬───────────────────────────────┘
                  │
                  ▼
┌─────────────────────────────────────────────────┐
│  FASE 3: Migración Batch (2-4h)                │
│  ✓ 100K registros en lotes de 500              │
│  ✓ 5-10 workers paralelos                       │
│  ✓ Checkpoints cada 10K                         │
└─────────────────┬───────────────────────────────┘
                  │
                  ▼
┌─────────────────────────────────────────────────┐
│  FASE 4: Validación (2-4h)                     │
│  ✓ Verificar integridad                         │
│  ✓ Shadow testing                               │
│  ✓ Performance benchmarks                       │
└─────────────────┬───────────────────────────────┘
                  │
                  ▼
┌─────────────────────────────────────────────────┐
│  FASE 5: Switchover Gradual (4h)               │
│  ✓ 5% → 25% → 50% → 100%                       │
│  ✓ Monitoreo intensivo                          │
│  ✓ Rollback ready                               │
└─────────────────┬───────────────────────────────┘
                  │
                  ▼
┌─────────────────────────────────────────────────┐
│  FASE 6: Limpieza (7-14 días)                  │
│  ✓ Mantener backup temporal                     │
│  ✓ Eliminar código legacy                       │
│  ✓ Documentar lecciones aprendidas              │
└─────────────────────────────────────────────────┘
```

---

## 🔧 Herramientas Necesarias

### Scripts
1. `backup.go` - Backup completo
2. `migrate.go` - Migración batch
3. `validate.go` - Validación de datos
4. `rollback.go` - Plan de reversión
5. `dual_write_adapter.go` - Escritura dual

### Monitoreo
- Logs estructurados (JSON)
- Métricas en tiempo real
- Alertas automatizadas
- Dashboard de migración

---

## ⚠️ Riesgos y Mitigaciones

| Riesgo | Probabilidad | Impacto | Mitigación |
|--------|-------------|---------|------------|
| Pérdida de datos | Baja | Crítico | Backups múltiples + validación |
| Downtime no planeado | Media | Alto | Dual-write + rollback rápido |
| Inconsistencias | Media | Medio | Validación exhaustiva + período de gracia |
| Performance degradado | Media | Medio | Canary release + monitoreo |
| Bugs en código nuevo | Alta | Medio | Testing extensivo + rollback |

---

## ✅ Checklist de Ejecución

### Pre-migración
- [ ] Backup completo verificado
- [ ] Scripts probados en staging
- [ ] Equipo alertado y disponible
- [ ] Rollback plan documentado
- [ ] Ventana de mantenimiento programada (opcional)

### Durante migración
- [ ] Monitoreo activo
- [ ] Logs siendo revisados
- [ ] Métricas normales
- [ ] Comunicación con stakeholders

### Post-migración
- [ ] Validación completa ejecutada
- [ ] Performance dentro de SLAs
- [ ] Documentación actualizada
- [ ] Postmortem realizado

---

## 📞 Escalación

**Nivel 1**: Equipo de desarrollo (0-15 min)
**Nivel 2**: Lead Engineer (15-30 min)
**Nivel 3**: Rollback automático (si errores > umbral)

---

## 📚 Referencias

- [Firestore Best Practices](https://firebase.google.com/docs/firestore/best-practices)
- [Data Migration Patterns](https://martinfowler.com/articles/evodb.html)
- [Blue-Green Deployments](https://martinfowler.com/bliki/BlueGreenDeployment.html)
