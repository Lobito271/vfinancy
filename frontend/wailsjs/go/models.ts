export namespace bindings {
	
	export class App {
	
	
	    static createFrom(source: any = {}) {
	        return new App(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	
	    }
	}

}

export namespace config {
	
	export class AppConfig {
	    Name: string;
	    Env: string;
	    Version: string;
	    Port: number;
	    WindowTitle: string;
	    Width: number;
	    Height: number;
	
	    static createFrom(source: any = {}) {
	        return new AppConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Name = source["Name"];
	        this.Env = source["Env"];
	        this.Version = source["Version"];
	        this.Port = source["Port"];
	        this.WindowTitle = source["WindowTitle"];
	        this.Width = source["Width"];
	        this.Height = source["Height"];
	    }
	}
	export class AuthConfig {
	    ArgonMemory: number;
	    ArgonIterations: number;
	    ArgonParallelism: number;
	    ArgonSaltLength: number;
	    ArgonKeyLength: number;
	    SessionTTL: number;
	    LockoutTTL: number;
	    MaxLoginAttempts: number;
	
	    static createFrom(source: any = {}) {
	        return new AuthConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ArgonMemory = source["ArgonMemory"];
	        this.ArgonIterations = source["ArgonIterations"];
	        this.ArgonParallelism = source["ArgonParallelism"];
	        this.ArgonSaltLength = source["ArgonSaltLength"];
	        this.ArgonKeyLength = source["ArgonKeyLength"];
	        this.SessionTTL = source["SessionTTL"];
	        this.LockoutTTL = source["LockoutTTL"];
	        this.MaxLoginAttempts = source["MaxLoginAttempts"];
	    }
	}
	export class LoggerConfig {
	    Level: string;
	    Format: string;
	    Output: string;
	
	    static createFrom(source: any = {}) {
	        return new LoggerConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Level = source["Level"];
	        this.Format = source["Format"];
	        this.Output = source["Output"];
	    }
	}
	export class DatabaseConfig {
	    Host: string;
	    Port: number;
	    User: string;
	    Password: string;
	    Name: string;
	    SSLMode: string;
	    MaxOpen: number;
	    MaxIdle: number;
	    MaxLifetime: number;
	    MigrationDir: string;
	
	    static createFrom(source: any = {}) {
	        return new DatabaseConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Host = source["Host"];
	        this.Port = source["Port"];
	        this.User = source["User"];
	        this.Password = source["Password"];
	        this.Name = source["Name"];
	        this.SSLMode = source["SSLMode"];
	        this.MaxOpen = source["MaxOpen"];
	        this.MaxIdle = source["MaxIdle"];
	        this.MaxLifetime = source["MaxLifetime"];
	        this.MigrationDir = source["MigrationDir"];
	    }
	}
	export class Config {
	    App: AppConfig;
	    Database: DatabaseConfig;
	    Logger: LoggerConfig;
	    Auth: AuthConfig;
	
	    static createFrom(source: any = {}) {
	        return new Config(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.App = this.convertValues(source["App"], AppConfig);
	        this.Database = this.convertValues(source["Database"], DatabaseConfig);
	        this.Logger = this.convertValues(source["Logger"], LoggerConfig);
	        this.Auth = this.convertValues(source["Auth"], AuthConfig);
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

export namespace logger {
	
	export class Logger {
	
	
	    static createFrom(source: any = {}) {
	        return new Logger(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	
	    }
	}

}

