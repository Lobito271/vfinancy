export interface BusinessInfoDTO {
  name: string;
  tradeName: string;
  taxId: string;
  address: string;
  phone: string;
  email: string;
  logo: string;
}

export interface LocalAuthStateDTO {
  configured: boolean;
  passwordEnabled: boolean;
  unlocked: boolean;
}

export interface LocalProfileDTO {
  id: string;
  name: string;
  passwordEnabled: boolean;
  activeCompanyId: string;
  theme: string;
  language: string;
  dateFormat: string;
  numberFormat: string;
  decimalPlaces: number;
  timezone: string;
}

export interface CompanyDTO {
  id: string;
  code: string;
  legalName: string;
  tradeName: string;
  taxId: string;
  address: string;
  phone: string;
  email: string;
  countryCode: string;
  functionalCurrency: string;
  timezone: string;
  fiscalYearStartMonth: number;
  isActive: boolean;
}

export interface CreateLocalProfileRequest {
  name: string;
  companyId: string;
}

export interface UpdateLocalProfileRequest {
  name: string;
  theme: string;
  language: string;
  dateFormat: string;
  numberFormat: string;
  decimalPlaces: number;
  timezone: string;
}

export interface CompanyRequest extends CompanyDTO {}

export interface PreferencesDTO {
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
  clearanceDays: number;
  clearanceWarningDays: number;
  saleNumberPrefix: string;
  purchaseNumberPrefix: string;
  journalNumberPrefix: string;
}

export interface CurrencyDTO {
  code: string;
  symbol: string;
  name: string;
  decimalPlaces: number;
  type: string;
  isActive: boolean;
}

export interface TaxDTO {
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
}

export interface AuditEventDTO {
  id: string;
  eventType: string;
  userId: string;
  description: string;
  ipAddress: string;
  device: string;
  occurredAt: string;
}

export interface AuditLogResult {
  events: AuditEventDTO[];
  total: number;
}

export interface ConnectionConfigDTO {
  host: string;
  port: number;
  user: string;
  password: string;
  database: string;
  sslMode: string;
}

export interface AppSettingsDTO {
  windowTitle: string;
  width: number;
  height: number;
  logLevel: string;
  logFormat: string;
}

export interface ModuleDTO {
  id: string;
  name: string;
  description: string;
  enabled: boolean;
}

export interface PaginationRequest {
  page: number;
  pageSize: number;
}

export interface PageResult<T = unknown> {
  items: T[];
  total: number;
  page: number;
  pageSize: number;
}

export interface CustomerDTO {
  id: string;
  documentType: string;
  documentNumber: string;
  businessName: string;
  tradeName: string;
  creditLimit: string;
  currentDebt: string;
  paymentTermDays: number;
  status: string;
  blockedReason: string;
  email: string;
  phone: string;
  address: string;
  createdAt: string;
}

export interface ListCustomersRequest extends PaginationRequest {
  search: string;
  status: string;
}

export interface CreateCustomerRequest {
  documentType: string;
  documentNumber: string;
  businessName: string;
  tradeName: string;
  creditLimit: string;
  paymentTermDays: number;
  email: string;
  phone: string;
  address: string;
}

export interface UpdateCustomerRequest {
  id: string;
  businessName: string;
  tradeName: string;
  creditLimit: string;
  paymentTermDays: number;
  email: string;
  phone: string;
  address: string;
  isActive?: boolean | null;
}

export interface ProductDTO {
  id: string;
  sku: string;
  barcode: string;
  description: string;
  categoryId: string;
  brandId: string;
  unitId: string;
  taxId: string;
  category: string;
  brand: string;
  unit: string;
  cost: string;
  salePrice: string;
  saleCurrency: string;
  minStock: string;
  maxStock: string;
  weight: string;
  isActive: boolean;
  isService: boolean;
  createdAt: string;
}

export interface UnitDTO {
  id: string;
  code: string;
  name: string;
  symbol: string;
  allowsDecimals: boolean;
}

export interface CategoryDTO {
  id: string;
  code: string;
  name: string;
}

export interface BrandDTO {
  id: string;
  code: string;
  name: string;
}

export interface CategoryCodeNameRequest {
  code: string;
  name: string;
}

