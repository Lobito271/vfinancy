import type {
  CardPaymentRequest,
  CardProjectionDTO,
  CreditCardDTO,
  SaveCreditCardRequest,
} from '../wails-types';
import { wailsClient } from '../bindings';

export type CreditCardInput = Omit<SaveCreditCardRequest, 'id'>;

export const treasuryService = {
  async listCreditCards(): Promise<CreditCardDTO[]> {
    return wailsClient.listCreditCards();
  },
  async createCreditCard(input: CreditCardInput): Promise<CreditCardDTO> {
    return wailsClient.issueCreditCard(input);
  },
  async updateCreditCard(id: string, input: CreditCardInput): Promise<CreditCardDTO> {
    return wailsClient.updateCreditCard({ ...input, id });
  },
  async deleteCreditCard(id: string): Promise<void> {
    await wailsClient.deleteCreditCard(id);
  },
  async payCard(cardId: string, amount: number): Promise<CreditCardDTO> {
    const req: CardPaymentRequest = { cardId, amount };
    return wailsClient.payCreditCard(req);
  },
  async getCardProjections(): Promise<CardProjectionDTO[]> {
    return wailsClient.getCardProjections();
  },
  async getExchangeRate(_from?: string, _to?: string): Promise<number> {
    const info = await wailsClient.latestExchangeRate();
    return info.rate;
  },
  async exchangeRateInfo() {
    return wailsClient.latestExchangeRate();
  },
};
