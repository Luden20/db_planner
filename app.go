package main

import (
	"bytes"
	"context"
	"db_planner/utils"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// Startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}
func (a *App) CreateNew(path string, name string) (*utils.DbProject, error) {
	prj, err := utils.CreateNew(path, name)
	if err != nil {
		return nil, err
	}
	return prj, nil
}
func (a *App) EjectProject() error {
	err := utils.Eject()
	if err != nil {
		return err
	}
	return nil
}
func (a *App) Save() error {
	err := utils.SaveChanges()
	if err != nil {
		return err
	}
	return nil
}
func (a *App) OpenPath(path string) (*utils.DbProject, error) {
	prj, err := utils.LoadProjectFromJson(path)
	if err != nil {
		return nil, err
	}
	return prj, nil
}
func (a *App) CreateExcelPath() (string, error) {
	return a.createExcelPath("")
}

func (a *App) CreateExcelPathWithoutRelations() (string, error) {
	return a.createExcelPath("-sin-relaciones")
}

func (a *App) createExcelPath(suffix string) (string, error) {
	prj, err := utils.GetActualProject()
	if err != nil {
		return "", err
	}
	defaultName := sanitizeFilename(prj.Name) + suffix + ".xlsx"
	return runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "Exportar a Excel",
		DefaultFilename: defaultName,
		Filters: []runtime.FileFilter{
			{DisplayName: "Excel", Pattern: "*.xlsx"},
			{DisplayName: "Todos", Pattern: "*"},
		},
	})
}
func (a *App) ExportToExcel(path string) error {
	prj, err := utils.GetActualProject()
	if err != nil {
		return err
	}
	if strings.TrimSpace(path) == "" {
		return fmt.Errorf("ruta de exportación no válida")
	}
	err = prj.ExportToExcel(path)
	if err != nil {
		return err
	}
	return nil
}
func (a *App) ExportCombinationsToExcel(path string) error {
	prj, err := utils.GetActualProject()
	if err != nil {
		return err
	}
	if strings.TrimSpace(path) == "" {
		return fmt.Errorf("ruta de exportación no válida")
	}
	err = prj.ExportCombinationsToExcel(path)
	if err != nil {
		return err
	}
	return nil
}
func sanitizeFilename(name string) string {
	replacer := strings.NewReplacer(
		"/", "_",
		"\\", "_",
		":", "-",
		"*", "-",
		"?", "",
		"\"", "",
		"<", "",
		">", "",
		"|", "-",
	)
	cleaned := strings.TrimSpace(replacer.Replace(name))
	if cleaned == "" {
		return "proyecto"
	}
	return cleaned
}
func (a *App) GetActualProject() (*utils.DbProject, error) {
	prj, err := utils.GetActualProject()
	if err != nil {
		return nil, err
	}
	return prj, nil
}
func (a *App) AddRelation(idEnt1 int, idEnt2 int, relation string) error {
	prj, err := utils.GetActualProject()
	if err != nil {
		return err
	}
	err = prj.AddRelation(idEnt1, idEnt2, relation)
	if err != nil {
		return err
	}
	return nil
}
func (a *App) AddAttribute(entityId int, name string, description string, attType string, attKeyType string, optional bool, attDomain []string) error {
	prj, err := utils.GetActualProject()
	if err != nil {
		return err
	}
	if err := prj.AddAttribute(entityId, name, description, attType, utils.AttributeKeyType(attKeyType), optional, attDomain); err != nil {
		return err
	}
	return nil
}
func (a *App) AddIntersectionAttribute(relationID int, name string, description string, attType string, optional bool, attDomain []string) error {
	prj, err := utils.GetActualProject()
	if err != nil {
		return err
	}
	if err := prj.AddIntersectionAttribute(relationID, name, description, attType, optional, attDomain); err != nil {
		return err
	}
	return nil
}
func (a *App) RemoveRelation(id int) error {
	prj, err := utils.GetActualProject()
	if err != nil {
		return err
	}
	err = prj.RemoveRelation(id)
	if err != nil {
		return err
	}
	return nil
}
func (a *App) EditAttribute(entityId int, attributeId int, name string, description string, attType string, attKeyType string, optional bool, attDomain []string) error {
	prj, err := utils.GetActualProject()
	if err != nil {
		return err
	}
	if err := prj.EditAttribute(entityId, attributeId, name, description, attType, utils.AttributeKeyType(attKeyType), optional, attDomain); err != nil {
		return err
	}
	return nil
}
func (a *App) EditIntersectionAttribute(relationID int, attributeId int, name string, description string, attType string, optional bool, attDomain []string) error {
	prj, err := utils.GetActualProject()
	if err != nil {
		return err
	}
	if err := prj.EditIntersectionAttribute(relationID, attributeId, name, description, attType, optional, attDomain); err != nil {
		return err
	}
	return nil
}
func (a *App) RemoveAttribute(entityId int, attributeId int) error {
	prj, err := utils.GetActualProject()
	if err != nil {
		return err
	}
	if err := prj.RemoveAttribute(entityId, attributeId); err != nil {
		return err
	}
	return nil
}
func (a *App) RemoveIntersectionAttribute(relationID int, attributeId int) error {
	prj, err := utils.GetActualProject()
	if err != nil {
		return err
	}
	if err := prj.RemoveIntersectionAttribute(relationID, attributeId); err != nil {
		return err
	}
	return nil
}
func (a *App) RemoveEntity(id int) error {
	prj, err := utils.GetActualProject()
	if err != nil {
		return err
	}
	if err := prj.RemoveEntity(id); err != nil {
		return err
	}
	return nil
}
func (a *App) GetIntersectionEntityByRelationID(relationID int) (*utils.IntersectionEntity, error) {
	p, err := a.GetActualProject()
	if err != nil {
		return nil, err
	}
	return p.GetIntersectionEntityByRelationID(relationID), nil
}

