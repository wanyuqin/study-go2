export namespace configs {
	
	export class DownloadConfig {
	    path: string;
	
	    static createFrom(source: any = {}) {
	        return new DownloadConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	    }
	}

}

export namespace main {
	
	export class NcmFile {
	    name: string;
	    path: string;
	    mod_time: string;
	    size: string;
	
	    static createFrom(source: any = {}) {
	        return new NcmFile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.path = source["path"];
	        this.mod_time = source["mod_time"];
	        this.size = source["size"];
	    }
	}

}

export namespace tools {
	
	export class ExtractLinkData {
	    id: string;
	    title: string;
	    type: string;
	    url: string;
	    quality: string;
	    size: string;
	    byte: number;
	    percentage: number;
	
	    static createFrom(source: any = {}) {
	        return new ExtractLinkData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.title = source["title"];
	        this.type = source["type"];
	        this.url = source["url"];
	        this.quality = source["quality"];
	        this.size = source["size"];
	        this.byte = source["byte"];
	        this.percentage = source["percentage"];
	    }
	}

}

