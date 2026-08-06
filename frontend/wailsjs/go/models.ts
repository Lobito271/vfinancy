export namespace bindings {
	
	export class App {
	
	
	    static createFrom(source: any = {}) {
	        return new App(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	
	    }
	}
	export class AppSettingsDTO {
	    windowTitle: string;
	    width: number;
	    height: number;
	    logLevel: string;
	    logFormat: string;
	
	    static createFrom(source: any = {}) {
	        return new AppSettingsDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.windowTitle = source["windowTitle"];
	        this.width = source["width"];
	        this.height = source["height"];
	        this.logLevel = source["logLevel"];
	        this.logFormat = source["logFormat"];
	    }
	}
	export class AuditEventDTO {
	    id: string;
	    eventType: string;
	    userId: string;
	    description: string;
	    ipAddress: string;
	    device: string;
	    occurredAt: string;
	
	    static createFrom(source: any = {}) {
	        return new AuditEventDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.eventType = source["eventType"];
	        this.userId = source["userId"];
	        this.description = source["description"];
	        this.ipAddress = source["ipAddress"];
	        this.device = source["device"];
	        this.occurredAt = source["occurredAt"];
	    }
	}
	export class AuditLogResult {
	    events: AuditEventDTO[];
	    total: number;
	
	    static createFrom(source: any = {}) {
	        return new AuditLogResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.events = this.convertValues(source["events"], AuditEventDTO);
	        this.total = source["total"];
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
	export class BusinessInfoDTO {
	    name: string;
	    tradeName: string;
	    taxId: string;
	    address: string;
	    phone: string;
	    email: string;
	    logo: string;
	
	    static createFrom(source: any = {}) {
	        return new BusinessInfoDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.tradeName = source["tradeName"];
	        this.taxId = source["taxId"];
	        this.address = source["address"];
	        this.phone = source["phone"];
	        this.email = source["email"];
	        this.logo = source["logo"];
	    }
	}
	export class ConnectionConfigDTO {
	    host: string;
	    port: number;
	    user: string;
	    password: string;
	    database: string;
	    sslMode: string;
	
	    static createFrom(source: any = {}) {
	        return new ConnectionConfigDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.host = source["host"];
	        this.port = source["port"];
	        this.user = source["user"];
	        this.password = source["password"];
	        this.database = source["database"];
	        this.sslMode = source["sslMode"];
	    }
	}
	export class CurrencyDTO {
	    code: string;
	    symbol: string;
	    name: string;
	    decimalPlaces: number;
	    type: string;
	    isActive: boolean;
	
	    static createFrom(source: any = {}) {
	        return new CurrencyDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.code = source["code"];
	        this.symbol = source["symbol"];
	        this.name = source["name"];
	        this.decimalPlaces = source["decimalPlaces"];
	        this.type = source["type"];
	        this.isActive = source["isActive"];
	    }
	}
	export class LoginRequest {
	    username: string;
	    password: string;
	    remember: boolean;
	
	    static createFrom(source: any = {}) {
	        return new LoginRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.username = source["username"];
	        this.password = source["password"];
	        this.remember = source["remember"];
	    }
	}
	export class LoginResponse {
	    token: string;
	    expiresAt: string;
	    userId: string;
	    fullName: string;
	    email: string;
	    username: string;
	    roles: string[];
	    companyId: string;
	    mustChangePassword: boolean;
	
	    static createFrom(source: any = {}) {
	        return new LoginResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.token = source["token"];
	        this.expiresAt = source["expiresAt"];
	        this.userId = source["userId"];
	        this.fullName = source["fullName"];
	        this.email = source["email"];
	        this.username = source["username"];
	        this.roles = source["roles"];
	        this.companyId = source["companyId"];
	        this.mustChangePassword = source["mustChangePassword"];
	    }
	}
	export class ModuleDTO {
	    id: string;
	    name: string;
	    description: string;
	    enabled: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ModuleDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.enabled = source["enabled"];
	    }
	}
	export class PreferencesDTO {
	    defaultCurrency: string;
	    defaultTaxCode: string;
	    expiryAlertDays: number;
	    defaultCountry: string;
	    dateFormat: string;
	    numberFormat: string;
	    decimalPlaces: number;
	    language: string;
	    theme: string;
	    timezone: string;
	    fiscalYearStart: number;
	    backupFolder: string;
	    exportFolder: string;
	    backupFrequency: string;
	
	    static createFrom(source: any = {}) {
	        return new PreferencesDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.defaultCurrency = source["defaultCurrency"];
	        this.defaultTaxCode = source["defaultTaxCode"];
	        this.expiryAlertDays = source["expiryAlertDays"];
	        this.defaultCountry = source["defaultCountry"];
	        this.dateFormat = source["dateFormat"];
	        this.numberFormat = source["numberFormat"];
	        this.decimalPlaces = source["decimalPlaces"];
	        this.language = source["language"];
	        this.theme = source["theme"];
	        this.timezone = source["timezone"];
	        this.fiscalYearStart = source["fiscalYearStart"];
	        this.backupFolder = source["backupFolder"];
	        this.exportFolder = source["exportFolder"];
	        this.backupFrequency = source["backupFrequency"];
	    }
	}
	export class ProfileDTO {
	    userId: string;
	    fullName: string;
	    username: string;
	    email: string;
	    avatarUrl: string;
	    theme: string;
	    language: string;
	    dateFormat: string;
	    numberFormat: string;
	    decimalPlaces: number;
	    timezone: string;
	    lastLoginAt: string;
	    isActive: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ProfileDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.userId = source["userId"];
	        this.fullName = source["fullName"];
	        this.username = source["username"];
	        this.email = source["email"];
	        this.avatarUrl = source["avatarUrl"];
	        this.theme = source["theme"];
	        this.language = source["language"];
	        this.dateFormat = source["dateFormat"];
	        this.numberFormat = source["numberFormat"];
	        this.decimalPlaces = source["decimalPlaces"];
	        this.timezone = source["timezone"];
	        this.lastLoginAt = source["lastLoginAt"];
	        this.isActive = source["isActive"];
	    }
	}
	export class TaxDTO {
	    id: string;
	    code: string;
	    name: string;
	    shortName: string;
	    countryCode: string;
	    defaultRate: number;
	    isInclusive: boolean;
	    isPercentage: boolean;
	    category: string;
	    isActive: boolean;
	
	    static createFrom(source: any = {}) {
	        return new TaxDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.code = source["code"];
	        this.name = source["name"];
	        this.shortName = source["shortName"];
	        this.countryCode = source["countryCode"];
	        this.defaultRate = source["defaultRate"];
	        this.isInclusive = source["isInclusive"];
	        this.isPercentage = source["isPercentage"];
	        this.category = source["category"];
	        this.isActive = source["isActive"];
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

