export namespace cleaner {
	
	export class CleanResult {
	    Path: string;
	    Size: number;
	    Success: boolean;
	    Error: any;
	    WasDryRun: boolean;
	
	    static createFrom(source: any = {}) {
	        return new CleanResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Path = source["Path"];
	        this.Size = source["Size"];
	        this.Success = source["Success"];
	        this.Error = source["Error"];
	        this.WasDryRun = source["WasDryRun"];
	    }
	}

}

export namespace services {
	
	export class Settings {
	    theme: string;
	    defaultView: string;
	    autoScan: boolean;
	    confirmDelete: boolean;
	    scanCategories: string[];
	    maxDepth: number;
	    checkAutoUpdate: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Settings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.theme = source["theme"];
	        this.defaultView = source["defaultView"];
	        this.autoScan = source["autoScan"];
	        this.confirmDelete = source["confirmDelete"];
	        this.scanCategories = source["scanCategories"];
	        this.maxDepth = source["maxDepth"];
	        this.checkAutoUpdate = source["checkAutoUpdate"];
	    }
	}
	export class UpdateInfo {
	    available: boolean;
	    currentVersion: string;
	    latestVersion: string;
	    releaseURL: string;
	    releaseNotes: string;
	    // Go type: time
	    publishedAt: any;
	
	    static createFrom(source: any = {}) {
	        return new UpdateInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.available = source["available"];
	        this.currentVersion = source["currentVersion"];
	        this.latestVersion = source["latestVersion"];
	        this.releaseURL = source["releaseURL"];
	        this.releaseNotes = source["releaseNotes"];
	        this.publishedAt = this.convertValues(source["publishedAt"], null);
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

export namespace types {
	
	export class ScanOptions {
	    IncludeXcode: boolean;
	    IncludeAndroid: boolean;
	    IncludeNode: boolean;
	    IncludeReactNative: boolean;
	    IncludeFlutter: boolean;
	    IncludeCache: boolean;
	    IncludePython: boolean;
	    IncludeRust: boolean;
	    IncludeGo: boolean;
	    IncludeHomebrew: boolean;
	    IncludeDocker: boolean;
	    IncludeJava: boolean;
	    MaxDepth: number;
	    ProjectRoot: string;
	
	    static createFrom(source: any = {}) {
	        return new ScanOptions(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.IncludeXcode = source["IncludeXcode"];
	        this.IncludeAndroid = source["IncludeAndroid"];
	        this.IncludeNode = source["IncludeNode"];
	        this.IncludeReactNative = source["IncludeReactNative"];
	        this.IncludeFlutter = source["IncludeFlutter"];
	        this.IncludeCache = source["IncludeCache"];
	        this.IncludePython = source["IncludePython"];
	        this.IncludeRust = source["IncludeRust"];
	        this.IncludeGo = source["IncludeGo"];
	        this.IncludeHomebrew = source["IncludeHomebrew"];
	        this.IncludeDocker = source["IncludeDocker"];
	        this.IncludeJava = source["IncludeJava"];
	        this.MaxDepth = source["MaxDepth"];
	        this.ProjectRoot = source["ProjectRoot"];
	    }
	}
	export class ScanResult {
	    path: string;
	    type: string;
	    size: number;
	    fileCount: number;
	    name: string;
	
	    static createFrom(source: any = {}) {
	        return new ScanResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.type = source["type"];
	        this.size = source["size"];
	        this.fileCount = source["fileCount"];
	        this.name = source["name"];
	    }
	}
	export class TreeNode {
	    Path: string;
	    Name: string;
	    Size: number;
	    IsDir: boolean;
	    Type: string;
	    Children: TreeNode[];
	    Scanned: boolean;
	    Depth: number;
	    FileCount: number;
	
	    static createFrom(source: any = {}) {
	        return new TreeNode(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Path = source["Path"];
	        this.Name = source["Name"];
	        this.Size = source["Size"];
	        this.IsDir = source["IsDir"];
	        this.Type = source["Type"];
	        this.Children = this.convertValues(source["Children"], TreeNode);
	        this.Scanned = source["Scanned"];
	        this.Depth = source["Depth"];
	        this.FileCount = source["FileCount"];
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