func (a *App) IntersectionHasAttributes(relationID int) (bool, error) {
	p, err := a.GetActualProject()
	if err != nil {
		return false, err
	}
	return p.IntersectionHasAttributes(relationID), nil
}

func (a *App) EditIntersectionEntityDescription(relationID int, description string) error {
	prj, err := utils.GetActualProject()
	if err != nil {
		return err
	}
	if err := prj.EditIntersectionEntityDescription(relationID, description); err != nil {
		return err
	}
	return nil
}
func (a *App) MarkEntityStatus(entityId int, status bool) error {
	prj, err := utils.GetActualProject()
	if err != nil {
		return err
	}
	if err := prj.MarkEntityStatus(entityId, status); err != nil {
		return err
	}
	return nil
}
func (a *App) MoveAttribute(entityId int, attributeId int, direction string) error {
	prj, err := utils.GetActualProject()
	if err != nil {
		return err
	}
	if err := prj.MoveAttribute(entityId, attributeId, direction); err != nil {
		return err
	}
	return nil
}
func (a *App) MoveIntersectionAttribute(relationID int, attributeId int, direction string) error {
	prj, err := utils.GetActualProject()
	if err != nil {
		return err
	}
	if err := prj.MoveIntersectionAttribute(relationID, attributeId, direction); err != nil {
		return err
	}
	return nil
}
func (a *App) GetEntity(id int) (*utils.Entity, error) {
	prj, err := utils.GetActualProject()
	if err != nil {
		return nil, err
	}
	ent := prj.GetEntity(id)
	if ent == nil {
		return nil, fmt.Errorf("Entity not found")
	}
	return ent, nil
}
func (a *App) GetCombinatory() ([]utils.RelationView, error) {
	prj, err := utils.GetActualProject()
	if err != nil {
		return nil, err
	}
	comb := prj.GetCombinatoryModel()
	return comb, nil
}

func (a *App) GetRelationTypes() []string {
	return utils.GetAllowedRelationTypes()
}

type schemaExportData struct {
	Entities             []utils.Entity             `json:"entities"`
	IntersectionEntities []utils.IntersectionEntity `json:"intersection_entities"`
	Relations            []utils.Relation           `json:"relations"`
}

type openAIResponsesRequest struct {
	Model string `json:"model"`
	Input string `json:"input"`
}