export interface CreateCategoryRequest extends CategoryCodeNameRequest {}

export interface UpdateCategoryRequest extends CategoryCodeNameRequest {
  id: string;
}

export interface CreateBrandRequest extends CategoryCodeNameRequest {}

export interface UpdateBrandRequest extends CategoryCodeNameRequest {
  id: string;
}

export interface ListProductsRequest extends PaginationRequest {
  search: string;
  status: string;
}

export interface CreateProductRequest {
  sku: string;
  barcode: string;
  description: string;
  categoryId: string;
  brandId: string;
  unitId: string;
  taxId: string;
  cost: string;
  salePrice: string;
  saleCurrency: string;
  minStock: string;
  maxStock: string;
  weight: string;
  isService: boolean;
}

export interface UpdateProductRequest {
  id: string;
  description: string;
  categoryId: string;
  brandId: string;
  unitId: string;
  cost: string;
  salePrice: string;
  minStock: string;
  maxStock: string;
  isActive?: boolean | null;
}

export interface SupplierDTO {
  id: string;
  documentType: string;
  documentNumber: string;
  businessName: string;
  tradeName: string;
  taxId: string;
  isInternational: boolean;
  defaultCurrency: string;
  currentDebt: string;
  paymentTermDays: number;
  status: string;
  email: string;
  phone: string;
  address: string;
  createdAt: string;
}

export interface ListSuppliersRequest extends PaginationRequest {
  search: string;
  status: string;
}

export interface CreateSupplierRequest {
  documentType: string;
  documentNumber: string;
  businessName: string;
  tradeName: string;
  taxId: string;
  isInternational: boolean;
  defaultCurrency: string;
  paymentTermDays: number;
  email: string;
  phone: string;
  address: string;
}

export interface UpdateSupplierRequest {
  id: string;
  businessName: string;
  tradeName: string;
  taxId: string;
  paymentTermDays: number;
  email: string;
  phone: string;
  address: string;
  isActive?: boolean | null;
}

export interface SaleItemDTO {
  id: string;
  productId: string;
  lineNumber: number;
  quantity: string;
  unitPrice: string;
  discountPercent: string;
  discountAmount: string;
  taxRate: string;
  taxAmount: string;
  costSnapshot: string;
  description: string;
}

export interface SaleDTO {
  id: string;
  number: string;
  customerId: string;
  customerName: string;
  date: string;
  status: string;
  subtotal: string;
  tax: string;
  discount: string;
  total: string;
  cost: string;
  profit: string;
  paid: string;
  balance: string;
  items: SaleItemDTO[];
}

export interface ListSalesRequest extends PaginationRequest {
  customerId: string;
  status: string;
}

export interface CreateSaleItemRequest {
  productId: string;
  quantity: string;
  unitPrice: string;
  discountPercent: string;
  discountAmount: string;
  taxRate: string;
  taxAmount: string;
  costSnapshot: string;
  description: string;
}

export interface CreateSaleRequest {
  customerId: string;
  currencyCode: string;
  exchangeRate: string;
  dueDate: string;
  notes: string;
  items: CreateSaleItemRequest[];
}

export interface CustomerPaymentDTO {
  id: string;
  number: string;
  customerId: string;
  paymentDate: string;
  amount: string;
  currencyCode: string;
  method: string;
  status: string;
  reference: string;
  notes: string;
}

export interface CustomerAdvanceDTO {
  id: string;
  number: string;
  customerId: string;
  advanceDate: string;
  amount: string;
  currencyCode: string;
  method: string;
}

export interface ListCustomerPaymentsRequest extends PaginationRequest {
  customerId: string;
}

export interface BankAccountDTO {
  id: string;
  bankName: string;
  accountNumber: string;
  accountType: string;
  currencyCode: string;
  currentBalance: string;
  isDefault: boolean;
  isActive: boolean;
}

export interface BankTransactionDTO {
  id: string;
  accountId: string;
  date: string;
  description: string;
  amount: string;
  type: string;
  balanceAfter: string;
  reference: string;
  isReconciled: boolean;
}

export interface ListBankTransactionsRequest extends PaginationRequest {
  accountId: string;
  reconciled: boolean | null;
}

