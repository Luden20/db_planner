package utils

import "fmt"

type BigProcess struct {
	Id          int
	Name        string
	Description string
	Processes   []Process
}

type Process struct {
	Id          int
	Name        string
	Description string
	Steps       []Step
}
type Step struct {
	Id          int
	Name        string
	Description string
	Order       int
	Resources   []StepResource
}
type StepResource struct {
	Id      int
	TableId int
	Role    string
}

var allowedResourceTypes = map[string]struct{}{
	"Input":  {},
	"Output": {},
}

func isAllowedResourceType(resourceType string) bool {
	_, ok := allowedResourceTypes[resourceType]
	return ok
}

func (bigProcess *BigProcess) getProcessByID(id int) *Process {
	for idx := range bigProcess.Processes {
		if bigProcess.Processes[idx].Id == id {
			return &bigProcess.Processes[idx]
		}
	}
	return nil
}

func (process *Process) getStepByID(id int) *Step {
	for idx := range process.Steps {
		if process.Steps[idx].Id == id {
			return &process.Steps[idx]
		}
	}
	return nil
}

func (step *Step) getResourceByID(id int) *StepResource {
	for idx := range step.Resources {
		if step.Resources[idx].Id == id {
			return &step.Resources[idx]
		}
	}
	return nil
}

func (p *DbProject) getProcessReference(processID int) *Process {
	for bigIdx := range p.BigProcesses {
		for procIdx := range p.BigProcesses[bigIdx].Processes {
			if p.BigProcesses[bigIdx].Processes[procIdx].Id == processID {
				return &p.BigProcesses[bigIdx].Processes[procIdx]
			}
		}
	}
	return nil
}

func (p *DbProject) ensureProcesses() {
	if p.BigProcesses == nil {
		p.BigProcesses = make([]BigProcess, 0)
	}
	for bigIdx := range p.BigProcesses {
		if p.BigProcesses[bigIdx].Processes == nil {
			p.BigProcesses[bigIdx].Processes = make([]Process, 0)
		}
		for procIdx := range p.BigProcesses[bigIdx].Processes {
			if p.BigProcesses[bigIdx].Processes[procIdx].Steps == nil {
				p.BigProcesses[bigIdx].Processes[procIdx].Steps = make([]Step, 0)
			}
			for stepIdx := range p.BigProcesses[bigIdx].Processes[procIdx].Steps {
				p.BigProcesses[bigIdx].Processes[procIdx].Steps[stepIdx].Order = stepIdx + 1
				if p.BigProcesses[bigIdx].Processes[procIdx].Steps[stepIdx].Resources == nil {
					p.BigProcesses[bigIdx].Processes[procIdx].Steps[stepIdx].Resources = make([]StepResource, 0)
				}
			}
		}
	}
}

func (process *Process) normalizeStepOrder() {
	for idx := range process.Steps {
		process.Steps[idx].Order = idx + 1
	}
}

func (p *DbProject) getBigProcessIndex(id int) int {
	for idx := range p.BigProcesses {
		if p.BigProcesses[idx].Id == id {
			return idx
		}
	}
	return -1
}

func (p *DbProject) GetBigProcessByID(id int) *BigProcess {
	idx := p.getBigProcessIndex(id)
	if idx < 0 {
		return nil
	}
	if p.BigProcesses[idx].Processes == nil {
		p.BigProcesses[idx].Processes = make([]Process, 0)
	}
	return &p.BigProcesses[idx]
}

func (p *DbProject) GetProcessByID(bigProcessID int, processID int) *Process {
	bigProcess := p.GetBigProcessByID(bigProcessID)
	if bigProcess == nil {
		return nil
	}
	return bigProcess.getProcessByID(processID)
}

func (p *DbProject) GetStepByID(bigProcessID int, processID int, stepID int) *Step {
	process := p.GetProcessByID(bigProcessID, processID)
	if process == nil {
		return nil
	}
	return process.getStepByID(stepID)
}

func (p *DbProject) AddBigProcess(name string, description string) error {
	p.BigProcesses = append(p.BigProcesses, BigProcess{
		Id:          p.BigProcessLastMax + 1,
		Name:        name,
		Description: description,
		Processes:   make([]Process, 0),
	})
	p.BigProcessLastMax++
	return nil
}

