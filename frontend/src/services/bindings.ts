import type {
  AdjustStockRequest,
  AppBindings,
  BusinessInfoDTO,
  CancelPurchaseOrderRequest,
  CancelSaleRequest,
  CompanyRequest,
  CreateCustomerRequest,
  CreateProductRequest,
  CreatePurchaseOrderRequest,
  CreateSaleRequest,
  CreateSupplierRequest,
  CreateCategoryRequest,
  CreateBrandRequest,
  IssueCreditCardRequest,
  IssueStockRequest,
  SetupWorkspaceRequest,
  ListCustomersRequest,
  ListInventoryBatchesRequest,
  ListInventoryMovementsRequest,
  ListNotificationsRequest,
  ListProductsRequest,
  ListPurchaseOrdersRequest,
  ListSalesRequest,
  ListSuppliersRequest,
  MarkPurchaseFaultyRequest,
  MarkPurchaseReceivedRequest,
  ReceiveStockRequest,
  RegisterCustomerOrderPaymentRequest,
  RegisterPurchasePaymentRequest,
  RegisterSalePaymentRequest,
  UpdateCategoryRequest,
  UpdateBrandRequest,
  UpdateCreditCardRequest,
  UpdateCustomerRequest,
  UpdateLocalProfileRequest,
  UpdateProductRequest,
  UpdateSupplierRequest,
  VoidStockRequest,
} from './wails-types';

// ponytail: not yet in wails-types.ts; wails regen will remove.
interface SyncConfigDTO {
  serverUrl: string;
  apiKey: string;
  enabled: boolean;
  pollIntervalSec: number;
}
let resolved: AppBindings | null = null;

function isWailsRuntime(): boolean {
  return typeof window !== 'undefined' && Boolean((window as { go?: unknown }).go);
}

async function resolveBindings(): Promise<AppBindings> {
  if (resolved) return resolved;
  if (!isWailsRuntime()) {
    throw new Error(
      'VFinancy bindings are unavailable outside the Wails runtime. ' +
        'Start the app with `wails dev` (or `wails build`) so the Go backend is bound to window.go.',
    );
  }
  const mod = await import('../../wailsjs/go/bindings/App');
  resolved = mod as unknown as AppBindings;
  return resolved;
}