export interface CreateBankAccountRequest {
  bankName: string;
  accountNumber: string;
  accountType: string;
  currencyCode: string;
  isDefault: boolean;
}

export interface UpdateBankAccountRequest {
  id: string;
  bankName: string;
  accountNumber: string;
  accountType: string;
  currencyCode: string;
  isDefault: boolean;
  isActive: boolean;
}

export interface CreateBankTransactionRequest {
  accountId: string;
  date: string;
  description: string;
  amount: string;
  type: string;
  reference: string;
}

export interface UpsertExchangeRateRequest {
  from: string;
  to: string;
  rate: string;
  effectiveDate: string;
  source: string;
}

export interface WarehouseDTO {
  id: string;
  code: string;
  name: string;
  isDefault: boolean;
  allowsClearance: boolean;
  isActive: boolean;
}

export interface InventoryBatchDTO {
  id: string;
  productId: string;
  warehouseId: string;
  lotNumber: string;
  arrivalDate: string;
  expiryDate: string;
  initialQuantity: string;
  currentQuantity: string;
  unitCost: string;
  currencyCode: string;
  status: string;
  maxSaleDate: string;
  isClearance: boolean;
}

export interface InventoryMovementDTO {
  id: string;
  productId: string;
  warehouseId: string;
  batchId: string;
  type: string;
  quantity: string;
  unitCost: string;
  occurredAt: string;
  notes: string;
}

export interface ListInventoryBatchesRequest extends PaginationRequest {
  onlyClearance: boolean;
}

export interface ListInventoryMovementsRequest extends PaginationRequest {
  productId: string;
}

export interface ReceiveStockRequest {
  productId: string;
  warehouseId: string;
  lotNumber: string;
  arrivalDate: string;
  quantity: string;
  unitCost: string;
  currencyCode: string;
}

export interface IssueStockRequest {
  batchId: string;
  quantity: string;
}

export interface AdjustStockRequest {
  batchId: string;
  delta: string;
  reason: string;
}

export interface VoidStockRequest {
  batchId: string;
  reason: string;
}

export interface AccountDTO {
  id: string;
  code: string;
  name: string;
  type: string;
  parentId: string;
  path: string;
  depth: number;
  isActive: boolean;
  allowsMovement: boolean;
  description: string;
}

export interface FiscalPeriodDTO {
  id: string;
  name: string;
  periodStart: string;
  periodEnd: string;
  status: string;
}

export interface CreateChartOfAccountRequest {
  code: string;
  name: string;
  type: string;
  parentId: string;
  allowsMovement: boolean;
  description: string;
}

export interface UpdateChartOfAccountRequest {
  id: string;
  code: string;
  name: string;
  type: string;
  parentId: string;
  allowsMovement: boolean;
  isActive: boolean;
  description: string;
}

export interface JournalEntryLineDTO {
  id: string;
  lineNumber: number;
  accountId: string;
  description: string;
  debit: string;
  credit: string;
}

export interface JournalEntryDTO {
  id: string;
  number: string;
  entryDate: string;
  description: string;
  source: string;
  sourceId: string;
  status: string;
  lines: JournalEntryLineDTO[];
  createdAt: string;
}

export interface ListJournalEntriesRequest extends PaginationRequest {
  status: string;
}

export interface CreateJournalEntryLineRequest {
  accountId: string;
  description: string;
  debit: string;
  credit: string;
}

export interface CreateJournalEntryRequest {
  fiscalPeriodId: string;
  entryDate: string;
  description: string;
  lines: CreateJournalEntryLineRequest[];
}

export interface PurchaseItemDTO {
  id: string;
  productId: string;
  lineNumber: number;
  quantity: string;
  unitPrice: string;
  discountPercent: string;
  discountAmount: string;
  taxRate: string;
  taxAmount: string;
  description: string;
}

export interface PurchaseOrderDTO {
  id: string;
  number: string;
  supplierId: string;
  orderDate: string;
  status: string;
  subtotal: string;
  discount: string;
  tax: string;
  total: string;
  paid: string;
  notes: string;
  items: PurchaseItemDTO[];
}

export interface ListPurchaseOrdersRequest extends PaginationRequest {
  status: string;
}

