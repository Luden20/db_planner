package main

import (
	"fmt"
	"regexp"
	"slices"
	"strings"

	"db_planner/utils"
)

type scriptColumn struct {
	Name        string
	Type        string
	Nullable    bool
	Description string
}

type scriptForeignKey struct {
	Name       string
	Columns    []string
	RefTable   string
	RefColumns []string
	Unique     bool
}

type scriptTable struct {
	Name        string
	Description string
	Columns     []scriptColumn
	PrimaryKey  []string
	ForeignKeys []scriptForeignKey
}

func generateCodeScript(exportData *schemaExportData, database string) (string, error) {
	if exportData == nil {
		return "", fmt.Errorf("esquema no cargado")
	}

	tablesByEntityID := make(map[int]*scriptTable)
	tablesByRelationID := make(map[int]*scriptTable)
	entityNamesByID := make(map[int]string)

	for _, entity := range exportData.Entities {
		table := buildScriptTable(entity)
		tablesByEntityID[entity.Id] = table
		entityNamesByID[entity.Id] = entity.Name
	}
	for _, item := range exportData.IntersectionEntities {
		table := buildScriptTable(item.Entity)
		tablesByEntityID[item.Entity.Id] = table
		tablesByRelationID[item.RelationID] = table
		entityNamesByID[item.Entity.Id] = item.Entity.Name
	}

	orderedTables := make([]*scriptTable, 0, len(exportData.Entities)+len(exportData.IntersectionEntities))
	for _, entity := range exportData.Entities {
		orderedTables = append(orderedTables, tablesByEntityID[entity.Id])
	}
	for _, item := range exportData.IntersectionEntities {
		orderedTables = append(orderedTables, tablesByEntityID[item.Entity.Id])
	}

	for _, table := range orderedTables {
		ensurePrimaryKey(table, database)
	}

	relationDocs := make([]string, 0, len(exportData.Relations))
	for _, relation := range exportData.Relations {
		docLine, err := applyRelationToScriptModel(relation, tablesByEntityID, tablesByRelationID, entityNamesByID, database)
		if err != nil {
			return "", err
		}
		if strings.TrimSpace(docLine) != "" {
			relationDocs = append(relationDocs, docLine)
		}
	}

	var builder strings.Builder
	builder.WriteString("-- Script generado localmente por db_planner\n")
	builder.WriteString("-- Incluye DDL, relaciones, notas de documentacion y plantillas INSERT.\n\n")
	builder.WriteString("-- Resumen\n")
	builder.WriteString(fmt.Sprintf("-- Base de datos destino: %s\n", database))
	builder.WriteString(fmt.Sprintf("-- Tablas: %d\n", len(orderedTables)))
	if len(relationDocs) > 0 {
		builder.WriteString("-- Relaciones:\n")
		for _, line := range relationDocs {
			builder.WriteString("-- - ")
			builder.WriteString(line)
			builder.WriteString("\n")
		}
	}
	builder.WriteString("\n")

	for _, table := range orderedTables {
		writeTableDDL(&builder, table, database)
		builder.WriteString("\n")
	}

	writeInsertTemplates(&builder, orderedTables)
	return strings.TrimSpace(builder.String()), nil
}

func buildScriptTable(entity utils.Entity) *scriptTable {
	table := &scriptTable{
		Name:        sanitizeSQLName(entity.Name, "table"),
		Description: strings.TrimSpace(entity.Description),
		Columns:     make([]scriptColumn, 0, len(entity.Attributes)),
		PrimaryKey:  make([]string, 0),
		ForeignKeys: make([]scriptForeignKey, 0),
	}

	for _, attribute := range entity.Attributes {
		columnName := sanitizeSQLName(attribute.Name, "column")
		columnType := strings.TrimSpace(attribute.Type)
		if columnType == "" {
			columnType = "string"
		}
		addColumn(table, scriptColumn{
			Name:        columnName,
			Type:        columnType,
			Nullable:    attribute.Optional,
			Description: joinAttributeDocumentation(attribute),
		})
		if attribute.KeyType == utils.AttributeKeyPK && !slices.Contains(table.PrimaryKey, columnName) {
			table.PrimaryKey = append(table.PrimaryKey, columnName)
		}
	}

	return table
}

func ensurePrimaryKey(table *scriptTable, database string) {
	if table == nil {
		return
	}
	if len(table.Columns) == 0 {
		addColumn(table, scriptColumn{
			Name:        "id",
			Type:        "int",
			Nullable:    false,
			Description: "Clave primaria generada automaticamente por falta de atributos.",
		})
	}
	if len(table.PrimaryKey) > 0 {
		return
	}
	if !hasColumn(table, "id") {
		addColumn(table, scriptColumn{
			Name:        "id",
			Type:        "int",
			Nullable:    false,
			Description: "Clave primaria generada automaticamente por falta de PK definida.",
		})
	}
	table.PrimaryKey = []string{"id"}
}

