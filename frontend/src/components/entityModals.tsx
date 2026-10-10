import { useEffect, useId, useState } from 'react'
import { service } from '../../wailsjs/go/models'
import { api } from '../data/api'
import type { Currency } from '../data/api'
import { useStore } from '../data/store'
import { parseError } from '../lib/errors'
import { centsText, textToCents } from '../lib/money'
import { Field, Modal, ModalActions, MoneyInput, Seg, TextInput } from './ui'

type Row = Record<string, any>

function ContactModal({ kind, existing, initialName, onClose, onSaved }: { kind: 'client' | 'supplier'; existing?: Row; initialName?: string; onClose: () => void; onSaved?: (row: Row) => void }) {
  const formId = useId()
  const [name, setName] = useState(String(existing?.name ?? initialName ?? ''))
  const [phone, setPhone] = useState(String(existing?.phone ?? ''))
  const [note, setNote] = useState(String(existing?.note ?? ''))
  const [busy, setBusy] = useState(false)
  const toast = useStore((state) => state.toast)
  const label = kind === 'client' ? 'cliente' : 'proveedor'
  return <Modal title={existing ? `Editar ${label}` : `Nuevo ${label}`} onClose={onClose} footer={<ModalActions onClose={onClose} busy={busy} formId={formId} label="Guardar" />}><form id={formId} className="stack" onSubmit={async (event) => { event.preventDefault(); setBusy(true); try { const id = Number(existing?.id ?? 0); const saved = kind === 'client' ? await api.Catalog.SaveClient(new service.Client({ id, name, phone, note })) : await api.Catalog.SaveSupplier(new service.Supplier({ id, name, phone, note })); useStore.getState().refresh(); onSaved?.(saved as Row); toast(existing ? 'Contacto actualizado' : 'Contacto creado'); onClose() } catch (error) { toast(parseError(error).message ?? 'No se pudo guardar el contacto', 'bad') } finally { setBusy(false) } }}><Field label="Nombre" required><TextInput autoFocus value={name} onChange={(event) => setName(event.target.value)} /></Field><Field label="Teléfono"><TextInput type="tel" value={phone} onChange={(event) => setPhone(event.target.value)} /></Field><Field label="Nota"><textarea className="input" rows={3} value={note} onChange={(event) => setNote(event.target.value)} /></Field></form></Modal>
}

export function ClientModal(props: { existing?: Row; initialName?: string; onClose: () => void; onSaved?: (row: Row) => void }) { return <ContactModal kind="client" {...props} /> }
export function SupplierModal(props: { existing?: Row; initialName?: string; onClose: () => void; onSaved?: (row: Row) => void }) { return <ContactModal kind="supplier" {...props} /> }

export function AddressModal({ clientId, address, onClose, onSaved }: { clientId: number; address?: Row; onClose: () => void; onSaved?: (row: Row) => void }) {
  const formId = useId()
  const [region, setRegion] = useState(String(address?.region ?? ''))
  const [street, setStreet] = useState(String(address?.address ?? ''))
  const [busy, setBusy] = useState(false)
  const toast = useStore((state) => state.toast)
  return <Modal title={address ? 'Editar dirección' : 'Nueva dirección'} onClose={onClose} footer={<ModalActions onClose={onClose} busy={busy} formId={formId} label="Guardar dirección" />}><form id={formId} className="stack" onSubmit={async (event) => { event.preventDefault(); setBusy(true); try { const saved = await api.Catalog.SaveAddress(new service.Address({ id: Number(address?.id ?? 0), clientId, region, address: street })); useStore.getState().refresh(); onSaved?.(saved as Row); toast('Dirección guardada'); onClose() } catch (error) { toast(parseError(error).message ?? 'No se pudo guardar la dirección', 'bad') } finally { setBusy(false) } }}><Field label="Región" required><TextInput value={region} onChange={(event) => setRegion(event.target.value)} /></Field><Field label="Dirección" required><TextInput value={street} onChange={(event) => setStreet(event.target.value)} /></Field></form></Modal>
}

