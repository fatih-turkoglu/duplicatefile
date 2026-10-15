export namespace db {
	
	export class File {
	    id: number;
	    path: string;
	    size: number;
	    modTime: number;
	    fullHash: string;
	    width?: number;
	    height?: number;
	    similarity?: number;
	
	    static createFrom(source: any = {}) {
	        return new File(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.path = source["path"];
	        this.size = source["size"];
	        this.modTime = source["modTime"];
	        this.fullHash = source["fullHash"];
	        this.width = source["width"];
	        this.height = source["height"];
	        this.similarity = source["similarity"];
	    }
	}
	export class DuplicateGroup {
	    hash: string;
	    size: number;
	    files: File[];
	    similar?: boolean;
	
	    static createFrom(source: any = {}) {
	        return new DuplicateGroup(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.hash = source["hash"];
	        this.size = source["size"];
	        this.files = this.convertValues(source["files"], File);
	        this.similar = source["similar"];
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
	
	export class Stats {
	    files: number;
	    scans: number;
	    sizeBytes: number;
	
	    static createFrom(source: any = {}) {
	        return new Stats(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.files = source["files"];
	        this.scans = source["scans"];
	        this.sizeBytes = source["sizeBytes"];
	    }
	}

}

export namespace fileops {
	
	export class DeleteOptions {
	    permanent: boolean;
	    similar: boolean;
	    threshold: number;
	
	    static createFrom(source: any = {}) {
	        return new DeleteOptions(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.permanent = source["permanent"];
	        this.similar = source["similar"];
	        this.threshold = source["threshold"];
	    }
	}
	export class RenameItem {
	    path: string;
	    newName: string;
	
	    static createFrom(source: any = {}) {
	        return new RenameItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.newName = source["newName"];
	    }
	}
	export class Result {
	    path: string;
	    newPath?: string;
	    error?: string;
	    code?: string;
	    warning?: string;
	
	    static createFrom(source: any = {}) {
	        return new Result(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.newPath = source["newPath"];
	        this.error = source["error"];
	        this.code = source["code"];
	        this.warning = source["warning"];
	    }
	}
	export class Report {
	    succeeded: number;
	    failed: number;
	    results: Result[];
	
	    static createFrom(source: any = {}) {
	        return new Report(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.succeeded = source["succeeded"];
	        this.failed = source["failed"];
	        this.results = this.convertValues(source["results"], Result);
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

export namespace i18n {
	
	export class File {
	    code: string;
	    name: string;
	    locale: string;
	    entries: Record<string, string>;
	    warnings: string[];
	
	    static createFrom(source: any = {}) {
	        return new File(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.code = source["code"];
	        this.name = source["name"];
	        this.locale = source["locale"];
	        this.entries = source["entries"];
	        this.warnings = source["warnings"];
	    }
	}
	export class Info {
	    code: string;
	    name: string;
	    locale: string;
	
	    static createFrom(source: any = {}) {
	        return new Info(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.code = source["code"];
	        this.name = source["name"];
	        this.locale = source["locale"];
	    }
	}

}

export namespace main {
	
	export class AppInfo {
	    version: string;
	    goVersion: string;
	
	    static createFrom(source: any = {}) {
	        return new AppInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.version = source["version"];
	        this.goVersion = source["goVersion"];
	    }
	}
	export class CompactResult {
	    before: number;
	    after: number;
	
	    static createFrom(source: any = {}) {
	        return new CompactResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.before = source["before"];
	        this.after = source["after"];
	    }
	}
	export class DatabaseInfo {
	    path: string;
	    isDefault: boolean;
	    warning?: string;
	    stats: db.Stats;
	
	    static createFrom(source: any = {}) {
	        return new DatabaseInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.isDefault = source["isDefault"];
	        this.warning = source["warning"];
	        this.stats = this.convertValues(source["stats"], db.Stats);
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
	export class LanguageSetup {
	    dir: string;
	    available: i18n.Info[];
	    system: string[];
	
	    static createFrom(source: any = {}) {
	        return new LanguageSetup(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.dir = source["dir"];
	        this.available = this.convertValues(source["available"], i18n.Info);
	        this.system = source["system"];
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

export namespace scanner {
	
	export class Options {
	    roots: string[];
	    minSize: number;
	    extensions: string[];
	    excludeDirs: string[];
	    skipHidden: boolean;
	    workers: number;
	    mode: string;
	    threshold: number;
	
	    static createFrom(source: any = {}) {
	        return new Options(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.roots = source["roots"];
	        this.minSize = source["minSize"];
	        this.extensions = source["extensions"];
	        this.excludeDirs = source["excludeDirs"];
	        this.skipHidden = source["skipHidden"];
	        this.workers = source["workers"];
	        this.mode = source["mode"];
	        this.threshold = source["threshold"];
	    }
	}

}

