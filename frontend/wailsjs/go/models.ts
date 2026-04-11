export namespace utils {
	
	export class AISettings {
	    HasAPIKey: boolean;
	    OpenAIModel: string;
	
	    static createFrom(source: any = {}) {
	        return new AISettings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.HasAPIKey = source["HasAPIKey"];
	        this.OpenAIModel = source["OpenAIModel"];
	    }
	}
	export class AnalysisIssue {
	    type: string;
	    entities: string[];
	    pathCount: number;
	    paths: string[][];
	
	    static createFrom(source: any = {}) {
	        return new AnalysisIssue(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.type = source["type"];
	        this.entities = source["entities"];
	        this.pathCount = source["pathCount"];
	        this.paths = source["paths"];
	    }
	}
	export class AnalysisReport {
	    issues: AnalysisIssue[];
	    edgeList: string[];
	
	    static createFrom(source: any = {}) {
	        return new AnalysisReport(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.issues = this.convertValues(source["issues"], AnalysisIssue);
	        this.edgeList = source["edgeList"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Attribute {
	    Id: number;
	    Name: string;
	    Description: string;
	    Type: string;
	    KeyType: string;
	    Optional: boolean;
	    Domain: string[];
	
	    static createFrom(source: any = {}) {
	        return new Attribute(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Id = source["Id"];
	        this.Name = source["Name"];
	        this.Description = source["Description"];
	        this.Type = source["Type"];
	        this.KeyType = source["KeyType"];
	        this.Optional = source["Optional"];
	        this.Domain = source["Domain"];
	    }
	}
	export class StepResource {
	    Id: number;
	    TableId: number;
	    Role: string;
	
	    static createFrom(source: any = {}) {
	        return new StepResource(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Id = source["Id"];
	        this.TableId = source["TableId"];
	        this.Role = source["Role"];
	    }
	}
	export class Step {
	    Id: number;
	    Name: string;
	    Description: string;
	    Order: number;
	    Resources: StepResource[];
	
	    static createFrom(source: any = {}) {
	        return new Step(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Id = source["Id"];
	        this.Name = source["Name"];
	        this.Description = source["Description"];
	        this.Order = source["Order"];
	        this.Resources = this.convertValues(source["Resources"], StepResource);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Process {
	    Id: number;
	    Name: string;
	    Description: string;
	    Steps: Step[];
	
	    static createFrom(source: any = {}) {
	        return new Process(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Id = source["Id"];
	        this.Name = source["Name"];
	        this.Description = source["Description"];
	        this.Steps = this.convertValues(source["Steps"], Step);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class BigProcess {
	    Id: number;
	    Name: string;
	    Description: string;
	    Processes: Process[];
	
	    static createFrom(source: any = {}) {
	        return new BigProcess(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Id = source["Id"];
	        this.Name = source["Name"];
	        this.Description = source["Description"];
	        this.Processes = this.convertValues(source["Processes"], Process);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class RoleTablePermission {
	    Id: number;
	    TableId: number;
	    InsertPermission: boolean;
	    DeletePermission: boolean;
	    UpdatePermission: boolean;
	    ViewPermission: boolean;
	
	    static createFrom(source: any = {}) {
	        return new RoleTablePermission(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Id = source["Id"];
	        this.TableId = source["TableId"];
	        this.InsertPermission = source["InsertPermission"];
	        this.DeletePermission = source["DeletePermission"];
	        this.UpdatePermission = source["UpdatePermission"];
	        this.ViewPermission = source["ViewPermission"];
	    }
	}
	export class ProcessPermission {
	    Id: number;
	    ProcessId: number;
	
	    static createFrom(source: any = {}) {
	        return new ProcessPermission(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Id = source["Id"];
	        this.ProcessId = source["ProcessId"];
	    }
	}
	export class Role {
	    Id: number;
	    Name: string;
	    Description: string;
	    ProcessPermissions: ProcessPermission[];
	    TablePermissions: RoleTablePermission[];
	
	    static createFrom(source: any = {}) {
	        return new Role(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Id = source["Id"];
	        this.Name = source["Name"];
	        this.Description = source["Description"];
	        this.ProcessPermissions = this.convertValues(source["ProcessPermissions"], ProcessPermission);
	        this.TablePermissions = this.convertValues(source["TablePermissions"], RoleTablePermission);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Relation {
	    Id: number;
	    IdEntity1: number;
	    IdEntity2: number;
	    Relation: string;
	
	    static createFrom(source: any = {}) {
	        return new Relation(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Id = source["Id"];
	        this.IdEntity1 = source["IdEntity1"];
	        this.IdEntity2 = source["IdEntity2"];
	        this.Relation = source["Relation"];
	    }
	}
	export class IntersectionEntity {
	    RelationID: number;
	    Entity: Entity;
	
	    static createFrom(source: any = {}) {
	        return new IntersectionEntity(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.RelationID = source["RelationID"];
	        this.Entity = this.convertValues(source["Entity"], Entity);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Entity {
	    Id: number;
	    Name: string;
	    Description: string;
	    Attributes: Attribute[];
	    Status?: boolean;
	    TableType: string;
	
	    static createFrom(source: any = {}) {
	        return new Entity(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Id = source["Id"];
	        this.Name = source["Name"];
	        this.Description = source["Description"];
	        this.Attributes = this.convertValues(source["Attributes"], Attribute);
	        this.Status = source["Status"];
	        this.TableType = source["TableType"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class DbProject {
	    Name: string;
	    Entities: Entity[];
	    EntitiesLastMax: number;
	    IntersectionEntities: IntersectionEntity[];
	    IntersectionEntitiesLastMax: number;
	    Relations: Relation[];
	    BigProcesses: BigProcess[];
	    Roles: Role[];
	    RelationsLastMax: number;
	    AttributesLastMax: number;
	    BigProcessLastMax: number;
	    ProcessLastMax: number;
	    StepsLastMax: number;
	    StepResourceLastMax: number;
	    RoleLastMax: number;
	    ProcessPermissionLastMax: number;
	    RoleTablePermissionLastMax: number;
	
	    static createFrom(source: any = {}) {
	        return new DbProject(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Name = source["Name"];
	        this.Entities = this.convertValues(source["Entities"], Entity);
	        this.EntitiesLastMax = source["EntitiesLastMax"];
	        this.IntersectionEntities = this.convertValues(source["IntersectionEntities"], IntersectionEntity);
	        this.IntersectionEntitiesLastMax = source["IntersectionEntitiesLastMax"];
	        this.Relations = this.convertValues(source["Relations"], Relation);
	        this.BigProcesses = this.convertValues(source["BigProcesses"], BigProcess);
	        this.Roles = this.convertValues(source["Roles"], Role);
	        this.RelationsLastMax = source["RelationsLastMax"];
	        this.AttributesLastMax = source["AttributesLastMax"];
	        this.BigProcessLastMax = source["BigProcessLastMax"];
	        this.ProcessLastMax = source["ProcessLastMax"];
	        this.StepsLastMax = source["StepsLastMax"];
	        this.StepResourceLastMax = source["StepResourceLastMax"];
	        this.RoleLastMax = source["RoleLastMax"];
	        this.ProcessPermissionLastMax = source["ProcessPermissionLastMax"];
	        this.RoleTablePermissionLastMax = source["RoleTablePermissionLastMax"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	
	
	
	
	export class RelationViewItem {
	    Id?: number;
	    Entity2: string;
	    IdEntity2: number;
	    Relation?: string;
	
	    static createFrom(source: any = {}) {
	        return new RelationViewItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Id = source["Id"];
	        this.Entity2 = source["Entity2"];
	        this.IdEntity2 = source["IdEntity2"];
	        this.Relation = source["Relation"];
	    }
	}
	export class RelationView {
	    PrincipalEntity: string;
	    IdPrincipalEntity: number;
	    Relations: RelationViewItem[];
	
	    static createFrom(source: any = {}) {
	        return new RelationView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.PrincipalEntity = source["PrincipalEntity"];
	        this.IdPrincipalEntity = source["IdPrincipalEntity"];
	        this.Relations = this.convertValues(source["Relations"], RelationViewItem);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	
	
	export class SQLGenerationResult {
	    Database: string;
	    Model: string;
	    SQL: string;
	    GeneratedScript: string;
	    AIError: string;
	    ExportJSON: string;
	
	    static createFrom(source: any = {}) {
	        return new SQLGenerationResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Database = source["Database"];
	        this.Model = source["Model"];
	        this.SQL = source["SQL"];
	        this.GeneratedScript = source["GeneratedScript"];
	        this.AIError = source["AIError"];
	        this.ExportJSON = source["ExportJSON"];
	    }
	}
	

}

