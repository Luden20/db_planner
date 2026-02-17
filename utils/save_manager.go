package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

var lock = &sync.Mutex{}

var singleInstance *DbProject

func GetActualProject() (*DbProject, error) {
	if singleInstance != nil {
		return singleInstance, nil
	}
	return nil, fmt.Errorf("project not initialized or set")
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
			singleInstance = &project
		} else {
			fmt.Println("Single instance already created.")
		}
	} else {
		fmt.Println("Single instance already created.")
	}
	return singleInstance, nil
}
