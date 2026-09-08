import type {
  CardProjectionDTO,
  CreditCardDTO,
  UpdateCreditCardRequest,
} from '../wails-types';
import { wailsClient } from '../bindings';

interface CreditCardInput {
  issuer: string;
  lastFour: string;
  cardHolder: string;
  expirationMonth: number;
  expirationYear: number;
  creditLimit: number;
  cutOffDay: number;
  paymentDueDay: number;
  currencyCode: string;
}

interface CreditCardUpdateInput {
  issuer: string;
  lastFour: string;
  cardHolder: string;
  creditLimit: number;
  cutOffDay: number;
  paymentDueDay: number;
  isActive: boolean;
}

interface CreditCard {
  id: string;
  issuer: string;
  lastFour: string;
  cardHolder: string;
  expirationMonth: number;
  expirationYear: number;
  creditLimit: number;
  currentBalance: number;
  cutOffDay: number;
  paymentDueDay: number;
  currencyCode: string;
  isActive: boolean;
}

interface CardProjection {
  cardId: string;
  issuer: string;
  lastFour: string;
  cardHolder: string;
  projectedUSD: number;
  cycleStart: string;
  nextCutOffDate: string;
  nextPaymentDate: string;
}

function toCreditCard(dto: CreditCardDTO): CreditCard {
  return {
    id: dto.id,
    issuer: dto.issuer,
    lastFour: dto.lastFour,
    cardHolder: dto.cardHolder,
    expirationMonth: dto.expirationMonth,
    expirationYear: dto.expirationYear,
    creditLimit: Number(dto.creditLimit),
    currentBalance: Number(dto.currentBalance),
    cutOffDay: dto.cutOffDay,
    paymentDueDay: dto.paymentDueDay,
    currencyCode: dto.currencyCode,
    isActive: dto.isActive,
  };
}

function toCardProjection(dto: CardProjectionDTO): CardProjection {
  return {
    cardId: dto.cardId,
    issuer: dto.issuer,
    lastFour: dto.lastFour,
    cardHolder: dto.cardHolder,
    projectedUSD: dto.projectedUSD,
    cycleStart: dto.cycleStart,
    nextCutOffDate: dto.nextCutOffDate,
    nextPaymentDate: dto.nextPaymentDate,
  };
}

export const treasuryService = {
  async listCreditCards(): Promise<CreditCard[]> {
    return (await wailsClient.listCreditCards()).map(toCreditCard);
  },
  async createCreditCard(input: CreditCardInput): Promise<CreditCard> {
    return toCreditCard(
      await wailsClient.issueCreditCard({
        issuer: input.issuer,
        lastFour: input.lastFour,
        cardHolder: input.cardHolder,
        expirationMonth: input.expirationMonth,
        expirationYear: input.expirationYear,
        creditLimit: input.creditLimit.toFixed(2),
        cutOffDay: input.cutOffDay,
        paymentDueDay: input.paymentDueDay,
        currencyCode: input.currencyCode,
      }),
    );
  },
  async updateCreditCard(id: string, input: CreditCardUpdateInput): Promise<CreditCard> {
    const req: UpdateCreditCardRequest = {
      id,
      issuer: input.issuer,
      lastFour: input.lastFour,
      cardHolder: input.cardHolder,
      creditLimit: input.creditLimit.toFixed(2),
      cutOffDay: input.cutOffDay,
      paymentDueDay: input.paymentDueDay,
      isActive: input.isActive,
    };
    return toCreditCard(await wailsClient.updateCreditCard(req));
  },
  async deleteCreditCard(id: string): Promise<void> {
    await wailsClient.deleteCreditCard(id);
  },
  async getCardProjections(): Promise<CardProjection[]> {
    return (await wailsClient.getCardProjections()).map(toCardProjection);
  },
  async payCard(cardId: string, amount: number): Promise<void> {
    await wailsClient.payCreditCard(cardId, amount);
  },
  async getExchangeRate(from: string, to: string): Promise<number> {
    if (from === to) return 1;
    return Number(await wailsClient.latestExchangeRate(from, to));
  },
};