export interface CreatePurchaseOrderItemRequest {
  productId: string;
  quantity: string;
  unitPrice: string;
  discountPercent: string;
  discountAmount: string;
  taxRate: string;
  taxAmount: string;
  description: string;
}

export interface CreatePurchaseOrderRequest {
  supplierId: string;
  currencyCode: string;
  exchangeRate: string;
  orderDate: string;
  notes: string;
  items: CreatePurchaseOrderItemRequest[];
}

export interface CancelPurchaseOrderRequest {
  id: string;
  reason: string;
}

export interface RegisterPurchasePaymentRequest {
  id: string;
  paymentDate: string;
  method: string;
  reference: string;
  notes: string;
}

export interface CancelSaleRequest {
  id: string;
  reason: string;
}

export interface RegisterSalePaymentRequest {
  id: string;
  paymentDate: string;
  method: string;
  reference: string;
  notes: string;
}

export interface AppBindings {
  GetLocalAuthState(): Promise<LocalAuthStateDTO>;
  GetLocalProfile(): Promise<LocalProfileDTO>;
  InitializeLocalProfile(req: CreateLocalProfileRequest): Promise<LocalProfileDTO>;
  UpdateLocalProfile(req: UpdateLocalProfileRequest): Promise<LocalProfileDTO>;
  UnlockLocalProfile(password: string): Promise<void>;
  SetLocalPassword(current: string, next: string): Promise<void>;
  RemoveLocalPassword(current: string): Promise<void>;
  LockLocalProfile(): Promise<void>;
  ListCompanies(): Promise<CompanyDTO[]>;
  GetActiveCompany(): Promise<CompanyDTO>;
  SetActiveCompany(id: string): Promise<void>;
  CreateCompany(req: CompanyRequest): Promise<CompanyDTO>;
  UpdateCompany(req: CompanyRequest): Promise<CompanyDTO>;

  GetBusinessInfo(): Promise<BusinessInfoDTO>;
  UpdateBusinessInfo(info: BusinessInfoDTO): Promise<void>;
  GetPreferences(): Promise<PreferencesDTO>;
  UpdatePreference(key: string, value: string): Promise<void>;
  GetCurrencies(): Promise<CurrencyDTO[]>;
  GetTaxes(): Promise<TaxDTO[]>;
  GetAllSettings(): Promise<Record<string, unknown>>;

  GetAuditLog(page: number, pageSize: number, eventType: string): Promise<AuditLogResult>;

  GetConnectionConfig(): Promise<ConnectionConfigDTO>;
  TestDatabaseConnection(cfg: ConnectionConfigDTO): Promise<string>;
  SaveConnectionConfig(cfg: ConnectionConfigDTO): Promise<void>;
  GetAppSettings(): Promise<AppSettingsDTO>;
  SaveAppSettings(settings: AppSettingsDTO): Promise<void>;
  GetModules(): Promise<ModuleDTO[]>;
  SetModuleEnabled(id: string, enabled: boolean): Promise<void>;

  ListCustomers(req: ListCustomersRequest): Promise<PageResult<CustomerDTO>>;
  GetCustomer(id: string): Promise<CustomerDTO>;
  CreateCustomer(req: CreateCustomerRequest): Promise<CustomerDTO>;
  UpdateCustomer(req: UpdateCustomerRequest): Promise<CustomerDTO>;
  RemoveCustomer(id: string): Promise<void>;

  ListProducts(req: ListProductsRequest): Promise<PageResult<ProductDTO>>;
  GetProduct(id: string): Promise<ProductDTO>;
  CreateProduct(req: CreateProductRequest): Promise<ProductDTO>;
  UpdateProduct(req: UpdateProductRequest): Promise<ProductDTO>;
  RemoveProduct(id: string): Promise<void>;
  ListUnits(): Promise<UnitDTO[]>;
  ListCategories(): Promise<CategoryDTO[]>;
  CreateCategory(req: CreateCategoryRequest): Promise<CategoryDTO>;
  UpdateCategory(req: UpdateCategoryRequest): Promise<CategoryDTO>;
  DeleteCategory(id: string): Promise<void>;
  ListBrands(): Promise<BrandDTO[]>;
  CreateBrand(req: CreateBrandRequest): Promise<BrandDTO>;
  UpdateBrand(req: UpdateBrandRequest): Promise<BrandDTO>;
  DeleteBrand(id: string): Promise<void>;

