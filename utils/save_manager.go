package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

type SingletonData struct {
	project *DbProject
	path    string
}

var lock = &sync.Mutex{}

var singleInstance *SingletonData

func Eject() error {
	defer lock.Unlock()
	lock.Lock()
	singleInstance = nil
	return nil
}
func SaveChanges() error {
	lock.Lock()
	defer lock.Unlock()
	prj, err := GetActualProject()
	if err != nil {
		return err
	}
	data, err := json.Marshal(prj)
	if err != nil {
		return err
	}
	path, err := GetActualPath()
	if err != nil {
		return err
	}
	err = os.WriteFile(*path, data, 0644)
	if err != nil {
		return err
	}
	return nil
}
func GetActualPath() (*string, error) {
	if singleInstance != nil {
		return &singleInstance.path, nil
	}
	return nil, fmt.Errorf("project not initialized or set")
}
func GetActualProject() (*DbProject, error) {
	if singleInstance != nil {
		return singleInstance.project, nil
	}
	return nil, fmt.Errorf("project not initialized or set")
}
func CreateNew(path string) (*DbProject, error) {
	if singleInstance == nil {
		lock.Lock()
		defer lock.Unlock()
		if singleInstance == nil {
			var project DbProject
			singleInstance = &SingletonData{
				project: &project,
				path:    path,
			}
		} else {
			fmt.Println("Single instance already created.")
		}
	} else {
		fmt.Println("Single instance already created.")
	}
	return singleInstance.project, nil
}
func LoadProjectFromJson(path string) (*DbProject, error) {
	if singleInstance == nil {
		lock.Lock()
		defer lock.Unlock()
		if singleInstance == nil {
			file, err := os.ReadFile(path)
			if err != nil {
				return nil, err
			}
			data := []byte(file)
			var project DbProject
			err = json.Unmarshal(data, &project)
			if err != nil {
				return nil, err
			}
			singleInstance = &SingletonData{
				project: &project,
				path:    path,
			}
		} else {
			fmt.Println("Single instance already created.")
		}
	} else {
		fmt.Println("Single instance already created.")
	}
	return singleInstance.project, nil
}
