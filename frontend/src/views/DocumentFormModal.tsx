import { useEffect, useId, useMemo, useRef, useState } from 'react'
import type { Dispatch, FormEvent, SetStateAction } from 'react'
import { Plus, Trash2 } from 'lucide-react'
import { service } from '../../wailsjs/go/models'
import { api } from '../data/api'
import type { Currency } from '../data/api'
import { useStore } from '../data/store'
import { parseError } from '../lib/errors'
import { todayISO } from '../lib/format'
import { centsText, textToCents, tcFromText, tcText } from '../lib/money'
import { Field, Modal, ModalActions, Money, MoneyInput, Seg, TcField, TextInput } from '../components/ui'
import { Autocomplete } from '../components/Autocomplete'
import { AddressModal, CardModal, ClientModal, ProductModal, SupplierModal } from '../components/entityModals'

type Catalog = { clients: Array<Record<string, any>>; suppliers: Array<Record<string, any>>; products: Array<Record<string, any>>; cards: Array<Record<string, any>> }
type Line = { uid: string; productId: string; lotId: string; qty: string; amount: string; description: string }
type ExtraLine = { uid: string; concept: string; amount: string; currency: Currency }
type Kind = 'purchase' | 'sale' | 'shipment'
type Defaults = { code: string; tc: string }
type QuickKind = 'client' | 'supplier' | 'product' | 'card'

const blankLine = (): Line => ({ uid: crypto.randomUUID(), productId: '', lotId: '', qty: '1', amount: '', description: '' })

function documentTitle(existing: unknown, kind: Kind, code: string): string {
  if (existing) return `Editar · ${code}`
  return kind === 'purchase' ? 'Nueva compra' : kind === 'shipment' ? 'Nuevo envío' : 'Nueva venta'
}

function submitLabel(existing: unknown, kind: Kind): string {
  if (existing) return 'Guardar cambios'
  return kind === 'purchase' ? 'Registrar compra' : kind === 'shipment' ? 'Registrar envío' : 'Registrar venta'
}

