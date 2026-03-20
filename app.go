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
func (a *App) AddAttribute(entityId int, name string, description string, attType string) error {
	prj, err := utils.GetActualProject()
	if err != nil {
		return err
	}
	if err := prj.AddAttribute(entityId, name, description, attType); err != nil {
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
func (a *App) EditAttribute(entityId int, attributeId int, name string, description string, attType string) error {
	prj, err := utils.GetActualProject()
	if err != nil {
		return err
	}
	if err := prj.EditAttribute(entityId, attributeId, name, description, attType); err != nil {
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