func applyRelationToScriptModel(
	relation utils.Relation,
	tablesByEntityID map[int]*scriptTable,
	tablesByRelationID map[int]*scriptTable,
	entityNamesByID map[int]string,
	database string,
) (string, error) {
	switch relation.Relation {
	case utils.RelationTypeNN:
		intersectionTable := tablesByRelationID[relation.Id]
		if intersectionTable == nil {
			return describeRelation(relation, entityNamesByID), nil
		}
		if err := attachForeignKey(tablesByEntityID[relation.IdEntity1], intersectionTable, false, database); err != nil {
			return "", err
		}
		if err := attachForeignKey(tablesByEntityID[relation.IdEntity2], intersectionTable, false, database); err != nil {
			return "", err
		}
		if len(intersectionTable.PrimaryKey) == 1 && intersectionTable.PrimaryKey[0] == "id" {
			intersectionTable.PrimaryKey = intersectionForeignKeyColumns(intersectionTable)
			if len(intersectionTable.PrimaryKey) == 0 {
				intersectionTable.PrimaryKey = []string{"id"}
			}
		}
	case utils.RelationType1N:
		return describeRelation(relation, entityNamesByID), attachForeignKey(tablesByEntityID[relation.IdEntity1], tablesByEntityID[relation.IdEntity2], false, database)
	case utils.RelationTypeN1:
		return describeRelation(relation, entityNamesByID), attachForeignKey(tablesByEntityID[relation.IdEntity2], tablesByEntityID[relation.IdEntity1], false, database)
	case utils.RelationType1Np:
		return describeRelation(relation, entityNamesByID), attachForeignKey(tablesByEntityID[relation.IdEntity1], tablesByEntityID[relation.IdEntity2], true, database)
	case utils.RelationTypeNp1:
		return describeRelation(relation, entityNamesByID), attachForeignKey(tablesByEntityID[relation.IdEntity2], tablesByEntityID[relation.IdEntity1], true, database)
	case utils.RelationType11:
		return describeRelation(relation, entityNamesByID), attachUniqueForeignKey(tablesByEntityID[relation.IdEntity1], tablesByEntityID[relation.IdEntity2], database)
	}
	return describeRelation(relation, entityNamesByID), nil
}

func attachForeignKey(parent *scriptTable, child *scriptTable, identifying bool, database string) error {
	if parent == nil || child == nil {
		return nil
	}
	if len(parent.PrimaryKey) == 0 {
		ensurePrimaryKey(parent, database)
	}

	childColumns := make([]string, 0, len(parent.PrimaryKey))
	parentColumns := make([]string, 0, len(parent.PrimaryKey))
	for _, parentPK := range parent.PrimaryKey {
		parentColumn := findColumn(parent, parentPK)
		if parentColumn == nil {
			return fmt.Errorf("pk no encontrada en %s", parent.Name)
		}
		childColumnName := sanitizeSQLName(parent.Name+"_"+parentPK, "column")
		addColumn(child, scriptColumn{
			Name:        childColumnName,
			Type:        parentColumn.Type,
			Nullable:    !identifying,
			Description: fmt.Sprintf("FK hacia %s(%s).", parent.Name, parentPK),
		})
		childColumns = append(childColumns, childColumnName)
		parentColumns = append(parentColumns, parentPK)
		if identifying && !slices.Contains(child.PrimaryKey, childColumnName) {
			child.PrimaryKey = append(child.PrimaryKey, childColumnName)
		}
	}

	addForeignKey(child, scriptForeignKey{
		Name:       sanitizeSQLName("fk_"+child.Name+"_"+parent.Name, "fk"),
		Columns:    childColumns,
		RefTable:   parent.Name,
		RefColumns: parentColumns,
	})
	return nil
}

func attachUniqueForeignKey(parent *scriptTable, child *scriptTable, database string) error {
	if err := attachForeignKey(parent, child, false, database); err != nil {
		return err
	}
	if len(child.ForeignKeys) > 0 {
		child.ForeignKeys[len(child.ForeignKeys)-1].Unique = true
	}
	return nil
}

func addColumn(table *scriptTable, column scriptColumn) {
	if table == nil {
		return
	}
	if existing := findColumn(table, column.Name); existing != nil {
		if existing.Type == "" {
			existing.Type = column.Type
		}
		if strings.TrimSpace(existing.Description) == "" {
			existing.Description = column.Description
		}
		if !column.Nullable {
			existing.Nullable = false
		}
		return
	}
	table.Columns = append(table.Columns, column)
}