export function DocumentFormModal({ kind, catalog, onClose, existing, defaults }: { kind: Kind; catalog: Catalog; onClose: () => void; existing?: service.Purchase | service.Sale; defaults?: Defaults }) {
  const displayCurrency = useStore((state) => state.currency)
  const refresh = useStore((state) => state.refresh)
  const toast = useStore((state) => state.toast)
  const settings = useStore((state) => state.settings)
  const formId = useId()
  const record = existing as any
  const editingPurchase = kind === 'purchase'
  // Estado inicial del formulario a partir del documento existente o de los valores por defecto.
  const [initial] = useState(() => ({
    code: record?.code ?? defaults?.code ?? '',
    date: record?.date ?? todayISO(),
    currency: (record?.currency ?? (kind === 'purchase' ? 'USD' : 'PEN')) as Currency,
    tc: record?.tc ? tcText(record.tc) : defaults?.tc ?? '3.7500',
    supplierId: String(record?.supplierId ?? ''),
    clientId: String(record?.clientId ?? ''),
    forClient: record?.forClientId != null,
    method: record?.paymentMethod ?? settings.last_payment_method ?? 'cash',
    cardId: String(record?.cardId ?? ''),
    items: record?.items?.length ? record.items.map((item: any) => ({ uid: crypto.randomUUID(), productId: String(item.productId), lotId: String(item.lotId ?? ''), qty: String(item.qty), amount: centsText(item.unitCostCents ?? item.unitPriceCents), description: item.description ?? '' })) : [blankLine()],
    extras: record?.extras?.map((extra: any) => ({ uid: crypto.randomUUID(), concept: extra.concept, amount: centsText(extra.amountCents), currency: extra.currency as Currency })) ?? [],
    region: record?.shipRegion ?? '',
    securityCode: record?.securityCode || String(Math.floor(Math.random() * 10000)).padStart(4, '0'),
  }))

  const [code, setCode] = useState(initial.code)
  const [date, setDate] = useState(initial.date)
  const [currency, setCurrency] = useState<Currency>(initial.currency)
  const [tc, setTc] = useState(initial.tc)
  const [supplierId, setSupplierId] = useState(initial.supplierId)
  const [clientId, setClientId] = useState(initial.clientId)
  const [forClient, setForClient] = useState(initial.forClient)
  const [method, setMethod] = useState(initial.method)
  const [cardId, setCardId] = useState(initial.cardId)
  const [items, setItems] = useState<Line[]>(initial.items)
  const [lotOptions, setLotOptions] = useState<Record<string, Array<Record<string, any>>>>({})
  const [extras, setExtras] = useState<ExtraLine[]>(initial.extras)
  const [paidNow, setPaidNow] = useState('')
  const [region, setRegion] = useState(initial.region)
  const addressRef = useRef(record?.shipAddress ?? '')
  const [addressId, setAddressId] = useState('')
  const [addingAddress, setAddingAddress] = useState(false)
  const [securityCode, setSecurityCode] = useState(initial.securityCode)
  const [preview, setPreview] = useState<service.DocumentTotals | null>(null)
  const [busy, setBusy] = useState(false)
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({})
  const [quickKind, setQuickKind] = useState<QuickKind | null>(null)
  const [quickName, setQuickName] = useState('')
  const quickProductIndexRef = useRef(0)
  const [addresses, setAddresses] = useState<Array<Record<string, any>>>([])
  const tcScaled = tcFromText(tc) ?? 37500

  useEffect(() => { if (clientId && kind === 'shipment') void api.Catalog.ListAddresses(Number(clientId)).then((rows) => setAddresses(rows as Array<Record<string, any>>)).catch(() => setAddresses([])); else setAddresses([]) }, [clientId, kind])

  const productIds = useMemo(() => [...new Set(items.map((line) => line.productId).filter(Boolean))], [items])
  useEffect(() => {
    for (const productId of productIds) {
      if (lotOptions[productId]) continue
      void api.Inventory.LotsForProduct(Number(productId)).then((rows) => {
        const currentSaleItems = existing && !editingPurchase ? (existing as service.Sale).items ?? [] : []
        setLotOptions((old) => ({ ...old, [productId]: (rows as Array<Record<string, any>>).map((lot) => ({ ...lot, available: Number(lot.available) + currentSaleItems.filter((item) => Number(item.lotId) === Number(lot.id)).reduce((sum, item) => sum + Number(item.qty), 0) })) }))
      }).catch(() => {})
    }
  }, [productIds, kind, lotOptions, existing, editingPurchase])

  const updateLine = (index: number, patch: Partial<Line>) => setItems((current) => current.map((line, i) => i === index ? { ...line, ...patch } : line))
  const setDefaultPrice = (index: number, productId: string) => {
    const product = catalog.products.find((candidate) => String(candidate.id) === productId)
    if (!product) return updateLine(index, { productId, lotId: '', amount: '' })
    const sourceCurrency = editingPurchase ? product.default_cost_currency : product.default_price_currency
    const value = editingPurchase
      ? (product.original_default_cost_cents ?? product.default_cost_cents)
      : (product.original_default_price_cents ?? product.default_price_cents)
    if (sourceCurrency !== currency) return updateLine(index, { productId, lotId: '', amount: '' })
    return updateLine(index, { productId, lotId: '', amount: value == null ? '' : `${Math.floor(Number(value) / 100)}.${String(Number(value) % 100).padStart(2, '0')}` })
  }
  const applyPaymentMethod = (value: string) => {
    setMethod(value)
    void api.Settings.SetSetting('last_payment_method', value).then(() => useStore.setState((current) => ({ settings: { ...current.settings, last_payment_method: value } })))
  }
  const openQuick = (next: QuickKind, name = '') => { setQuickName(name); setQuickKind(next) }
  const openProductQuick = (name: string, index: number) => { setQuickName(name); quickProductIndexRef.current = index; setQuickKind('product') }
  const requestAddAddress = () => {
    if (!clientId) { toast('Elige primero al cliente', 'warn'); return }
    setAddingAddress(true)
  }

  useEffect(() => {
    const timer = window.setTimeout(async () => {
      try {
        const lineItems = items.filter((line) => line.productId && Number(line.qty) > 0).map((line) => editingPurchase
          ? new service.PurchaseItem({ productId: Number(line.productId), qty: Number(line.qty), description: line.description, unitCostCents: textToCents(line.amount) ?? 0 })
          : new service.SaleItem({ productId: Number(line.productId), lotId: Number(line.lotId) || 0, qty: Number(line.qty), description: line.description, unitPriceCents: textToCents(line.amount) ?? 0 }))
        const extraItems = extras.filter((extra) => extra.concept.trim()).map((extra) => new service.Extra({ concept: extra.concept, amountCents: textToCents(extra.amount) ?? 0, currency: extra.currency }))
        const body = editingPurchase
          ? new service.Purchase({ currency, tc: tcScaled, items: lineItems as service.PurchaseItem[], extras: extraItems })
          : new service.Sale({ kind, currency, tc: tcScaled, items: lineItems as service.SaleItem[], extras: extraItems })
        const total = editingPurchase ? await api.Purchases.PreviewTotal(body as service.Purchase, displayCurrency) : await api.Sales.PreviewTotal(body as service.Sale, displayCurrency)
        setPreview(total)
      } catch { setPreview(null) }
    }, 250)
    return () => window.clearTimeout(timer)
  }, [items, extras, currency, tcScaled, displayCurrency, kind, editingPurchase])

  const submit = async (event: FormEvent) => {
    event.preventDefault(); setBusy(true); setFieldErrors({})
    try {
      const rate = tcFromText(tc)
      if (!rate) throw new Error('TC debe estar entre 0.5000 y 20.0000')
      const extraItems = extras.filter((extra) => extra.concept.trim() && textToCents(extra.amount) !== null).map((extra) => new service.Extra({ concept: extra.concept.trim(), amountCents: textToCents(extra.amount) ?? 0, currency: extra.currency }))
      if (editingPurchase) {
        const purchaseItems = items.filter((line) => line.productId).map((line) => new service.PurchaseItem({ productId: Number(line.productId), description: line.description, qty: Number(line.qty), unitCostCents: textToCents(line.amount) ?? 0 }))
        await api.Purchases.Save(new service.Purchase({ id: Number(existing?.id ?? 0), code, supplierId: Number(supplierId), date, paymentMethod: method, cardId: method === 'card' ? Number(cardId) : undefined, currency, tc: rate, forClientId: forClient && clientId ? Number(clientId) : undefined, items: purchaseItems, extras: extraItems }))
      } else {
        const saleItems: service.SaleItem[] = (await Promise.all(items.filter((item) => item.productId).map(async (line): Promise<service.SaleItem[]> => {
          const qty = Number(line.qty)
          if (existing && line.lotId) {
            return [new service.SaleItem({ productId: Number(line.productId), lotId: Number(line.lotId), qty, description: line.description, unitPriceCents: textToCents(line.amount) ?? 0 })]
          }
          if (line.lotId) {
            const choices = (lotOptions[line.productId] ?? []).filter((lot) => Number(lot.available) > 0 && (lot.reserved_client_id == null || Number(lot.reserved_client_id) === Number(clientId)))
            const selectedIndex = choices.findIndex((lot) => Number(lot.id) === Number(line.lotId))
            if (selectedIndex < 0) throw new Error(`Lote ${line.lotId} no tiene stock disponible para este cliente`)
            const group: service.SaleItem[] = []
            let remaining = qty
            for (const lot of choices.slice(selectedIndex)) {
              if (remaining <= 0) break
              const take = Math.min(remaining, Number(lot.available))
              group.push(new service.SaleItem({ productId: Number(line.productId), lotId: Number(lot.id), qty: take, description: line.description, unitPriceCents: textToCents(line.amount) ?? 0 }))
              remaining -= take
            }
            if (remaining > 0) throw new Error(`Stock insuficiente: faltan ${remaining} unidades`)
            return group
          }
          const suggested = await api.Inventory.SuggestLots(Number(line.productId), qty)
          return suggested.map((lot) => new service.SaleItem({ productId: Number(line.productId), lotId: lot.lotId, qty: lot.suggestedQty, description: line.description, unitPriceCents: textToCents(line.amount) ?? 0 }))
        }))).flat()
        await api.Sales.Save(new service.Sale({ id: Number(existing?.id ?? 0), code, kind, clientId: clientId ? Number(clientId) : undefined, date, currency, tc: rate, items: saleItems, extras: extraItems, paidNowCents: existing ? 0 : textToCents(paidNow) ?? 0, paymentMethod: method, shipRegion: region || record?.shipRegion, shipAddress: addressRef.current || record?.shipAddress, securityCode }))
      }
      refresh(); toast(`${code} registrado`); onClose()
    } catch (error) {
      const parsed = parseError(error); setFieldErrors(parsed.fields); toast(parsed.message ?? Object.values(parsed.fields)[0] ?? 'No se pudo guardar el documento', 'bad')
    } finally { setBusy(false) }
  }

  return <Modal title={documentTitle(existing, kind, code)} onClose={onClose} wide footer={<ModalActions onClose={onClose} busy={busy} formId={formId} label={submitLabel(existing, kind)} />}><form id={formId} className="stack" onSubmit={submit}>
    <PartyFields editingPurchase={editingPurchase} catalog={catalog} code={code} setCode={setCode} date={date} setDate={setDate} supplierId={supplierId} clientId={clientId} setSupplierId={setSupplierId} setClientId={setClientId} fieldErrors={fieldErrors} onQuickCreate={openQuick} />
    {editingPurchase && <PurchasePaymentFields method={method} cardId={cardId} setCardId={setCardId} catalog={catalog} fieldErrors={fieldErrors} onMethod={applyPaymentMethod} onQuickCreate={openQuick} />}
    {editingPurchase && <OrderFieldset forClient={forClient} setForClient={setForClient} clientId={clientId} setClientId={setClientId} catalog={catalog} fieldErrors={fieldErrors} onQuickCreate={openQuick} />}
    <CurrencyFields currency={currency} setCurrency={setCurrency} tc={tc} setTc={setTc} fieldErrors={fieldErrors} />
    <ItemsFieldset editingPurchase={editingPurchase} items={items} setItems={setItems} lotOptions={lotOptions} clientId={clientId} catalog={catalog} currency={currency} updateLine={updateLine} setDefaultPrice={setDefaultPrice} onCreateProduct={openProductQuick} fieldErrors={fieldErrors} />
    <ExtrasFieldset editingPurchase={editingPurchase} extras={extras} setExtras={setExtras} currency={currency} />
    {kind === 'shipment' && <ShipmentFields addresses={addresses} addressId={addressId} setAddressId={setAddressId} region={region} setRegion={setRegion} securityCode={securityCode} setSecurityCode={setSecurityCode} addressRef={addressRef} fieldErrors={fieldErrors} onAddAddress={requestAddAddress} />}
    {!editingPurchase && <PaidNowFields currency={currency} paidNow={paidNow} setPaidNow={setPaidNow} method={method} onMethod={applyPaymentMethod} />}
    <TotalsPreview preview={preview} displayCurrency={displayCurrency} />
  </form><QuickCreateModals
    quickKind={quickKind} quickName={quickName} clientId={clientId} addingAddress={addingAddress}
    onClose={() => { setQuickKind(null); setQuickName('') }}
    onSavedClient={(row) => { setClientId(String(row.id)); setAddressId(''); setAddresses([]); setQuickKind(null); setQuickName('') }}
    onSavedSupplier={(row) => { setSupplierId(String(row.id)); setQuickKind(null); setQuickName('') }}
    onSavedCard={(row) => { setCardId(String(row.id)); setQuickKind(null); setQuickName('') }}
    onSavedProduct={(row) => { setItems((current) => current.map((line, index) => index === quickProductIndexRef.current ? { ...line, productId: String(row.id) } : line)); setQuickKind(null); setQuickName('') }}
    onSavedAddress={(saved) => { setAddresses((old) => [...old, saved]); setAddressId(String(saved.id)); setRegion(saved.region); addressRef.current = saved.address; setAddingAddress(false) }}
    onCloseAddress={() => setAddingAddress(false)}
  /></Modal>
}