type openAIResponsesResponse struct {
	OutputText string `json:"output_text"`
	Output     []struct {
		Type    string `json:"type"`
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	} `json:"output"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func (a *App) ExportEntitiesToJSON(entityIds []int, intersectionIds []int) (string, error) {
	project, err := a.GetActualProject()
	if err != nil {
		return "", err
	}

	exportData, err := buildSchemaExport(project, entityIds, intersectionIds)
	if err != nil {
		return "", err
	}

	jsonData, err := json.MarshalIndent(exportData, "", "  ")
	if err != nil {
		return "", err
	}

	return string(jsonData), nil
}

func (a *App) GetAISettings() (*utils.AISettings, error) {
	return utils.GetAISettings()
}

func (a *App) SaveOpenAIAPIKey(apiKey string) (*utils.AISettings, error) {
	return utils.SaveOpenAIAPIKey(apiKey)
}

func (a *App) GenerateSQLFromEntities(entityIds []int, intersectionIds []int, database string) (*utils.SQLGenerationResult, error) {
	targetDatabase := strings.TrimSpace(database)
	if targetDatabase == "" {
		return nil, fmt.Errorf("selecciona una base de datos destino")
	}

	settings, err := utils.LoadAppConfig()
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(settings.OpenAIAPIKey) == "" {
		return nil, fmt.Errorf("configura una API key de OpenAI antes de generar SQL")
	}

	project, err := a.GetActualProject()
	if err != nil {
		return nil, err
	}

	exportData, err := buildSchemaExport(project, entityIds, intersectionIds)
	if err != nil {
		return nil, err
	}

	exportJSONBytes, err := json.MarshalIndent(exportData, "", "  ")
	if err != nil {
		return nil, err
	}
	exportJSON := string(exportJSONBytes)

	sqlCode, err := generateSQLWithOpenAI(exportJSON, targetDatabase, settings.OpenAIModel, settings.OpenAIAPIKey)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(sqlCode) == "" {
		return nil, fmt.Errorf("la IA no devolvió código SQL")
	}

	return &utils.SQLGenerationResult{
		Database:   targetDatabase,
		Model:      settings.OpenAIModel,
		SQL:        strings.TrimSpace(sqlCode),
		ExportJSON: exportJSON,
	}, nil
}

func buildSchemaExport(project *utils.DbProject, entityIds []int, intersectionIds []int) (*schemaExportData, error) {
	if project == nil {
		return nil, fmt.Errorf("proyecto no cargado")
	}

	selectedEntities := make([]utils.Entity, 0)
	entityIdMap := make(map[int]bool)
	for _, id := range entityIds {
		entityIdMap[id] = true
	}

	for _, entity := range project.Entities {
		if entityIdMap[entity.Id] {
			selectedEntities = append(selectedEntities, entity)
		}
	}

	selectedIntersections := make([]utils.IntersectionEntity, 0)
	intersectionIdMap := make(map[int]bool)
	for _, id := range intersectionIds {
		intersectionIdMap[id] = true
	}

	for _, intersection := range project.IntersectionEntities {
		// Incluir si está explícitamente seleccionada O si sus padres están seleccionados
		include := false
		if intersectionIdMap[intersection.Entity.Id] {
			include = true
		} else {
			// Encontrar la relación para saber los padres
			var rel *utils.Relation
			for _, r := range project.Relations {
				if r.Id == intersection.RelationID {
					rel = &r
					break
				}
			}
			if rel != nil && entityIdMap[rel.IdEntity1] && entityIdMap[rel.IdEntity2] {
				include = true
			}
		}

		if include {
			selectedIntersections = append(selectedIntersections, intersection)
		}
	}

	// También incluir las relaciones entre las entidades seleccionadas (fuertes o intersecciones vía sus padres)
	selectedRelations := make([]utils.Relation, 0)
	for _, rel := range project.Relations {
		if entityIdMap[rel.IdEntity1] && entityIdMap[rel.IdEntity2] {
			selectedRelations = append(selectedRelations, rel)
		}
	}

	if len(selectedEntities) == 0 && len(selectedIntersections) == 0 {
		return nil, fmt.Errorf("selecciona al menos una tabla para exportar")
	}

	return &schemaExportData{
		Entities:             selectedEntities,
		IntersectionEntities: selectedIntersections,
		Relations:            selectedRelations,
	}, nil
}

func (a *App) PickProjectJSON() (string, error) {
	return runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Selecciona proyecto.json",
		Filters: []runtime.FileFilter{
			{DisplayName: "JSON", Pattern: "*.json"},
			{DisplayName: "Todos", Pattern: "*"},
		},
	})
}
func (a *App) CreateProjectJSONPath() (string, error) {
	return runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "Crear proyecto.json",
		DefaultFilename: "proyecto.json",
		Filters: []runtime.FileFilter{
			{DisplayName: "JSON", Pattern: "*.json"},
			{DisplayName: "Todos", Pattern: "*"},
		},
	})
}

func (a *App) AddEntity(name string, description string) error {
	prj, err := utils.GetActualProject()
	if err != nil {
		return err
	}
	prj.AddEntity(name, description)
	if err != nil {
		return err
	}
	return nil
}
func (a *App) EditEntity(id int, name string, description string) error {
	prj, err := utils.GetActualProject()
	if err != nil {
		return err
	}
	err = prj.EditEntity(id, name, description)
	if err != nil {
		return err
	}
	return nil
}
func (a *App) MoveEntity(id int, direction string) error {
	prj, err := utils.GetActualProject()
	if err != nil {
		return err
	}
	if err := prj.MoveEntity(id, direction); err != nil {
		return err
	}
	return nil
}

func generateSQLWithOpenAI(exportJSON string, database string, model string, apiKey string) (string, error) {
	prompt := fmt.Sprintf(`Genera el script SQL DDL para %s a partir del siguiente esquema JSON.

Reglas:
- Responde solo con SQL valido.
- No uses bloques markdown.
- Crea tablas, claves primarias, claves foraneas, tipos y restricciones razonables segun el modelo.
- Respeta las entidades de interseccion y sus relaciones.
- Si alguna definicion esta incompleta, elige un tipo SQL conservador y consistente.
- Usa nombres limpios y listos para ejecutar.

Esquema JSON:
%s`, database, exportJSON)

	requestBody := openAIResponsesRequest{
		Model: model,
		Input: prompt,
	}

	payload, err := json.Marshal(requestBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest(http.MethodPost, "https://api.openai.com/v1/responses", bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(apiKey))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 90 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode >= 400 {
		var failure openAIResponsesResponse
		if err := json.Unmarshal(body, &failure); err == nil && failure.Error != nil && failure.Error.Message != "" {
			return "", fmt.Errorf("openai devolvió un error: %s", failure.Error.Message)
		}
		return "", fmt.Errorf("openai devolvió un error HTTP %d", resp.StatusCode)
	}

	var response openAIResponsesResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return "", err
	}

	outputText := extractOpenAIText(response)
	if strings.TrimSpace(outputText) == "" {
		return "", fmt.Errorf("openai no devolvió texto utilizable")
	}

	return outputText, nil
}

func extractOpenAIText(response openAIResponsesResponse) string {
	if strings.TrimSpace(response.OutputText) != "" {
		return strings.TrimSpace(response.OutputText)
	}

	parts := make([]string, 0)
	for _, item := range response.Output {
		for _, content := range item.Content {
			switch content.Type {
			case "output_text", "text", "summary_text", "reasoning_text":
				if strings.TrimSpace(content.Text) != "" {
					parts = append(parts, strings.TrimSpace(content.Text))
				}
			}
		}
	}

	return strings.TrimSpace(strings.Join(parts, "\n\n"))
}

func (a *App) AddBigProcess(name string, description string) error {
	prj, err := utils.GetActualProject()
	if err != nil {
		return err
	}
	return prj.AddBigProcess(name, description)
}

func (a *App) EditBigProcess(id int, name string, description string) error {
	prj, err := utils.GetActualProject()
	if err != nil {
		return err
	}
	return prj.EditBigProcess(id, name, description)
}

func (a *App) RemoveBigProcess(id int) error {
	prj, err := utils.GetActualProject()
	if err != nil {
		return err
	}
	return prj.RemoveBigProcess(id)
}

func (a *App) MoveBigProcess(id int, direction string) error {
	prj, err := utils.GetActualProject()
	if err != nil {
		return err
	}
	return prj.MoveBigProcess(id, direction)
}

func (a *App) AddProcess(bigProcessID int, name string, description string) error {
	prj, err := utils.GetActualProject()
	if err != nil {
		return err
	}
	return prj.AddProcess(bigProcessID, name, description)
}

func (a *App) EditProcess(bigProcessID int, processID int, name string, description string) error {
	prj, err := utils.GetActualProject()
	if err != nil {
		return err
	}
	return prj.EditProcess(bigProcessID, processID, name, description)
}

func (a *App) RemoveProcess(bigProcessID int, processID int) error {
	prj, err := utils.GetActualProject()
	if err != nil {
		return err
	}
	return prj.RemoveProcess(bigProcessID, processID)
}

func (a *App) MoveProcess(bigProcessID int, processID int, direction string) error {
	prj, err := utils.GetActualProject()
	if err != nil {
		return err
	}
	return prj.MoveProcess(bigProcessID, processID, direction)
}

func (a *App) AddStep(bigProcessID int, processID int, name string, description string) error {
	prj, err := utils.GetActualProject()
	if err != nil {
		return err
	}
	return prj.AddStep(bigProcessID, processID, name, description)
}

func (a *App) EditStep(bigProcessID int, processID int, stepID int, name string, description string) error {
	prj, err := utils.GetActualProject()
	if err != nil {
		return err
	}
	return prj.EditStep(bigProcessID, processID, stepID, name, description)
}

func (a *App) RemoveStep(bigProcessID int, processID int, stepID int) error {
	prj, err := utils.GetActualProject()
	if err != nil {
		return err
	}
	return prj.RemoveStep(bigProcessID, processID, stepID)
}

func (a *App) MoveStep(bigProcessID int, processID int, stepID int, direction string) error {
	prj, err := utils.GetActualProject()
	if err != nil {
		return err
	}
	return prj.MoveStep(bigProcessID, processID, stepID, direction)
}

func (a *App) AddResource(bigProcessID int, processID int, stepID int, tableID int, role string) error {
	prj, err := utils.GetActualProject()
	if err != nil {
		return err
	}
	return prj.AddResource(bigProcessID, processID, stepID, tableID, role)
}

func (a *App) EditResource(bigProcessID int, processID int, stepID int, resourceID int, tableID int, role string) error {
	prj, err := utils.GetActualProject()
	if err != nil {
		return err
	}
	return prj.EditResource(bigProcessID, processID, stepID, resourceID, tableID, role)
}

func (a *App) RemoveResource(bigProcessID int, processID int, stepID int, resourceID int) error {
	prj, err := utils.GetActualProject()
	if err != nil {
		return err
	}
	return prj.RemoveResource(bigProcessID, processID, stepID, resourceID)
}

func (a *App) GetRole(id int) (*utils.Role, error) {
	prj, err := utils.GetActualProject()
	if err != nil {
		return nil, err
	}
	role := prj.GetRole(id)
	if role == nil {
		return nil, fmt.Errorf("role not found")
	}
	return role, nil
}

func (a *App) AddRole(name string, description string) error {
	prj, err := utils.GetActualProject()
	if err != nil {
		return err
	}
	return prj.AddRole(name, description)
}

func (a *App) EditRole(id int, name string, description string) error {
	prj, err := utils.GetActualProject()
	if err != nil {
		return err
	}
	return prj.EditRole(id, name, description)
}

func (a *App) RemoveRole(id int) error {
	prj, err := utils.GetActualProject()
	if err != nil {
		return err
	}
	return prj.RemoveRole(id)
}

func (a *App) AddRoleProcessPermission(roleID int, processID int) error {
	prj, err := utils.GetActualProject()
	if err != nil {
		return err
	}
	return prj.AddRoleProcessPermission(roleID, processID)
}

func (a *App) EditRoleProcessPermission(roleID int, permissionID int, processID int) error {
	prj, err := utils.GetActualProject()
	if err != nil {
		return err
	}
	return prj.EditRoleProcessPermission(roleID, permissionID, processID)
}

func (a *App) RemoveRoleProcessPermission(roleID int, permissionID int) error {
	prj, err := utils.GetActualProject()
	if err != nil {
		return err
	}
	return prj.RemoveRoleProcessPermission(roleID, permissionID)
}

func (a *App) AddRoleTablePermission(roleID int, tableID int, insertPermission bool, deletePermission bool, updatePermission bool, viewPermission bool) error {
	prj, err := utils.GetActualProject()
	if err != nil {
		return err
	}
	return prj.AddRoleTablePermission(roleID, tableID, insertPermission, deletePermission, updatePermission, viewPermission)
}

func (a *App) EditRoleTablePermission(roleID int, permissionID int, tableID int, insertPermission bool, deletePermission bool, updatePermission bool, viewPermission bool) error {
	prj, err := utils.GetActualProject()
	if err != nil {
		return err
	}
	return prj.EditRoleTablePermission(roleID, permissionID, tableID, insertPermission, deletePermission, updatePermission, viewPermission)
}

func (a *App) RemoveRoleTablePermission(roleID int, permissionID int) error {
	prj, err := utils.GetActualProject()
	if err != nil {
		return err
	}
	return prj.RemoveRoleTablePermission(roleID, permissionID)
}
