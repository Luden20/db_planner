package main

import (
	"context"
	"db_planner/utils"
	"fmt"
	"strings"

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
func (a *App) AddAttribute(entityId int, name string, description string, attType string, attKeyType string, attDomain []string) error {
	prj, err := utils.GetActualProject()
	if err != nil {
		return err
	}
	if err := prj.AddAttribute(entityId, name, description, attType, utils.AttributeKeyType(attKeyType), attDomain); err != nil {
		return err
	}
	return nil
}
func (a *App) AddIntersectionAttribute(relationID int, name string, description string, attType string, attDomain []string) error {
	prj, err := utils.GetActualProject()
	if err != nil {
		return err
	}
	if err := prj.AddIntersectionAttribute(relationID, name, description, attType, attDomain); err != nil {
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
func (a *App) EditAttribute(entityId int, attributeId int, name string, description string, attType string, attKeyType string, attDomain []string) error {
	prj, err := utils.GetActualProject()
	if err != nil {
		return err
	}
	if err := prj.EditAttribute(entityId, attributeId, name, description, attType, utils.AttributeKeyType(attKeyType), attDomain); err != nil {
		return err
	}
	return nil
}
func (a *App) EditIntersectionAttribute(relationID int, attributeId int, name string, description string, attType string, attDomain []string) error {
	prj, err := utils.GetActualProject()
	if err != nil {
		return err
	}
	if err := prj.EditIntersectionAttribute(relationID, attributeId, name, description, attType, attDomain); err != nil {
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