function PartyFields({ editingPurchase, catalog, code, setCode, date, setDate, supplierId, clientId, setSupplierId, setClientId, fieldErrors, onQuickCreate }: { editingPurchase: boolean; catalog: Catalog; code: string; setCode: (value: string) => void; date: string; setDate: (value: string) => void; supplierId: string; clientId: string; setSupplierId: (value: string) => void; setClientId: (value: string) => void; fieldErrors: Record<string, string>; onQuickCreate: (kind: QuickKind, name?: string) => void }) {
  return <div className="formgrid formgrid--3"><Field label="Código" required error={fieldErrors.code}><TextInput className="input--mono" value={code} onChange={(event) => setCode(event.target.value)} /></Field><Field label="Fecha" required><TextInput type="date" value={date} onChange={(event) => setDate(event.target.value)} /></Field><Field label={editingPurchase ? 'Proveedor' : 'Cliente'} required={editingPurchase} error={fieldErrors.supplierId ?? fieldErrors.clientId}><Autocomplete options={(editingPurchase ? catalog.suppliers : catalog.clients).map((row) => ({ value: String(row.id), label: row.name, sub: row.phone || undefined }))} value={editingPurchase ? supplierId : clientId} onChange={(value) => editingPurchase ? setSupplierId(value) : setClientId(value)} placeholder={editingPurchase ? 'Proveedor…' : 'Cliente general'} onCreate={(name) => onQuickCreate(editingPurchase ? 'supplier' : 'client', name)} /></Field></div>
}

