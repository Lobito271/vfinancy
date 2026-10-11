import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { ArrowLeft, MapPin, Pencil, Phone, Plus } from 'lucide-react'
import { service } from '../../wailsjs/go/models'
import { api, allPages } from '../data/api'
import { useStore } from '../data/store'
import { statusView } from '../data/status'
import { parseError } from '../lib/errors'
import { fmtDate, monthLabel } from '../lib/format'
import { Badge, ColHead, Empty, Loading, Money, StatusBadge, ViewHead } from '../components/ui'
import { AddressModal, ClientModal, SupplierModal } from '../components/entityModals'
import { FilterBar, useServerFilters } from '../components/filters'

type Row = Record<string, any>
type Kind = 'client' | 'supplier'

export function ContactsPage({ kind, id = 0 }: { kind: Kind; id?: number }) {
  const currency = useStore((state) => state.currency)
  const revision = useStore((state) => state.revision)
  const toast = useStore((state) => state.toast)
  const navigate = useNavigate()
  const entity = kind === 'client' ? 'clients' : 'suppliers'
  const route = kind === 'client' ? 'clientes' : 'proveedores'
  const filters = useServerFilters(entity)
  const filterKey = JSON.stringify(filters.filters.map((f) => [f.field, f.op, f.value, f.end]))
  const [records, setRecords] = useState<Row[]>([])
  const [history, setHistory] = useState<Row[]>([])
  const [addresses, setAddresses] = useState<Row[]>([])
  const [loading, setLoading] = useState(true)
  const [detailLoading, setDetailLoading] = useState(false)
  const [creating, setCreating] = useState(false)
  const [addressModal, setAddressModal] = useState<Row | null | false>(false)
  const [editing, setEditing] = useState(false)

  useEffect(() => {
    let active = true
    setLoading(true)
    const fetch = kind === 'client' ? api.Catalog.ListClients : api.Catalog.ListSuppliers
    void allPages(fetch, currency, id ? [] : filters.filters).then((rows) => { if (active) { setRecords(rows); setLoading(false) } }).catch((error) => { if (active) { toast(parseError(error).message ?? 'No se pudo cargar la lista', 'bad'); setLoading(false) } })
    return () => { active = false }
  }, [kind, currency, revision, filterKey, id, filters.filters, toast])

  const record = records.find((row) => Number(row.id) === id)
  useEffect(() => {
    if (!id) { setHistory([]); setAddresses([]); setEditing(false) }
  }, [id])
  useEffect(() => {
    if (!id) return
    let active = true
    setDetailLoading(true)
    void (async () => {
      const selected = record ?? (kind === 'client' ? await allPages(api.Catalog.ListClients, currency) : await allPages(api.Catalog.ListSuppliers, currency)).find((row) => Number(row.id) === id)
      if (!selected) { if (active) setDetailLoading(false); return }
      let nextAddresses: Row[] = []
      if (kind === 'client') nextAddresses = await api.Catalog.ListAddresses(id) as Row[]
      const byName = [new service.Filter({ field: kind === 'client' ? 'client' : 'supplier', op: 'contains', value: selected.name })]
      const purchases = await allPages(api.Purchases.List, currency, byName)
      const rows: Row[] = kind === 'supplier'
        ? purchases.filter((row) => Number(row.supplier_id) === id).map((row) => ({ ...row, historyKind: 'purchase' }))
        : [
            ...purchases.filter((row) => Number(row.for_client_id) === id).map((row) => ({ ...row, historyKind: 'purchase' })),
            ...(await allPages(api.Sales.List, currency, byName)).filter((row) => Number(row.client_id) === id).map((row) => ({ ...row, historyKind: row.kind })),
          ]
      if (active) { setAddresses(nextAddresses); setHistory(rows); setDetailLoading(false) }
    })().catch((error) => { if (active) { setDetailLoading(false); toast(parseError(error).message ?? 'No se pudo cargar el historial', 'bad') } })
    return () => { active = false }
  }, [id, kind, currency, revision, record?.id, record, toast])

  useEffect(() => { useStore.getState().registerNew(() => setCreating(true)); return () => useStore.getState().registerNew(null) }, [])
  const title = kind === 'client' ? 'Clientes' : 'Proveedores'

  if (id) {
    if (loading || detailLoading) return <Loading />
    if (!record) return <Empty>{kind === 'client' ? 'Cliente' : 'Proveedor'} no encontrado</Empty>
    const state = statusView('clients', record.status)
    const grouped = groupHistory(history)
    return <>
      <ViewHead title={record.name} subtitle={[record.phone || 'sin teléfono', record.note || 'sin nota'].filter(Boolean).join(' · ')} actions={<div className="hstack"><button className="iconbtn" type="button" aria-label={`Volver a ${title.toLowerCase()}`} onClick={() => navigate(`/${route}`)}><ArrowLeft size={16}/></button><button className="btn btn--sm" type="button" onClick={() => setEditing(true)}><Pencil size={13}/> Editar</button></div>} />
      {kind === 'client' && <><div className="kpis"><div className="kpi"><span className="kpi__label">Saldo por cobrar</span><span className="kpi__value">{Number(record.balanceCents ?? 0) > 0 ? <Money cents={record.balanceCents}/> : <span className="muted">0</span>}</span></div><div className="kpi"><span className="kpi__label">Estado</span><span className="kpi__value">{state ? <Badge state={state}/> : 'Sin saldo'}</span></div></div><h2 className="sectitle">Direcciones de envío</h2>{addresses.length ? <ul className="addrlist">{addresses.map((address) => <li key={address.id}><MapPin size={12}/><b>{address.region}</b> / {address.address}<button className="btn btn--xs" type="button" onClick={() => setAddressModal(address)}>Editar</button></li>)}</ul> : <span className="muted">Sin direcciones guardadas</span>}<button className="btn btn--xs" type="button" onClick={() => setAddressModal(null)}><Plus size={13}/> Agregar dirección</button></>}
      <h2 className="sectitle">{kind === 'client' ? 'Historial por mes' : 'Compras por mes'}</h2>
      {!grouped.length ? <Empty>No hay historial registrado</Empty> : grouped.map(([month, rows]) => <section className="jmonth" key={month}><header className="jmonth__head"><h2>{monthLabel(month)}</h2><span className="jmonth__stats">{rows.length} registro{rows.length === 1 ? '' : 's'}</span></header><ColHead cols={['Fecha · Código','Estado','Tipo',{label:'Monto',num:true},'']}/><div className="jmonth__list">{rows.map((row) => { const purchase = row.historyKind === 'purchase'; const entityName = purchase ? 'purchases' : 'sales'; const view = statusView(entityName, row.status); const route = purchase ? 'compras' : row.kind === 'shipment' ? 'envios' : 'ventas'; return <button key={`${row.historyKind}-${row.id}`} type="button" className={`listrow listrow--click ${view ? `stripe stripe--${view.tone}` : ''} ${view?.key === 'voided' ? 'listrow--void' : ''}`} onClick={() => navigate(`/${route}?abrir=${encodeURIComponent(row.code)}`)}><span className="listrow__main"><span className="listrow__title">{fmtDate(row.date)}</span><span className="listrow__meta">{row.code} · {purchase ? 'Compra (encargo)' : row.kind === 'shipment' ? 'Envío' : 'Venta'}</span></span><span className="listrow__state"><StatusBadge state={view}/></span><span className="listrow__date">{purchase ? 'Compra (encargo)' : row.kind === 'shipment' ? 'Envío' : 'Venta'}</span><span className="listrow__amount num"><Money cents={row.totalCents ?? 0}/></span></button>})}</div></section>)}
      {addressModal !== false && <AddressModal clientId={id} address={addressModal || undefined} onClose={() => setAddressModal(false)}/>}
      {editing && (kind === 'client' ? <ClientModal existing={record} onClose={() => setEditing(false)}/> : <SupplierModal existing={record} onClose={() => setEditing(false)}/>)}
    </>
  }

  return <>
    <ViewHead title={title} subtitle="Entidades de consulta con ficha propia" actions={<button className="btn btn--primary btn--sm" type="button" onClick={() => setCreating(true)}><Plus size={14}/> Nuevo {kind === 'client' ? 'cliente' : 'proveedor'}</button>}/>
    <div className="pagebar"></div>{((records.length)>0 || filters.filters.length>0) && <div className="listtoolbar"><FilterBar state={filters}/></div>}
    {loading ? <Loading/> : !records.length ? <Empty>Sin {title.toLowerCase()} que coincidan con los filtros</Empty> : <><ColHead cols={[kind === 'client' ? 'Cliente' : 'Proveedor', kind === 'client' ? 'Estado' : '', '', kind === 'client' ? { label: 'Saldo', num: true } : '', '']}/><div className="catlist">{records.map((row) => { const state = kind === 'client' ? statusView('clients', row.status) : null; return <button key={row.id} type="button" className={`listrow listrow--click ${state ? `stripe stripe--${state.tone}` : ''} ${state?.key === 'voided' ? 'listrow--void' : ''}`} onClick={() => navigate(`/${route}/${row.id}`)}><span className="listrow__main"><span className="listrow__title">{row.name}</span><span className="listrow__meta"><Phone size={11}/>{row.phone || 'sin teléfono'}</span></span>{kind === 'client' && <span className="listrow__state"><StatusBadge state={state}/></span>}{kind === 'client' && <span className="listrow__amount num">{Number(row.balanceCents ?? 0) > 0 ? <Money cents={row.balanceCents}/> : <span className="muted">sin saldo</span>}</span>}</button>})}</div></>}
    {creating && (kind === 'client' ? <ClientModal onClose={() => setCreating(false)}/> : <SupplierModal onClose={() => setCreating(false)}/>)}
  </>
}

function groupHistory(rows: Row[]): Array<[string, Row[]]> { const groups = new Map<string, Row[]>(); for (const row of rows) { const month = String(row.date).slice(0,7); groups.set(month,[...(groups.get(month) ?? []),row]) } return [...groups.entries()].sort((a,b)=>b[0].localeCompare(a[0])) }
