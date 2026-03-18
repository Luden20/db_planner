export namespace utils {
	
	export class Attribute {
	    Id: number;
	    Name: string;
	    Description: string;
	    Type: string;
	
	    static createFrom(source: any = {}) {
	        return new Attribute(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Id = source["Id"];
	        this.Name = source["Name"];
	        this.Description = source["Description"];
	        this.Type = source["Type"];
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
	export class Entity {
	    Id: number;
	    Name: string;
	    Description: string;
	    Attributes: Attribute[];
	    Status?: boolean;
	
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
	    Relations: Relation[];
	    RelationsLastMax: number;
	    AttributesLastMax: number;
	
	    static createFrom(source: any = {}) {
	        return new DbProject(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Name = source["Name"];
	        this.Entities = this.convertValues(source["Entities"], Entity);
	        this.EntitiesLastMax = source["EntitiesLastMax"];
	        this.Relations = this.convertValues(source["Relations"], Relation);
	        this.RelationsLastMax = source["RelationsLastMax"];
	        this.AttributesLastMax = source["AttributesLastMax"];
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

}

