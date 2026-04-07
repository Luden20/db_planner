package utils

import "fmt"

type Role struct {
	Id                 int
	Name               string
	Description        string
	ProcessPermissions []ProcessPermission
	TablePermissions   []RoleTablePermission
}

type ProcessPermission struct {
	Id        int
	ProcessId int
}

type RoleTablePermission struct {
	Id               int
	TableId          int
	InsertPermission bool
	DeletePermission bool
	UpdatePermission bool
	ViewPermission   bool
}

func (role *Role) getProcessPermissionByID(id int) *ProcessPermission {
	for idx := range role.ProcessPermissions {
		if role.ProcessPermissions[idx].Id == id {
			return &role.ProcessPermissions[idx]
		}
	}
	return nil
}

func (role *Role) getTablePermissionByID(id int) *RoleTablePermission {
	for idx := range role.TablePermissions {
		if role.TablePermissions[idx].Id == id {
			return &role.TablePermissions[idx]
		}
	}
	return nil
}

func (role *Role) getProcessPermissionByProcessID(processID int) *ProcessPermission {
	for idx := range role.ProcessPermissions {
		if role.ProcessPermissions[idx].ProcessId == processID {
			return &role.ProcessPermissions[idx]
		}
	}
	return nil
}

func (role *Role) getTablePermissionByTableID(tableID int) *RoleTablePermission {
	for idx := range role.TablePermissions {
		if role.TablePermissions[idx].TableId == tableID {
			return &role.TablePermissions[idx]
		}
	}
	return nil
}

func (p *DbProject) ensureRoles() {
	if p.Roles == nil {
		p.Roles = make([]Role, 0)
	}
	for idx := range p.Roles {
		if p.Roles[idx].ProcessPermissions == nil {
			p.Roles[idx].ProcessPermissions = make([]ProcessPermission, 0)
		}
		if p.Roles[idx].TablePermissions == nil {
			p.Roles[idx].TablePermissions = make([]RoleTablePermission, 0)
		}
	}
}

func (p *DbProject) roleIndex(id int) int {
	for idx := range p.Roles {
		if p.Roles[idx].Id == id {
			return idx
		}
	}
	return -1
}

func (p *DbProject) GetRole(id int) *Role {
	idx := p.roleIndex(id)
	if idx < 0 {
		return nil
	}
	if p.Roles[idx].ProcessPermissions == nil {
		p.Roles[idx].ProcessPermissions = make([]ProcessPermission, 0)
	}
	if p.Roles[idx].TablePermissions == nil {
		p.Roles[idx].TablePermissions = make([]RoleTablePermission, 0)
	}
	return &p.Roles[idx]
}

func (p *DbProject) AddRole(name string, description string) error {
	p.Roles = append(p.Roles, Role{
		Id:                 p.RoleLastMax + 1,
		Name:               name,
		Description:        description,
		ProcessPermissions: make([]ProcessPermission, 0),
		TablePermissions:   make([]RoleTablePermission, 0),
	})
	p.RoleLastMax++
	return nil
}

func (p *DbProject) EditRole(id int, name string, description string) error {
	role := p.GetRole(id)
	if role == nil {
		return fmt.Errorf("role not found")
	}
	role.Name = name
	role.Description = description
	return nil
}

func (p *DbProject) RemoveRole(id int) error {
	idx := p.roleIndex(id)
	if idx < 0 {
		return fmt.Errorf("role not found")
	}
	p.Roles = append(p.Roles[:idx], p.Roles[idx+1:]...)
	return nil
}

func (p *DbProject) AddRoleProcessPermission(roleID int, processID int) error {
	role := p.GetRole(roleID)
	if role == nil {
		return fmt.Errorf("role not found")
	}
	if p.getProcessReference(processID) == nil {
		return fmt.Errorf("process not found")
	}
	if role.getProcessPermissionByProcessID(processID) != nil {
		return fmt.Errorf("process permission already exists")
	}

	role.ProcessPermissions = append(role.ProcessPermissions, ProcessPermission{
		Id:        p.ProcessPermissionLastMax + 1,
		ProcessId: processID,
	})
	p.ProcessPermissionLastMax++
	return nil
}