function PurchasePaymentFields({ method, cardId, setCardId, catalog, fieldErrors, onMethod, onQuickCreate }: { method: string; cardId: string; setCardId: (value: string) => void; catalog: Catalog; fieldErrors: Record<string, string>; onMethod: (value: string) => void; onQuickCreate: (kind: QuickKind, name?: string) => void }) {
  return <div className="formgrid"><Field label="Método de pago"><Seg options={[{ value: 'card', label: 'Tarjeta' }, { value: 'cash', label: 'Efectivo' }, { value: 'wallet', label: 'Billetera digital' }]} value={method} onChange={onMethod} /></Field>{method === 'card' && <Field label="Tarjeta" required error={fieldErrors.cardId}><Autocomplete options={catalog.cards.map((row) => ({ value: String(row.id), label: `${row.name} · ••${row.last4}` }))} value={cardId} onChange={setCardId} placeholder="Tarjeta…" onCreate={(name) => onQuickCreate('card', name)} /></Field>}</div>
}

function OrderFieldset({ forClient, setForClient, clientId, setClientId, catalog, fieldErrors, onQuickCreate }: { forClient: boolean; setForClient: (value: boolean) => void; clientId: string; setClientId: (value: string) => void; catalog: Catalog; fieldErrors: Record<string, string>; onQuickCreate: (kind: QuickKind, name?: string) => void }) {
  return <fieldset className="formsection"><legend className="formsection__title">Encargo</legend><label className="checkline"><input type="checkbox" checked={forClient} onChange={(event) => setForClient(event.target.checked)} /> Para cliente: reservar los lotes al recibir</label>{forClient && <Field label="Cliente" required error={fieldErrors.forClientId}><Autocomplete options={catalog.clients.map((row) => ({ value: String(row.id), label: row.name, sub: row.phone || undefined }))} value={clientId} onChange={setClientId} placeholder="Cliente…" onCreate={() => onQuickCreate('client')} /></Field>}</fieldset>
}

