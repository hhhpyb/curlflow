export namespace domain {
	
	export class HttpRequest {
	    url: string;
	    method: string;
	    headers: Record<string, string>;
	    body: string;
	    compressed: boolean;
	
	    static createFrom(source: any = {}) {
	        return new HttpRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.url = source["url"];
	        this.method = source["method"];
	        this.headers = source["headers"];
	        this.body = source["body"];
	        this.compressed = source["compressed"];
	    }
	}
	export class HttpResponse {
	    statusCode: number;
	    time: number;
	    body: string;
	    headers: Record<string, string>;
	    error: string;
	
	    static createFrom(source: any = {}) {
	        return new HttpResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.statusCode = source["statusCode"];
	        this.time = source["time"];
	        this.body = source["body"];
	        this.headers = source["headers"];
	        this.error = source["error"];
	    }
	}
	export class MetaData {
	    id: string;
	    key: string;
	    status: string;
	    summary: string;
	    description: string;
	    tags: string[];
	    swagger_path: string;
	    last_synced_at: number;
	    param_docs: Record<string, string>;
	
	    static createFrom(source: any = {}) {
	        return new MetaData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.key = source["key"];
	        this.status = source["status"];
	        this.summary = source["summary"];
	        this.description = source["description"];
	        this.tags = source["tags"];
	        this.swagger_path = source["swagger_path"];
	        this.last_synced_at = source["last_synced_at"];
	        this.param_docs = source["param_docs"];
	    }
	}
	export class RequestFile {
	    _meta: MetaData;
	    data: HttpRequest;
	
	    static createFrom(source: any = {}) {
	        return new RequestFile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this._meta = this.convertValues(source["_meta"], MetaData);
	        this.data = this.convertValues(source["data"], HttpRequest);
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

export namespace main {
	
	export class AppConfig {
	    proxyUrl: string;
	    insecure: boolean;
	    timeout: number;
	
	    static createFrom(source: any = {}) {
	        return new AppConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.proxyUrl = source["proxyUrl"];
	        this.insecure = source["insecure"];
	        this.timeout = source["timeout"];
	    }
	}

}

