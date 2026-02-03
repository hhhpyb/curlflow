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

}