func addForeignKey(table *scriptTable, fk scriptForeignKey) {
	for _, existing := range table.ForeignKeys {
		if existing.Name == fk.Name {
			return
		}
	}
	table.ForeignKeys = append(table.ForeignKeys, fk)
}

func hasColumn(table *scriptTable, columnName string) bool {
	return findColumn(table, columnName) != nil
}

func findColumn(table *scriptTable, columnName string) *scriptColumn {
	if table == nil {
		return nil
	}
	for idx := range table.Columns {
		if table.Columns[idx].Name == columnName {
			return &table.Columns[idx]
		}
	}
	return nil
}

func intersectionForeignKeyColumns(table *scriptTable) []string {
	columns := make([]string, 0)
	for _, fk := range table.ForeignKeys {
		for _, col := range fk.Columns {
			if !slices.Contains(columns, col) {
				columns = append(columns, col)
			}
		}
	}
	return columns
}

func writeTableDDL(builder *strings.Builder, table *scriptTable, database string) {
	builder.WriteString("-- ------------------------------------------------------------------\n")
	builder.WriteString("-- Tabla: ")
	builder.WriteString(table.Name)
	builder.WriteString("\n")
	if table.Description != "" {
		builder.WriteString("-- ")
		builder.WriteString(table.Description)
		builder.WriteString("\n")
	}
	for _, column := range table.Columns {
		if strings.TrimSpace(column.Description) == "" {
			continue
		}
		builder.WriteString("--   ")
		builder.WriteString(column.Name)
		builder.WriteString(": ")
		builder.WriteString(column.Description)
		builder.WriteString("\n")
	}
	builder.WriteString("CREATE TABLE ")
	builder.WriteString(table.Name)
	builder.WriteString(" (\n")

	lines := make([]string, 0, len(table.Columns)+1)
	for _, column := range table.Columns {
		line := fmt.Sprintf("  %s %s", column.Name, mapAttributeType(column.Type, database))
		if !column.Nullable {
			line += " NOT NULL"
		}
		lines = append(lines, line)
	}
	if len(table.PrimaryKey) > 0 {
		lines = append(lines, fmt.Sprintf("  CONSTRAINT %s PRIMARY KEY (%s)",
			sanitizeSQLName("pk_"+table.Name, "pk"),
			strings.Join(table.PrimaryKey, ", "),
		))
	}
	builder.WriteString(strings.Join(lines, ",\n"))
	builder.WriteString("\n);\n")

	for _, fk := range table.ForeignKeys {
		builder.WriteString("ALTER TABLE ")
		builder.WriteString(table.Name)
		builder.WriteString(" ADD CONSTRAINT ")
		builder.WriteString(fk.Name)
		builder.WriteString(" FOREIGN KEY (")
		builder.WriteString(strings.Join(fk.Columns, ", "))
		builder.WriteString(") REFERENCES ")
		builder.WriteString(fk.RefTable)
		builder.WriteString(" (")
		builder.WriteString(strings.Join(fk.RefColumns, ", "))
		builder.WriteString(")")
		if fk.Unique {
			builder.WriteString(";\n")
			builder.WriteString("CREATE UNIQUE INDEX ")
			builder.WriteString(sanitizeSQLName("ux_"+table.Name+"_"+strings.Join(fk.Columns, "_"), "ux"))
			builder.WriteString(" ON ")
			builder.WriteString(table.Name)
			builder.WriteString(" (")
			builder.WriteString(strings.Join(fk.Columns, ", "))
			builder.WriteString(")")
		}
		builder.WriteString(";\n")
	}

	if strings.EqualFold(database, "PostgreSQL") {
		writePostgresComments(builder, table)
	}
}

func writePostgresComments(builder *strings.Builder, table *scriptTable) {
	if table.Description != "" {
		builder.WriteString("COMMENT ON TABLE ")
		builder.WriteString(table.Name)
		builder.WriteString(" IS ")
		builder.WriteString(quoteSQLString(table.Description))
		builder.WriteString(";\n")
	}
	for _, column := range table.Columns {
		if strings.TrimSpace(column.Description) == "" {
			continue
		}
		builder.WriteString("COMMENT ON COLUMN ")
		builder.WriteString(table.Name)
		builder.WriteString(".")
		builder.WriteString(column.Name)
		builder.WriteString(" IS ")
		builder.WriteString(quoteSQLString(column.Description))
		builder.WriteString(";\n")
	}
}