function CurrencyFields({ currency, setCurrency, tc, setTc, fieldErrors }: { currency: Currency; setCurrency: (value: Currency) => void; tc: string; setTc: (value: string) => void; fieldErrors: Record<string, string> }) {
  return <div className="formgrid"><Field label="Moneda"><Seg options={[{ value: 'PEN', label: 'S/' }, { value: 'USD', label: '$' }]} value={currency} onChange={setCurrency} /></Field><Field label="TC (S/ por $)" required error={fieldErrors.tc}><TcField value={tc} onChange={setTc} /></Field></div>
}

function ItemsFieldset({ editingPurchase, items, setItems, lotOptions, clientId, catalog, currency, updateLine, setDefaultPrice, onCreateProduct, fieldErrors }: { editingPurchase: boolean; items: Line[]; setItems: Dispatch<SetStateAction<Line[]>>; lotOptions: Record<string, Array<Record<string, any>>>; clientId: string; catalog: Catalog; currency: Currency; updateLine: (index: number, patch: Partial<Line>) => void; setDefaultPrice: (index: number, productId: string) => void; onCreateProduct: (name: string, index: number) => void; fieldErrors: Record<string, string> }) {
  return <fieldset className="formsection"><legend className="formsection__title">Ítems</legend><div className="formsection__body stack">{items.map((line, index) => { const availableLots = (lotOptions[line.productId] ?? []).filter((lot) => Number(lot.available) > 0 && (lot.reserved_client_id == null || Number(lot.reserved_client_id) === Number(clientId))); const selectedLotId = line.lotId || String(availableLots[0]?.id ?? ''); const selectedLot = availableLots.find((lot) => String(lot.id) === selectedLotId); return <div className="lineditor__row lineditor__row--item" key={line.uid}><Field label="Producto" required><Autocomplete options={catalog.products.map((product) => ({ value: String(product.id), label: product.name, sub: product.sku }))} value={line.productId} onChange={(value) => setDefaultPrice(index, value)} placeholder="Producto…" onCreate={(name) => onCreateProduct(name, index)} /></Field>{!editingPurchase && <Field label="Lote · disponible" required><select className="select" value={selectedLotId} disabled={!availableLots.length} onChange={(event) => updateLine(index, { lotId: event.target.value })}><option value="">{availableLots.length ? 'Elegir lote' : 'Sin stock disponible'}</option>{availableLots.map((lot) => <option key={lot.id} value={lot.id}>{lot.code} · {lot.available} uds</option>)}</select>{selectedLot && Number(line.qty) > Number(selectedLot.available) && <small className="muted">Se dividirá en los siguientes lotes FIFO.</small>}</Field>}<Field label="Cantidad" required><TextInput className="input--quantity" type="number" min="1" step="1" value={line.qty} onChange={(event) => updateLine(index, { qty: event.target.value })} /></Field><Field label={editingPurchase ? 'Costo unitario' : 'Precio unitario'} required><MoneyInput currency={currency} value={line.amount} onChange={(amount) => updateLine(index, { amount })} /></Field><button className="iconbtn lineditor__remove" type="button" aria-label="Quitar ítem" disabled={items.length === 1} onClick={() => setItems((current) => current.filter((_, i) => i !== index))}><Trash2 size={14} /></button></div>})}<button className="btn btn--xs" type="button" onClick={() => setItems((current) => [...current, blankLine()])}><Plus size={13} /> Añadir ítem</button>{fieldErrors.items && <small className="field__error">{fieldErrors.items}</small>}</div></fieldset>
}