  ListSuppliers(req: ListSuppliersRequest): Promise<PageResult<SupplierDTO>>;
  GetSupplier(id: string): Promise<SupplierDTO>;
  CreateSupplier(req: CreateSupplierRequest): Promise<SupplierDTO>;
  UpdateSupplier(req: UpdateSupplierRequest): Promise<SupplierDTO>;
  RemoveSupplier(id: string): Promise<void>;

  ListSales(req: ListSalesRequest): Promise<PageResult<SaleDTO>>;
  GetSale(id: string): Promise<SaleDTO>;
  CreateSale(req: CreateSaleRequest): Promise<SaleDTO>;
  CancelSale(req: CancelSaleRequest): Promise<SaleDTO>;
  RegisterSalePayment(req: RegisterSalePaymentRequest): Promise<SaleDTO>;
  ListCustomerPayments(req: ListCustomerPaymentsRequest): Promise<PageResult<CustomerPaymentDTO>>;
  ListCustomerAdvances(customerId: string): Promise<CustomerAdvanceDTO[]>;

  ListBankAccounts(): Promise<BankAccountDTO[]>;
  GetBankAccount(id: string): Promise<BankAccountDTO>;
  CreateBankAccount(req: CreateBankAccountRequest): Promise<BankAccountDTO>;
  UpdateBankAccount(req: UpdateBankAccountRequest): Promise<BankAccountDTO>;
  DeleteBankAccount(id: string): Promise<void>;
  ListBankTransactions(req: ListBankTransactionsRequest): Promise<PageResult<BankTransactionDTO>>;
  CreateBankTransaction(req: CreateBankTransactionRequest): Promise<BankTransactionDTO>;
  ReconcileBankTransaction(id: string): Promise<BankTransactionDTO>;
  UpsertExchangeRate(req: UpsertExchangeRateRequest): Promise<void>;
  LatestExchangeRate(from: string, to: string): Promise<string>;

  ListInventoryBatches(req: ListInventoryBatchesRequest): Promise<PageResult<InventoryBatchDTO>>;
  ListInventoryMovements(req: ListInventoryMovementsRequest): Promise<PageResult<InventoryMovementDTO>>;
  GetClearanceCandidates(): Promise<InventoryBatchDTO[]>;
  ReceiveStock(req: ReceiveStockRequest): Promise<InventoryBatchDTO>;
  IssueStock(req: IssueStockRequest): Promise<void>;
  AdjustStock(req: AdjustStockRequest): Promise<void>;
  VoidStock(req: VoidStockRequest): Promise<void>;
  ListWarehouses(): Promise<WarehouseDTO[]>;

  ListChartOfAccounts(): Promise<AccountDTO[]>;
  CreateChartOfAccount(req: CreateChartOfAccountRequest): Promise<AccountDTO>;
  UpdateChartOfAccount(req: UpdateChartOfAccountRequest): Promise<AccountDTO>;
  DeleteChartOfAccount(id: string): Promise<void>;
  ListFiscalPeriods(): Promise<FiscalPeriodDTO[]>;
  ListJournalEntries(req: ListJournalEntriesRequest): Promise<PageResult<JournalEntryDTO>>;
  CreateJournalEntry(req: CreateJournalEntryRequest): Promise<JournalEntryDTO>;
  PostJournalEntry(id: string): Promise<JournalEntryDTO>;

  ListPurchaseOrders(req: ListPurchaseOrdersRequest): Promise<PageResult<PurchaseOrderDTO>>;
  GetPurchaseOrder(id: string): Promise<PurchaseOrderDTO>;
  CreatePurchaseOrder(req: CreatePurchaseOrderRequest): Promise<PurchaseOrderDTO>;
  CancelPurchaseOrder(req: CancelPurchaseOrderRequest): Promise<void>;
  RegisterPurchasePayment(req: RegisterPurchasePaymentRequest): Promise<PurchaseOrderDTO>;
}