export const wailsClient = {
  async getLocalAuthState() {
    const b = await resolveBindings();
    return b.GetLocalAuthState();
  },
  async getLocalProfile() {
    const b = await resolveBindings();
    return b.GetLocalProfile();
  },
  async updateLocalProfile(req: UpdateLocalProfileRequest) {
    const b = await resolveBindings();
    return b.UpdateLocalProfile(req);
  },
  async unlockLocalProfile(password: string) {
    const b = await resolveBindings();
    return b.UnlockLocalProfile(password);
  },
  async setLocalPassword(current: string, next: string) {
    const b = await resolveBindings();
    return b.SetLocalPassword(current, next);
  },
  async removeLocalPassword(current: string) {
    const b = await resolveBindings();
    return b.RemoveLocalPassword(current);
  },
  async lockLocalProfile() {
    const b = await resolveBindings();
    return b.LockLocalProfile();
  },
  async setupWorkspace(req: SetupWorkspaceRequest) {
    const b = await resolveBindings();
    return b.SetupWorkspace(req);
  },
  async listCompanies() {
    const b = await resolveBindings();
    return b.ListCompanies();
  },
  async getActiveCompany() {
    const b = await resolveBindings();
    return b.GetActiveCompany();
  },
  async setActiveCompany(id: string) {
    const b = await resolveBindings();
    return b.SetActiveCompany(id);
  },
  async createCompany(req: CompanyRequest) {
    const b = await resolveBindings();
    return b.CreateCompany(req);
  },
  async updateCompany(req: CompanyRequest) {
    const b = await resolveBindings();
    return b.UpdateCompany(req);
  },
  async deactivateCompany(id: string) {
    const b = await resolveBindings();
    return (b as any).DeactivateCompany(id);
  },
  async getBusinessInfo() {
    const b = await resolveBindings();
    return b.GetBusinessInfo();
  },
  async updateBusinessInfo(info: BusinessInfoDTO) {
    const b = await resolveBindings();
    return b.UpdateBusinessInfo(info);
  },
  async getPreferences() {
    const b = await resolveBindings();
    return b.GetPreferences();
  },
  async updatePreference(key: string, value: string) {
    const b = await resolveBindings();
    return b.UpdatePreference(key, value);
  },
  async getCurrencies() {
    const b = await resolveBindings();
    return b.GetCurrencies();
  },
  async getTaxes() {
    const b = await resolveBindings();
    return b.GetTaxes();
  },
  async getAllSettings() {
    const b = await resolveBindings();
    return b.GetAllSettings();
  },

  async listCustomers(req: ListCustomersRequest) {
    const b = await resolveBindings();
    return b.ListCustomers(req);
  },
  async getCustomer(id: string) {
    const b = await resolveBindings();
    return b.GetCustomer(id);
  },
  async createCustomer(req: CreateCustomerRequest) {
    const b = await resolveBindings();
    return b.CreateCustomer(req);
  },
  async updateCustomer(req: UpdateCustomerRequest) {
    const b = await resolveBindings();
    return b.UpdateCustomer(req);
  },
  async removeCustomer(id: string) {
    const b = await resolveBindings();
    return b.RemoveCustomer(id);
  },

  async listProducts(req: ListProductsRequest) {
    const b = await resolveBindings();
    return b.ListProducts(req);
  },
  async getProduct(id: string) {
    const b = await resolveBindings();
    return b.GetProduct(id);
  },
  async createProduct(req: CreateProductRequest) {
    const b = await resolveBindings();
    return b.CreateProduct(req);
  },
  async updateProduct(req: UpdateProductRequest) {
    const b = await resolveBindings();
    return b.UpdateProduct(req);
  },
  async removeProduct(id: string) {
    const b = await resolveBindings();
    return b.RemoveProduct(id);
  },
  async listUnits() {
    const b = await resolveBindings();
    return b.ListUnits();
  },
  async listCategories() {
    const b = await resolveBindings();
    return b.ListCategories();
  },
  async createCategory(req: CreateCategoryRequest) {
    const b = await resolveBindings();
    return b.CreateCategory(req);
  },
  async updateCategory(req: UpdateCategoryRequest) {
    const b = await resolveBindings();
    return b.UpdateCategory(req);
  },
  async deleteCategory(id: string) {
    const b = await resolveBindings();
    return b.DeleteCategory(id);
  },
  async listBrands() {
    const b = await resolveBindings();
    return b.ListBrands();
  },
  async createBrand(req: CreateBrandRequest) {
    const b = await resolveBindings();
    return b.CreateBrand(req);
  },
  async updateBrand(req: UpdateBrandRequest) {
    const b = await resolveBindings();
    return b.UpdateBrand(req);
  },
  async deleteBrand(id: string) {
    const b = await resolveBindings();
    return b.DeleteBrand(id);
  },

  async listSuppliers(req: ListSuppliersRequest) {
    const b = await resolveBindings();
    return b.ListSuppliers(req);
  },
  async getSupplier(id: string) {
    const b = await resolveBindings();
    return b.GetSupplier(id);
  },
  async createSupplier(req: CreateSupplierRequest) {
    const b = await resolveBindings();
    return b.CreateSupplier(req);
  },
  async updateSupplier(req: UpdateSupplierRequest) {
    const b = await resolveBindings();
    return b.UpdateSupplier(req);
  },
  async removeSupplier(id: string) {
    const b = await resolveBindings();
    return b.RemoveSupplier(id);
  },

  async listSales(req: ListSalesRequest) {
    const b = await resolveBindings();
    return b.ListSales(req);
  },
  async getSale(id: string) {
    const b = await resolveBindings();
    return b.GetSale(id);
  },
  async createSale(req: CreateSaleRequest) {
    const b = await resolveBindings();
    return b.CreateSale(req);
  },
  async cancelSale(req: CancelSaleRequest) {
    const b = await resolveBindings();
    return b.CancelSale(req);
  },
  async registerSalePayment(req: RegisterSalePaymentRequest) {
    const b = await resolveBindings();
    return b.RegisterSalePayment(req);
  },

  async listCreditCards() {
    const b = await resolveBindings();
    return b.ListCreditCards();
  },
  async issueCreditCard(req: IssueCreditCardRequest) {
    const b = await resolveBindings();
    return b.IssueCreditCard(req);
  },
  async updateCreditCard(req: UpdateCreditCardRequest) {
    const b = await resolveBindings();
    return b.UpdateCreditCard(req);
  },
  async deleteCreditCard(id: string) {
    const b = await resolveBindings();
    return b.DeleteCreditCard(id);
  },
  async getCardProjections() {
    const b = await resolveBindings();
    return b.GetCardProjections();
  },
  async payCreditCard(cardId: string, amount: number) {
    const b = await resolveBindings();
    return b.PayCreditCard({ cardId, amount: amount.toFixed(2) });
  },
  async latestExchangeRate(from: string, to: string) {
    const b = await resolveBindings();
    return b.LatestExchangeRate(from, to);
  },

  async listInventoryBatches(req: ListInventoryBatchesRequest) {
    const b = await resolveBindings();
    return b.ListInventoryBatches(req);
  },
  async listInventoryMovements(req: ListInventoryMovementsRequest) {
    const b = await resolveBindings();
    return b.ListInventoryMovements(req);
  },
  async receiveStock(req: ReceiveStockRequest) {
    const b = await resolveBindings();
    return b.ReceiveStock(req);
  },
  async issueStock(req: IssueStockRequest) {
    const b = await resolveBindings();
    return b.IssueStock(req);
  },
  async adjustStock(req: AdjustStockRequest) {
    const b = await resolveBindings();
    return b.AdjustStock(req);
  },
  async voidStock(req: VoidStockRequest) {
    const b = await resolveBindings();
    return b.VoidStock(req);
  },
  async listWarehouses() {
    const b = await resolveBindings();
    return b.ListWarehouses();
  },

  async listPurchaseOrders(req: ListPurchaseOrdersRequest) {
    const b = await resolveBindings();
    return b.ListPurchaseOrders(req);
  },
  async getPurchaseOrder(id: string) {
    const b = await resolveBindings();
    return b.GetPurchaseOrder(id);
  },
  async createPurchaseOrder(req: CreatePurchaseOrderRequest) {
    const b = await resolveBindings();
    return b.CreatePurchaseOrder(req);
  },
  async cancelPurchaseOrder(req: CancelPurchaseOrderRequest) {
    const b = await resolveBindings();
    return b.CancelPurchaseOrder(req);
  },
  async registerPurchasePayment(req: RegisterPurchasePaymentRequest) {
    const b = await resolveBindings();
    return b.RegisterPurchasePayment(req);
  },
  async markPurchaseFaulty(req: MarkPurchaseFaultyRequest) {
    const b = await resolveBindings();
    return b.MarkPurchaseFaulty(req);
  },
  async markPurchaseReceived(req: MarkPurchaseReceivedRequest) {
    const b = await resolveBindings();
    return b.MarkPurchaseReceived(req);
  },
  async registerCustomerOrderPayment(req: RegisterCustomerOrderPaymentRequest) {
    const b = await resolveBindings();
    return b.RegisterCustomerOrderPayment(req);
  },

  async listNotifications(req: ListNotificationsRequest) {
    const b = await resolveBindings();
    return b.ListNotifications(req);
  },
  async unreadNotificationCount() {
    const b = await resolveBindings();
    return b.UnreadNotificationCount();
  },
  async markNotificationsRead(ids: string[]) {
    const b = await resolveBindings();
    return b.MarkNotificationsRead(ids);
  },
  async markAllNotificationsRead() {
    const b = await resolveBindings();
    return b.MarkAllNotificationsRead();
  },
  async deleteNotification(id: string) {
    const b = await resolveBindings();
    return b.DeleteNotification(id);
  },
  async generateClearanceNotifications() {
    const b = await resolveBindings();
    return b.GenerateClearanceNotifications();
  },

  async createBackup() {
    const b = await resolveBindings();
    return (b as any).CreateBackup();
  },
  async getSyncConfig() {
    const b = await resolveBindings();
    return (b as any).GetSyncConfig();
  },
  async saveSyncConfig(cfg: SyncConfigDTO) {
    const b = await resolveBindings();
    return (b as any).SaveSyncConfig(cfg);
  },
  async testSyncConnection(serverUrl: string, apiKey: string) {
    const b = await resolveBindings();
    return (b as any).TestSyncConnection(serverUrl, apiKey);
  },
};
