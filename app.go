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

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

func (a *App) OpenPath(path string) (*utils.DbProject, error) {
	prj, err := utils.LoadProjectFromJson(path)
	if err != nil {
		return nil, err
	}
	return prj, nil
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