func (p *DbProject) EditRoleProcessPermission(roleID int, permissionID int, processID int) error {
	role := p.GetRole(roleID)
	if role == nil {
		return fmt.Errorf("role not found")
	}
	if p.getProcessReference(processID) == nil {
		return fmt.Errorf("process not found")
	}
	permission := role.getProcessPermissionByID(permissionID)
	if permission == nil {
		return fmt.Errorf("process permission not found")
	}
	if existing := role.getProcessPermissionByProcessID(processID); existing != nil && existing.Id != permissionID {
		return fmt.Errorf("process permission already exists")
	}
	permission.ProcessId = processID
	return nil
}

func (p *DbProject) RemoveRoleProcessPermission(roleID int, permissionID int) error {
	role := p.GetRole(roleID)
	if role == nil {
		return fmt.Errorf("role not found")
	}
	for idx := range role.ProcessPermissions {
		if role.ProcessPermissions[idx].Id == permissionID {
			role.ProcessPermissions = append(role.ProcessPermissions[:idx], role.ProcessPermissions[idx+1:]...)
			return nil
		}
	}
	return fmt.Errorf("process permission not found")
}

func (p *DbProject) AddRoleTablePermission(roleID int, tableID int, insertPermission bool, deletePermission bool, updatePermission bool, viewPermission bool) error {
	role := p.GetRole(roleID)
	if role == nil {
		return fmt.Errorf("role not found")
	}
	if p.GetEntity(tableID) == nil {
		return fmt.Errorf("entity not found")
	}
	if role.getTablePermissionByTableID(tableID) != nil {
		return fmt.Errorf("table permission already exists")
	}

	role.TablePermissions = append(role.TablePermissions, RoleTablePermission{
		Id:               p.RoleTablePermissionLastMax + 1,
		TableId:          tableID,
		InsertPermission: insertPermission,
		DeletePermission: deletePermission,
		UpdatePermission: updatePermission,
		ViewPermission:   viewPermission,
	})
	p.RoleTablePermissionLastMax++
	return nil
}

func (p *DbProject) EditRoleTablePermission(roleID int, permissionID int, tableID int, insertPermission bool, deletePermission bool, updatePermission bool, viewPermission bool) error {
	role := p.GetRole(roleID)
	if role == nil {
		return fmt.Errorf("role not found")
	}
	if p.GetEntity(tableID) == nil {
		return fmt.Errorf("entity not found")
	}
	permission := role.getTablePermissionByID(permissionID)
	if permission == nil {
		return fmt.Errorf("table permission not found")
	}
	if existing := role.getTablePermissionByTableID(tableID); existing != nil && existing.Id != permissionID {
		return fmt.Errorf("table permission already exists")
	}

	permission.TableId = tableID
	permission.InsertPermission = insertPermission
	permission.DeletePermission = deletePermission
	permission.UpdatePermission = updatePermission
	permission.ViewPermission = viewPermission
	return nil
}

func (p *DbProject) RemoveRoleTablePermission(roleID int, permissionID int) error {
	role := p.GetRole(roleID)
	if role == nil {
		return fmt.Errorf("role not found")
	}
	for idx := range role.TablePermissions {
		if role.TablePermissions[idx].Id == permissionID {
			role.TablePermissions = append(role.TablePermissions[:idx], role.TablePermissions[idx+1:]...)
			return nil
		}
	}
	return fmt.Errorf("table permission not found")
}

func (p *DbProject) removeRolePermissionsByProcessID(processID int) {
	for roleIdx := range p.Roles {
		filtered := make([]ProcessPermission, 0, len(p.Roles[roleIdx].ProcessPermissions))
		for _, permission := range p.Roles[roleIdx].ProcessPermissions {
			if permission.ProcessId == processID {
				continue
			}
			filtered = append(filtered, permission)
		}
		p.Roles[roleIdx].ProcessPermissions = filtered
	}
}

func (p *DbProject) removeRolePermissionsByTableID(tableID int) {
	for roleIdx := range p.Roles {
		filtered := make([]RoleTablePermission, 0, len(p.Roles[roleIdx].TablePermissions))
		for _, permission := range p.Roles[roleIdx].TablePermissions {
			if permission.TableId == tableID {
				continue
			}
			filtered = append(filtered, permission)
		}
		p.Roles[roleIdx].TablePermissions = filtered
	}
}