func writeInsertTemplates(builder *strings.Builder, tables []*scriptTable) {
	builder.WriteString("-- ------------------------------------------------------------------\n")
	builder.WriteString("-- Plantillas INSERT\n\n")
	for _, table := range tables {
		builder.WriteString("-- INSERT template: ")
		builder.WriteString(table.Name)
		builder.WriteString("\n")
		builder.WriteString("-- INSERT INTO ")
		builder.WriteString(table.Name)
		builder.WriteString(" (\n")
		for idx, column := range table.Columns {
			builder.WriteString("--   ")
			builder.WriteString(column.Name)
			if idx < len(table.Columns)-1 {
				builder.WriteString(",")
			}
			builder.WriteString("\n")
		}
		builder.WriteString("-- ) VALUES (\n")
		for idx, column := range table.Columns {
			builder.WriteString("--   /* ")
			builder.WriteString(column.Name)
			builder.WriteString(" */")
			if idx < len(table.Columns)-1 {
				builder.WriteString(",")
			}
			builder.WriteString("\n")
		}
		builder.WriteString("-- );\n\n")
	}
}

func describeRelation(relation utils.Relation, entityNamesByID map[int]string) string {
	left := entityNamesByID[relation.IdEntity1]
	right := entityNamesByID[relation.IdEntity2]
	if left == "" {
		left = fmt.Sprintf("Entidad%d", relation.IdEntity1)
	}
	if right == "" {
		right = fmt.Sprintf("Entidad%d", relation.IdEntity2)
	}
	return fmt.Sprintf("%s %s %s", left, relation.Relation, right)
}

func joinAttributeDocumentation(attribute utils.Attribute) string {
	parts := make([]string, 0, 3)
	if strings.TrimSpace(attribute.Description) != "" {
		parts = append(parts, strings.TrimSpace(attribute.Description))
	}
	if len(attribute.Domain) > 0 {
		parts = append(parts, "Dominio: "+strings.Join(attribute.Domain, ", "))
	}
	if attribute.KeyType == utils.AttributeKeyPK {
		parts = append(parts, "Clave primaria.")
	}
	return strings.Join(parts, " ")
}

func mapAttributeType(raw string, database string) string {
	value := strings.ToLower(strings.TrimSpace(raw))
	if value == "" || value == strings.ToLower("Por definir") {
		value = "string"
	}
	switch {
	case strings.Contains(value, "bool"):
		switch strings.ToLower(database) {
		case "sql server":
			return "BIT"
		case "sqlite":
			return "INTEGER"
		default:
			return "BOOLEAN"
		}
	case strings.Contains(value, "date") && strings.Contains(value, "time"):
		switch strings.ToLower(database) {
		case "sql server":
			return "DATETIME2"
		case "sqlite":
			return "TEXT"
		case "mysql":
			return "DATETIME"
		default:
			return "TIMESTAMP"
		}
	case strings.Contains(value, "date"):
		if strings.EqualFold(database, "SQLite") {
			return "TEXT"
		}
		return "DATE"
	case strings.Contains(value, "decimal"), strings.Contains(value, "numeric"), strings.Contains(value, "money"):
		return "DECIMAL(10,2)"
	case strings.Contains(value, "float"), strings.Contains(value, "double"), strings.Contains(value, "real"):
		if strings.EqualFold(database, "SQLite") {
			return "REAL"
		}
		return "DOUBLE PRECISION"
	case strings.Contains(value, "int"), strings.Contains(value, "number"):
		return "INTEGER"
	case strings.Contains(value, "char"), strings.Contains(value, "text"), strings.Contains(value, "string"), strings.Contains(value, "uuid"):
		switch strings.ToLower(database) {
		case "sql server":
			return "NVARCHAR(255)"
		default:
			return "VARCHAR(255)"
		}
	default:
		switch strings.ToLower(database) {
		case "sql server":
			return "NVARCHAR(255)"
		case "sqlite":
			return "TEXT"
		default:
			return "VARCHAR(255)"
		}
	}
}

func quoteSQLString(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "''") + "'"
}

var sqlNameCleaner = regexp.MustCompile(`[^a-z0-9_]+`)

func sanitizeSQLName(value string, fallback string) string {
	normalized := strings.ToLower(strings.TrimSpace(value))
	normalized = strings.ReplaceAll(normalized, "-", "_")
	normalized = strings.ReplaceAll(normalized, " ", "_")
	normalized = sqlNameCleaner.ReplaceAllString(normalized, "_")
	normalized = strings.Trim(normalized, "_")
	if normalized == "" {
		normalized = fallback
	}
	if normalized[0] >= '0' && normalized[0] <= '9' {
		normalized = fallback + "_" + normalized
	}
	return normalized
}