function ExtrasFieldset({ editingPurchase, extras, setExtras, currency }: { editingPurchase: boolean; extras: ExtraLine[]; setExtras: Dispatch<SetStateAction<ExtraLine[]>>; currency: Currency }) {
  return <fieldset className="formsection"><legend className="formsection__title">{editingPurchase ? 'Costos extras' : 'Extras'}</legend><div className="formsection__body stack">{extras.map((extra, index) => <div className="lineditor__row lineditor__row--extra" key={extra.uid}><Field label="Concepto" required><ConceptInput value={extra.concept} onChange={(concept) => setExtras((current) => current.map((item, i) => i === index ? { ...item, concept } : item))} /></Field><Field label="Monto" required><MoneyInput currency={extra.currency} value={extra.amount} onChange={(amount) => setExtras((current) => current.map((item, i) => i === index ? { ...item, amount } : item))} /></Field>{editingPurchase && <Field label="Moneda"><Seg options={[{ value: 'PEN', label: 'S/' }, { value: 'USD', label: '$' }]} value={extra.currency} onChange={(value) => setExtras((current) => current.map((item, i) => i === index ? { ...item, currency: value } : item))} /></Field>}<button className="iconbtn lineditor__remove" type="button" aria-label="Quitar extra" onClick={() => setExtras((current) => current.filter((_, i) => i !== index))}><Trash2 size={14} /></button></div>)}<button className="btn btn--xs" type="button" onClick={() => setExtras((current) => [...current, { uid: crypto.randomUUID(), concept: '', amount: '', currency }])}><Plus size={13} /> Añadir extra</button></div></fieldset>
}