func (p *DbProject) EditBigProcess(id int, name string, description string) error {
	bigProcess := p.GetBigProcessByID(id)
	if bigProcess == nil {
		return fmt.Errorf("big process not found")
	}
	bigProcess.Name = name
	bigProcess.Description = description
	return nil
}

func (p *DbProject) RemoveBigProcess(id int) error {
	idx := p.getBigProcessIndex(id)
	if idx < 0 {
		return fmt.Errorf("big process not found")
	}
	for _, process := range p.BigProcesses[idx].Processes {
		p.removeRolePermissionsByProcessID(process.Id)
	}
	p.BigProcesses = append(p.BigProcesses[:idx], p.BigProcesses[idx+1:]...)
	return nil
}

func (p *DbProject) MoveBigProcess(id int, direction string) error {
	idx := p.getBigProcessIndex(id)
	if idx < 0 {
		return fmt.Errorf("big process not found")
	}
	if len(p.BigProcesses) <= 1 {
		return nil
	}

	switch direction {
	case "up":
		if idx == 0 {
			return nil
		}
		p.BigProcesses[idx], p.BigProcesses[idx-1] = p.BigProcesses[idx-1], p.BigProcesses[idx]
	case "down":
		if idx == len(p.BigProcesses)-1 {
			return nil
		}
		p.BigProcesses[idx], p.BigProcesses[idx+1] = p.BigProcesses[idx+1], p.BigProcesses[idx]
	default:
		return fmt.Errorf("unknown direction %s", direction)
	}
	return nil
}

func (p *DbProject) AddProcess(bigProcessID int, name string, description string) error {
	bigProcess := p.GetBigProcessByID(bigProcessID)
	if bigProcess == nil {
		return fmt.Errorf("big process not found")
	}
	bigProcess.Processes = append(bigProcess.Processes, Process{Id: p.ProcessLastMax + 1, Description: description, Name: name, Steps: make([]Step, 0)})
	p.ProcessLastMax = p.ProcessLastMax + 1
	return nil
}

func (p *DbProject) EditProcess(bigProcessID int, processID int, name string, description string) error {
	process := p.GetProcessByID(bigProcessID, processID)
	if process == nil {
		return fmt.Errorf("process not found")
	}
	process.Name = name
	process.Description = description
	return nil
}

func (p *DbProject) RemoveProcess(bigProcessID int, processID int) error {
	bigProcess := p.GetBigProcessByID(bigProcessID)
	if bigProcess == nil {
		return fmt.Errorf("big process not found")
	}
	for idx := range bigProcess.Processes {
		if bigProcess.Processes[idx].Id == processID {
			p.removeRolePermissionsByProcessID(processID)
			bigProcess.Processes = append(bigProcess.Processes[:idx], bigProcess.Processes[idx+1:]...)
			return nil
		}
	}
	return fmt.Errorf("process not found")
}

func (p *DbProject) MoveProcess(bigProcessID int, processID int, direction string) error {
	bigProcess := p.GetBigProcessByID(bigProcessID)
	if bigProcess == nil {
		return fmt.Errorf("big process not found")
	}
	if len(bigProcess.Processes) <= 1 {
		return nil
	}
	idx := -1
	for pos := range bigProcess.Processes {
		if bigProcess.Processes[pos].Id == processID {
			idx = pos
			break
		}
	}
	if idx < 0 {
		return fmt.Errorf("process not found")
	}
	switch direction {
	case "up":
		if idx == 0 {
			return nil
		}
		bigProcess.Processes[idx], bigProcess.Processes[idx-1] = bigProcess.Processes[idx-1], bigProcess.Processes[idx]
	case "down":
		if idx == len(bigProcess.Processes)-1 {
			return nil
		}
		bigProcess.Processes[idx], bigProcess.Processes[idx+1] = bigProcess.Processes[idx+1], bigProcess.Processes[idx]
	default:
		return fmt.Errorf("unknown direction %s", direction)
	}
	return nil
}

