import type {
  AdjustStockRequest,
  AppBindings,
  CancelPurchaseRequest,
  CancelSaleRequest,
  CardPaymentRequest,
  CreatePurchaseRequest,
  CreateSaleRequest,
  ListSalesRequest,
  PaginationRequest,
  PurchaseFilterRequest,
  ReceiveStockRequest,
  SaveCreditCardRequest,
  SaveCustomerRequest,
  SaveProductRequest,
  SalePaymentRequest,
  SetLocalPasswordRequest,
  SetupWorkspaceRequest,
  SyncConfig,
  RecoverWithTokenRequest,
  UpdateLocalProfileRequest,
  VoidStockRequest,
} from './wails-types';

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
  async setupWorkspace(req: SetupWorkspaceRequest) {
    const b = await resolveBindings();
    return b.SetupWorkspace(req);
  },
  async updateLocalProfile(req: UpdateLocalProfileRequest) {
    const b = await resolveBindings();
    return b.UpdateLocalProfile(req);
  },
  async unlockLocalProfile(password: string) {
    const b = await resolveBindings();
    return b.UnlockLocalProfile(password);
  },
  async recoverWithToken(req: RecoverWithTokenRequest) {
    const b = await resolveBindings();
    return b.RecoverWithToken(req);
  },
  async setLocalPassword(req: SetLocalPasswordRequest) {
    const b = await resolveBindings();
    return b.SetLocalPassword(req);
  },
  async removeLocalPassword(current: string) {
    const b = await resolveBindings();
    return b.RemoveLocalPassword(current);
  },
  async getRecoveryToken() {
    const b = await resolveBindings();
    return b.GetRecoveryToken();
  },
  async lockLocalProfile() {
    const b = await resolveBindings();
    return b.LockLocalProfile();
  },

  async getPreferences() {
    const b = await resolveBindings();
    return b.GetPreferences();
  },
  async updatePreference(key: string, value: number | string) {
    const b = await resolveBindings();
    return b.UpdatePreference(key, value);
  },
  async getClearanceDays() {
    const b = await resolveBindings();
    return b.GetClearanceDays();
  },

  async getSyncConfig() {
    const b = await resolveBindings();
    return b.GetSyncConfig();
  },
  async saveSyncConfig(cfg: SyncConfig) {
    const b = await resolveBindings();
    return b.SaveSyncConfig(cfg);
  },
  async testSyncConnection(cfg: SyncConfig) {
    const b = await resolveBindings();
    return b.TestSyncConnection(cfg);
  },
  async syncNow() {
    const b = await resolveBindings();
    return b.SyncNow();
  },

  async chooseBackupFolder() {
    const b = await resolveBindings();
    return b.ChooseBackupFolder();
  },
  async setBackupFolder(folder: string) {
    const b = await resolveBindings();
    return b.SetBackupFolder(folder);
  },
  async setBackupFrequency(frequency: string) {
    const b = await resolveBindings();
    return b.SetBackupFrequency(frequency);
  },
  async createBackup() {
    const b = await resolveBindings();
    return b.CreateBackup();
  },

  async listCustomers(req: PaginationRequest, search: string) {
    const b = await resolveBindings();
    return b.ListCustomers(req, search);
  },
  async customerOptions() {
    const b = await resolveBindings();
    return b.CustomerOptions();
  },
  async createCustomer(req: SaveCustomerRequest) {
    const b = await resolveBindings();
    return b.CreateCustomer(req);
  },
  async updateCustomer(req: SaveCustomerRequest) {
    const b = await resolveBindings();
    return b.UpdateCustomer(req);
  },
  async removeCustomer(id: string) {
    const b = await resolveBindings();
    return b.RemoveCustomer(id);
  },
  async getCustomerByDocument(docType: string, docNumber: string) {
    const b = await resolveBindings();
    return b.GetCustomerByDocument(docType, docNumber);
  },

  async listProducts(req: PaginationRequest, search: string) {
    const b = await resolveBindings();
    return b.ListProducts(req, search);
  },
  async productOptions() {
    const b = await resolveBindings();
    return b.ProductOptions();
  },
  async getProduct(id: string) {
    const b = await resolveBindings();
    return b.GetProduct(id);
  },
  async createProduct(req: SaveProductRequest) {
    const b = await resolveBindings();
    return b.CreateProduct(req);
  },
  async updateProduct(req: SaveProductRequest) {
    const b = await resolveBindings();
    return b.UpdateProduct(req);
  },
  async removeProduct(id: string) {
    const b = await resolveBindings();
    return b.RemoveProduct(id);
  },
  async setProductActive(id: string, active: boolean) {
    const b = await resolveBindings();
    return b.SetProductActive(id, active);
  },
  async getProductStock(id: string) {
    const b = await resolveBindings();
    return b.GetProductStock(id);
  },

  async listInventoryBatches(req: PaginationRequest, onlyClearance: boolean, search: string) {
    const b = await resolveBindings();
    return b.ListInventoryBatches(req, onlyClearance, search);
  },
  async listInventoryMovements(req: PaginationRequest, productId: string) {
    const b = await resolveBindings();
    return b.ListInventoryMovements(req, productId);
  },
  async receiveStock(req: ReceiveStockRequest) {
    const b = await resolveBindings();
    return b.ReceiveStock(req);
  },
  async adjustStock(req: AdjustStockRequest) {
    const b = await resolveBindings();
    return b.AdjustStock(req);
  },
  async voidStock(req: VoidStockRequest) {
    const b = await resolveBindings();
    return b.VoidStock(req);
  },
  async listClearanceProducts() {
    const b = await resolveBindings();
    return b.ListClearanceProducts();
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
  async registerSalePayment(req: SalePaymentRequest) {
    const b = await resolveBindings();
    return b.RegisterSalePayment(req);
  },
  async listSalePayments(req: PaginationRequest, customerId: string, saleId: string) {
    const b = await resolveBindings();
    return b.ListSalePayments(req, customerId, saleId);
  },
  async listSaleCollections(from: string, to: string) {
    const b = await resolveBindings();
    return b.ListSaleCollections(from, to);
  },

  async listPurchaseOrders(req: PurchaseFilterRequest) {
    const b = await resolveBindings();
    return b.ListPurchaseOrders(req);
  },
  async getPurchaseOrder(id: string) {
    const b = await resolveBindings();
    return b.GetPurchaseOrder(id);
  },
  async createPurchase(req: CreatePurchaseRequest) {
    const b = await resolveBindings();
    return b.CreatePurchase(req);
  },
  async markPurchaseReceived(id: string, receivedDate: string) {
    const b = await resolveBindings();
    return b.MarkPurchaseReceived(id, receivedDate);
  },
  async cancelPurchase(req: CancelPurchaseRequest) {
    const b = await resolveBindings();
    return b.CancelPurchase(req);
  },
  async markPurchaseFaulty(req: CancelPurchaseRequest) {
    const b = await resolveBindings();
    return b.MarkPurchaseFaulty(req);
  },
  async createImportLot(description: string, purchaseIds: string[]) {
    const b = await resolveBindings();
    return b.CreateImportLot(description, purchaseIds);
  },
  async addToImportLot(lotId: string, purchaseIds: string[]) {
    const b = await resolveBindings();
    return b.AddToImportLot(lotId, purchaseIds);
  },
  async removeFromImportLot(lotId: string, purchaseId: string) {
    const b = await resolveBindings();
    return b.RemoveFromImportLot(lotId, purchaseId);
  },
  async getImportLot(id: string) {
    const b = await resolveBindings();
    return b.GetImportLot(id);
  },
  async closeImportLot(id: string) {
    const b = await resolveBindings();
    return b.CloseImportLot(id);
  },
  async listImportLots(req: PaginationRequest, search: string) {
    const b = await resolveBindings();
    return b.ListImportLots(req, search);
  },
  async listLotMembers(lotId: string) {
    const b = await resolveBindings();
    return b.ListLotMembers(lotId);
  },

  async listCreditCards() {
    const b = await resolveBindings();
    return b.ListCreditCards();
  },
  async issueCreditCard(req: SaveCreditCardRequest) {
    const b = await resolveBindings();
    return b.IssueCreditCard(req);
  },
  async updateCreditCard(req: SaveCreditCardRequest) {
    const b = await resolveBindings();
    return b.UpdateCreditCard(req);
  },
  async deleteCreditCard(id: string) {
    const b = await resolveBindings();
    return b.DeleteCreditCard(id);
  },
  async payCreditCard(req: CardPaymentRequest) {
    const b = await resolveBindings();
    return b.PayCreditCard(req);
  },
  async getCardProjections() {
    const b = await resolveBindings();
    return b.GetCardProjections();
  },
  async latestExchangeRate() {
    const b = await resolveBindings();
    return b.LatestExchangeRate();
  },
};
