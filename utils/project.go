package utils

type DbProject struct {
	Name                        string
	Entities                    []Entity
	EntitiesLastMax             int
	IntersectionEntities        []IntersectionEntity
	IntersectionEntitiesLastMax int
	Relations                   []Relation
	BigProcesses                []BigProcess
	Roles                       []Role
	RelationsLastMax            int
	AttributesLastMax           int
	BigProcessLastMax           int
	ProcessLastMax              int
	StepsLastMax                int
	StepResourceLastMax         int
	RoleLastMax                 int
	ProcessPermissionLastMax    int
	RoleTablePermissionLastMax  int
}

func (p *DbProject) syncCounters() {
	maxBigProcessID := 0
	maxProcessID := 0
	maxStepID := 0
	maxStepResourceID := 0
	maxRoleID := 0
	maxProcessPermissionID := 0
	maxRoleTablePermissionID := 0
	for _, bigProcess := range p.BigProcesses {
		if bigProcess.Id > maxBigProcessID {
			maxBigProcessID = bigProcess.Id
		}
		for _, process := range bigProcess.Processes {
			if process.Id > maxProcessID {
				maxProcessID = process.Id
			}
			for _, step := range process.Steps {
				if step.Id > maxStepID {
					maxStepID = step.Id
				}
				for _, resource := range step.Resources {
					if resource.Id > maxStepResourceID {
						maxStepResourceID = resource.Id
					}
				}
			}
		}
	}
	if p.BigProcessLastMax < maxBigProcessID {
		p.BigProcessLastMax = maxBigProcessID
	}
	if p.ProcessLastMax < maxProcessID {
		p.ProcessLastMax = maxProcessID
	}
	if p.StepsLastMax < maxStepID {
		p.StepsLastMax = maxStepID
	}
	if p.StepResourceLastMax < maxStepResourceID {
		p.StepResourceLastMax = maxStepResourceID
	}
	p.ensureRoles()
	for _, role := range p.Roles {
		if role.Id > maxRoleID {
			maxRoleID = role.Id
		}
		for _, permission := range role.ProcessPermissions {
			if permission.Id > maxProcessPermissionID {
				maxProcessPermissionID = permission.Id
			}
		}
		for _, permission := range role.TablePermissions {
			if permission.Id > maxRoleTablePermissionID {
				maxRoleTablePermissionID = permission.Id
			}
		}
	}
	if p.RoleLastMax < maxRoleID {
		p.RoleLastMax = maxRoleID
	}
	if p.ProcessPermissionLastMax < maxProcessPermissionID {
		p.ProcessPermissionLastMax = maxProcessPermissionID
	}
	if p.RoleTablePermissionLastMax < maxRoleTablePermissionID {
		p.RoleTablePermissionLastMax = maxRoleTablePermissionID
	}
	p.ensureEntities()
	maxEntityID := 0
	for _, entity := range p.Entities {
		if entity.Id > maxEntityID {
			maxEntityID = entity.Id
		}
	}
	if p.EntitiesLastMax < maxEntityID {
		p.EntitiesLastMax = maxEntityID
	}

	maxIntersectionEntityID := 0
	for _, item := range p.IntersectionEntities {
		if item.Entity.Id > maxIntersectionEntityID {
			maxIntersectionEntityID = item.Entity.Id
		}
	}
	if p.IntersectionEntitiesLastMax < maxIntersectionEntityID {
		p.IntersectionEntitiesLastMax = maxIntersectionEntityID
	}

	maxRelationID := -1
	for _, relation := range p.Relations {
		if relation.Id > maxRelationID {
			maxRelationID = relation.Id
		}
	}
	nextRelationID := maxRelationID + 1
	if p.RelationsLastMax < nextRelationID {
		p.RelationsLastMax = nextRelationID
	}

	maxAttributeID := 0
	p.ensureAttributes()
	for _, entity := range p.Entities {
		for _, att := range entity.Attributes {
			if att.Id > maxAttributeID {
				maxAttributeID = att.Id
			}
		}
	}
	for _, item := range p.IntersectionEntities {
		for _, att := range item.Entity.Attributes {
			if att.Id > maxAttributeID {
				maxAttributeID = att.Id
			}
		}
	}
	if p.AttributesLastMax < maxAttributeID {
		p.AttributesLastMax = maxAttributeID
	}
}