export function ProductModal({ existing, initialName, onClose, onSaved }: { existing?: Row; initialName?: string; onClose: () => void; onSaved?: (row: Row) => void }) {
  const formId = useId()
  const [sku, setSku] = useState(String(existing?.sku ?? ''))
  const [name, setName] = useState(String(existing?.name ?? initialName ?? ''))
  const [cost, setCost] = useState(existing ? centsText(Number(existing.original_default_cost_cents ?? existing.default_cost_cents ?? 0)) : '')
  const [price, setPrice] = useState(existing ? centsText(Number(existing.original_default_price_cents ?? existing.default_price_cents ?? 0)) : '')
  const [costCurrency, setCostCurrency] = useState<Currency>((existing?.default_cost_currency ?? 'USD') as Currency)
  const [priceCurrency, setPriceCurrency] = useState<Currency>((existing?.default_price_currency ?? 'PEN') as Currency)
  const [busy, setBusy] = useState(false)
  const toast = useStore((state) => state.toast)
  useEffect(() => { if (!existing) void api.Catalog.NextCode('product').then(setSku).catch(() => {}) }, [existing])
  return <Modal title={existing ? 'Editar producto' : 'Nuevo producto'} onClose={onClose} footer={<ModalActions onClose={onClose} busy={busy} formId={formId} label={existing ? 'Guardar cambios' : 'Guardar producto'} />}><form id={formId} className="stack" onSubmit={async (event) => { event.preventDefault(); setBusy(true); try { const saved = await api.Catalog.SaveProduct(new service.Product({ id: Number(existing?.id ?? 0), sku, name, defaultCostCents: textToCents(cost) ?? 0, defaultCostCurrency: costCurrency, defaultPriceCents: textToCents(price) ?? 0, defaultPriceCurrency: priceCurrency })); useStore.getState().refresh(); onSaved?.(saved as Row); toast(existing ? 'Producto actualizado' : 'Producto creado'); onClose() } catch (error) { toast(parseError(error).message ?? 'No se pudo guardar el producto', 'bad') } finally { setBusy(false) } }}><Field label="Nombre" required><TextInput value={name} onChange={(event) => setName(event.target.value)} /></Field><Field label="SKU" required><TextInput className="input--mono" value={sku} onChange={(event) => setSku(event.target.value)} /></Field><Field label="Costo por defecto"><MoneyInput currency={costCurrency} value={cost} onChange={setCost} /></Field><Seg options={[{ value: 'USD', label: '$' }, { value: 'PEN', label: 'S/' }]} value={costCurrency} onChange={setCostCurrency} /><Field label="Precio por defecto"><MoneyInput currency={priceCurrency} value={price} onChange={setPrice} /></Field><Seg options={[{ value: 'PEN', label: 'S/' }, { value: 'USD', label: '$' }]} value={priceCurrency} onChange={setPriceCurrency} /></form></Modal>
}

export function CardModal({ existing, initialName, onClose, onSaved }: { existing?: Row; initialName?: string; onClose: () => void; onSaved?: (row: Row) => void }) {
  const formId = useId()
  const [name, setName] = useState(String(existing?.name ?? initialName ?? ''))
  const [last4, setLast4] = useState(String(existing?.last4 ?? ''))
  const [cutDay, setCutDay] = useState(String(existing?.cutDay ?? existing?.cut_day ?? 20))
  const [dueDay, setDueDay] = useState(String(existing?.dueDay ?? existing?.due_day ?? 5))
  const [limit, setLimit] = useState(existing?.limitCents == null ? '' : centsText(Number(existing.originalLimitCents ?? existing.limitCents)))
  const [currency, setCurrency] = useState<Currency>((existing?.limitCurrency ?? 'PEN') as Currency)
  const [busy, setBusy] = useState(false)
  const toast = useStore((state) => state.toast)
  return <Modal title={existing ? 'Editar tarjeta' : 'Nueva tarjeta'} onClose={onClose} footer={<ModalActions onClose={onClose} busy={busy} formId={formId} label={existing ? 'Guardar cambios' : 'Guardar tarjeta'} />}><form id={formId} className="stack" onSubmit={async (event) => { event.preventDefault(); setBusy(true); try { const rate = await api.Settings.DefaultTC(); const saved = await api.Catalog.SaveCard(new service.Card({ id: Number(existing?.id ?? 0), name, last4, cutDay: Number(cutDay), dueDay: Number(dueDay), limitCents: limit ? textToCents(limit) : undefined, limitCurrency: currency, limitTc: rate })); useStore.getState().refresh(); onSaved?.(saved as Row); toast(existing ? 'Tarjeta actualizada' : 'Tarjeta creada'); onClose() } catch (error) { toast(parseError(error).message ?? 'No se pudo guardar la tarjeta', 'bad') } finally { setBusy(false) } }}><Field label="Nombre" required><TextInput value={name} onChange={(event) => setName(event.target.value)} /></Field><Field label="Últimos 4 dígitos" required><TextInput maxLength={4} inputMode="numeric" value={last4} onChange={(event) => setLast4(event.target.value.replace(/\D/g, '').slice(0, 4))} /></Field><div className="formgrid"><Field label="Día de corte" required><TextInput type="number" min="1" max="31" value={cutDay} onChange={(event) => setCutDay(event.target.value)} /></Field><Field label="Día de pago" required><TextInput type="number" min="1" max="31" value={dueDay} onChange={(event) => setDueDay(event.target.value)} /></Field></div><Field label="Límite (opcional)"><MoneyInput currency={currency} value={limit} onChange={setLimit} /></Field><Seg options={[{ value: 'PEN', label: 'S/' }, { value: 'USD', label: '$' }]} value={currency} onChange={setCurrency} /></form></Modal>
}