func (p *DbProject) AddStep(bigProcessID int, processID int, name string, description string) error {
	process := p.GetProcessByID(bigProcessID, processID)
	if process == nil {
		return fmt.Errorf("process not found")
	}
	process.Steps = append(process.Steps, Step{Id: p.StepsLastMax + 1, Name: name, Description: description, Order: len(process.Steps) + 1, Resources: make([]StepResource, 0)})
	p.StepsLastMax++
	process.normalizeStepOrder()
	return nil
}

func (p *DbProject) EditStep(bigProcessID int, processID int, stepID int, name string, description string) error {
	step := p.GetStepByID(bigProcessID, processID, stepID)
	if step == nil {
		return fmt.Errorf("step not found")
	}
	step.Name = name
	step.Description = description
	return nil
}

func (p *DbProject) RemoveStep(bigProcessID int, processID int, stepID int) error {
	process := p.GetProcessByID(bigProcessID, processID)
	if process == nil {
		return fmt.Errorf("process not found")
	}
	for idx := range process.Steps {
		if process.Steps[idx].Id == stepID {
			process.Steps = append(process.Steps[:idx], process.Steps[idx+1:]...)
			process.normalizeStepOrder()
			return nil
		}
	}
	return fmt.Errorf("step not found")
}

func (p *DbProject) MoveStep(bigProcessID int, processID int, stepID int, direction string) error {
	process := p.GetProcessByID(bigProcessID, processID)
	if process == nil {
		return fmt.Errorf("process not found")
	}
	if len(process.Steps) <= 1 {
		return nil
	}
	idx := -1
	for pos := range process.Steps {
		if process.Steps[pos].Id == stepID {
			idx = pos
			break
		}
	}
	if idx < 0 {
		return fmt.Errorf("step not found")
	}
	switch direction {
	case "up":
		if idx == 0 {
			return nil
		}
		process.Steps[idx], process.Steps[idx-1] = process.Steps[idx-1], process.Steps[idx]
	case "down":
		if idx == len(process.Steps)-1 {
			return nil
		}
		process.Steps[idx], process.Steps[idx+1] = process.Steps[idx+1], process.Steps[idx]
	default:
		return fmt.Errorf("unknown direction %s", direction)
	}
	process.normalizeStepOrder()
	return nil
}

func (p *DbProject) AddResource(bigProcessID int, processID int, stepID int, tableID int, role string) error {
	if !isAllowedResourceType(role) {
		return fmt.Errorf("invalid resource type: %s", role)
	}
	entity := p.GetEntity(tableID)
	if entity == nil {
		return fmt.Errorf("entity not found")
	}
	step := p.GetStepByID(bigProcessID, processID, stepID)
	if step == nil {
		return fmt.Errorf("step not found")
	}
	step.Resources = append(step.Resources, StepResource{Id: p.StepResourceLastMax + 1, TableId: tableID, Role: role})
	p.StepResourceLastMax++
	return nil
}

func (p *DbProject) EditResource(bigProcessID int, processID int, stepID int, resourceID int, tableID int, role string) error {
	if !isAllowedResourceType(role) {
		return fmt.Errorf("invalid resource type: %s", role)
	}
	entity := p.GetEntity(tableID)
	if entity == nil {
		return fmt.Errorf("entity not found")
	}
	step := p.GetStepByID(bigProcessID, processID, stepID)
	if step == nil {
		return fmt.Errorf("step not found")
	}
	resource := step.getResourceByID(resourceID)
	if resource == nil {
		return fmt.Errorf("resource not found")
	}
	resource.TableId = tableID
	resource.Role = role
	return nil
}

func (p *DbProject) RemoveResource(bigProcessID int, processID int, stepID int, resourceID int) error {
	step := p.GetStepByID(bigProcessID, processID, stepID)
	if step == nil {
		return fmt.Errorf("step not found")
	}
	for idx := range step.Resources {
		if step.Resources[idx].Id == resourceID {
			step.Resources = append(step.Resources[:idx], step.Resources[idx+1:]...)
			return nil
		}
	}
	return fmt.Errorf("resource not found")
}
