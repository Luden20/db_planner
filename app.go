package main

import (
	"context"
	"db_planner/utils"
	"fmt"

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
func (a *App) CreateNew(path string) (*utils.DbProject, error) {
	prj, err := utils.CreateNew(path)
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
