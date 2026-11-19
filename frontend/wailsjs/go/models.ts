export namespace main {
	
	export class ConditionDTO {
	    kind: string;
	    value: string;
	
	    static createFrom(source: any = {}) {
	        return new ConditionDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.kind = source["kind"];
	        this.value = source["value"];
	    }
	}
	export class EntryDTO {
	    key: string;
	    value: string;
	    rawLine?: string;
	
	    static createFrom(source: any = {}) {
	        return new EntryDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.key = source["key"];
	        this.value = source["value"];
	        this.rawLine = source["rawLine"];
	    }
	}
	export class BlockDTO {
	    id: string;
	    type: string;
	    schemaName?: string;
	    subsection?: string;
	    entries?: EntryDTO[];
	    path?: string;
	    children?: BlockDTO[];
	    conditions?: ConditionDTO[];
	    file?: BlockDTO;
	
	    static createFrom(source: any = {}) {
	        return new BlockDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.type = source["type"];
	        this.schemaName = source["schemaName"];
	        this.subsection = source["subsection"];
	        this.entries = this.convertValues(source["entries"], EntryDTO);
	        this.path = source["path"];
	        this.children = this.convertValues(source["children"], BlockDTO);
	        this.conditions = this.convertValues(source["conditions"], ConditionDTO);
	        this.file = this.convertValues(source["file"], BlockDTO);
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
	
	
	export class FieldDTO {
	    key: string;
	    label: string;
	    type: string;
	    options?: string[];
	    default?: string;
	    common: boolean;
	    docs?: string;
	
	    static createFrom(source: any = {}) {
	        return new FieldDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.key = source["key"];
	        this.label = source["label"];
	        this.type = source["type"];
	        this.options = source["options"];
	        this.default = source["default"];
	        this.common = source["common"];
	        this.docs = source["docs"];
	    }
	}
	export class SectionSchemaDTO {
	    name: string;
	    label: string;
	    icon: string;
	    hasSubsection: boolean;
	    subsectionLabel?: string;
	    subsectionPlaceholder?: string;
	    fields: FieldDTO[];
	    freeForm: boolean;
	
	    static createFrom(source: any = {}) {
	        return new SectionSchemaDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.label = source["label"];
	        this.icon = source["icon"];
	        this.hasSubsection = source["hasSubsection"];
	        this.subsectionLabel = source["subsectionLabel"];
	        this.subsectionPlaceholder = source["subsectionPlaceholder"];
	        this.fields = this.convertValues(source["fields"], FieldDTO);
	        this.freeForm = source["freeForm"];
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