function ShipmentFields({ addresses, addressId, setAddressId, region, setRegion, securityCode, setSecurityCode, addressRef, fieldErrors, onAddAddress }: { addresses: Array<Record<string, any>>; addressId: string; setAddressId: (value: string) => void; region: string; setRegion: (value: string) => void; securityCode: string; setSecurityCode: (value: string) => void; addressRef: { current: string }; fieldErrors: Record<string, string>; onAddAddress: () => void }) {
  return <div className="formgrid formgrid--3"><Field label="Dirección de envío" required error={fieldErrors.shipAddress}><Autocomplete options={addresses.map((item) => ({ value: String(item.id), label: `${item.region} · ${item.address}` }))} value={addressId} onChange={(value) => { setAddressId(value); const match = addresses.find((item) => String(item.id) === value); if (match) { setRegion(match.region); addressRef.current = match.address } }} placeholder="Busca una dirección" onCreate={onAddAddress} /></Field><Field label="Región" required error={fieldErrors.shipRegion}><TextInput value={region} onChange={(event) => setRegion(event.target.value)} /></Field><Field label="Código de seguridad" required><TextInput className="input--mono" inputMode="numeric" maxLength={4} value={securityCode} onChange={(event) => setSecurityCode(event.target.value.replace(/\D/g, '').slice(0, 4))} /></Field></div>
}

function PaidNowFields({ currency, paidNow, setPaidNow, method, onMethod }: { currency: Currency; paidNow: string; setPaidNow: (value: string) => void; method: string; onMethod: (value: string) => void }) {
  return <div className="formgrid"><Field label="Pagado ahora"><MoneyInput currency={currency} value={paidNow} onChange={setPaidNow} /></Field><Field label="Método de pago"><Seg options={[{ value: 'cash', label: 'Efectivo' }, { value: 'card', label: 'Tarjeta' }, { value: 'wallet', label: 'Billetera digital' }]} value={method} onChange={onMethod} /></Field></div>
}

function TotalsPreview({ preview, displayCurrency }: { preview: service.DocumentTotals | null; displayCurrency: Currency }) {
  return <div className="formtotals" aria-live="polite">{preview ? <><div className="sumrow"><span>Base</span><Money cents={preview.subtotalCents} currency={displayCurrency} /></div><div className="sumrow"><span>Extras</span><Money cents={preview.extrasCents} currency={displayCurrency} /></div><div className="sumrow sumrow--strong"><span>Total · moneda del switch</span><Money cents={preview.totalCents} currency={displayCurrency} /></div></> : <span className="muted">Ingresa los ítems para calcular el total.</span>}</div>
}

function QuickCreateModals({ quickKind, quickName, clientId, addingAddress, onClose, onSavedClient, onSavedSupplier, onSavedCard, onSavedProduct, onSavedAddress, onCloseAddress }: { quickKind: QuickKind | null; quickName: string; clientId: string; addingAddress: boolean; onClose: () => void; onSavedClient: (row: any) => void; onSavedSupplier: (row: any) => void; onSavedCard: (row: any) => void; onSavedProduct: (row: any) => void; onSavedAddress: (row: any) => void; onCloseAddress: () => void }) {
  return <>
    {quickKind === 'client' && <ClientModal initialName={quickName} onClose={onClose} onSaved={onSavedClient} />}
    {quickKind === 'supplier' && <SupplierModal initialName={quickName} onClose={onClose} onSaved={onSavedSupplier} />}
    {quickKind === 'card' && <CardModal initialName={quickName} onClose={onClose} onSaved={onSavedCard} />}
    {quickKind === 'product' && <ProductModal initialName={quickName} onClose={onClose} onSaved={onSavedProduct} />}
    {addingAddress && <AddressModal clientId={Number(clientId)} onClose={onCloseAddress} onSaved={onSavedAddress} />}
  </>
}

function ConceptInput({ value, onChange }: { value: string; onChange: (value: string) => void }) {
  const id = useId()
  const [suggestions, setSuggestions] = useState<string[]>([])
  useEffect(() => {
    const timer = window.setTimeout(() => { void api.Catalog.Suggest('concept', value).then((rows) => setSuggestions(rows.map((row) => String(row.label ?? '')))).catch(() => setSuggestions([])) }, 180)
    return () => window.clearTimeout(timer)
  }, [value])
  return <><TextInput list={id} value={value} onChange={(event) => onChange(event.target.value)} /><datalist id={id}>{suggestions.map((suggestion) => <option value={suggestion} key={suggestion} />)}</datalist></>
}
