import { useEffect, useEffectEvent, useId, useRef, useState } from 'react'
import type { InputHTMLAttributes, ReactNode } from 'react'
import { createPortal } from 'react-dom'
import { X } from 'lucide-react'
import { cx } from '../lib/format'
import { formatMoney, sanitizeMoney, sanitizeTc, tcFromText, tcText, type Currency } from '../lib/money'
import { useStore } from '../data/store'
import { api } from '../data/api'
import type { StatusView, Tone } from '../data/status'

export function Badge({ state, tone, children }: { state?: StatusView | null; tone?: Tone; children?: ReactNode }) {
  if (!state && !children) return null
  return <span className={cx('badge', (state?.tone ?? tone) && `badge--${state?.tone ?? tone}`)}>{children ?? state?.label}</span>
}
export function StatusBadge({ state }: { state?: StatusView | null }) { return <Badge state={state} /> }
export function Money({ cents, currency, original, documentCurrency, tc, className }: { cents: number; currency?: Currency; original?: number; documentCurrency?: Currency; tc?: number; className?: string }) {
  const display = useStore((s) => s.currency)
  const supplied = currency ?? display
  const mainCurrency = original !== undefined && documentCurrency === supplied && supplied !== display ? display : supplied
  const originalCurrency = documentCurrency ?? supplied
  const showOriginal = original !== undefined && originalCurrency !== mainCurrency
  return <span className={cx('money', mainCurrency === 'USD' && 'money--usd', className)}>{formatMoney(cents, mainCurrency)}{showOriginal && <small className="money__orig">{formatMoney(original, originalCurrency)}{tc ? ` · TC ${tcText(tc)}` : ''}</small>}</span>
}
export function Field({ label, children, error, required }: { label: string; children: ReactNode; error?: string; required?: boolean }) {
  const id = useId()
  return <label className="field" htmlFor={id}><span className="field__label">{label}{required && <b className="field__req"> *</b>}</span>{typeof children === 'object' && children !== null ? children : <span id={id}>{children}</span>}{error && <small className="field__error">{error}</small>}</label>
}
export function TextInput(props: InputHTMLAttributes<HTMLInputElement>) { return <input {...props} className={cx('input', props.className)} /> }
export function MoneyInput({ currency, value, onChange, label }: { currency: Currency; value: string; onChange: (text: string) => void; label?: string }) {
  return <div className="moneyfield"><span className="moneyfield__symbol" aria-hidden="true">{currency === 'USD' ? '$' : 'S/'}</span><input className="input input--tabs" inputMode="decimal" aria-label={label} value={value} onChange={(event) => onChange(sanitizeMoney(event.target.value))} /></div>
}
export function TcField({ value, onChange }: { value: string; onChange: (value: string) => void }) {
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const fallback = useStore((s) => s.settings.tc_fallback ?? '37500')
  return <div className="tcinput"><div className="tcinput__row"><input className="input input--tabs tcinput__field" inputMode="decimal" aria-label="TC (S/ por $)" value={value} onChange={(event) => onChange(sanitizeTc(event.target.value))} /><button className="btn btn--sm" type="button" disabled={loading} onClick={async () => { setLoading(true); setError(''); try { const rate = await api.Settings.FetchTC(); onChange(tcText(rate)) } catch { setError('No se pudo consultar el TC'); if (!value) onChange(tcText(Number(fallback))) } finally { setLoading(false) } }}>{loading ? 'Consultando…' : 'Consultar'}</button></div>{error && <small className="field__error">{error}</small>}</div>
}
export function Seg<T extends string>({ options, value, onChange, label, small = false }: { options: Array<{ value: T; label: string }>; value: T; onChange: (value: T) => void; label?: string; small?: boolean }) {
  return <div className={cx('seg', small && 'seg--sm')} role="group" aria-label={label}>{options.map((item) => <button key={item.value} type="button" aria-pressed={value === item.value} className={cx('seg__opt', value === item.value && 'is-active')} onClick={() => onChange(item.value)}>{item.label}</button>)}</div>
}
export function ViewHead({ title, subtitle, actions }: { title: string; subtitle?: string; actions?: ReactNode }) { return <header className="viewhead"><div><h1 className="viewhead__title">{title}</h1>{subtitle && <p className="viewhead__sub">{subtitle}</p>}</div>{actions && <div className="viewhead__actions">{actions}</div>}</header> }
export function Card({ title, subtitle, children, className = '', flush = false }: { title: string; subtitle?: string; children: ReactNode; className?: string; flush?: boolean }) { return <section className={cx('card', className)}><header className="card__head"><div><h3 className="card__title">{title}</h3>{subtitle && <p className="card__sub">{subtitle}</p>}</div></header><div className={cx('card__body', flush && 'card__body--flush')}>{children}</div></section> }
export type ColSpec = string | { label: string; num?: boolean }
export function ColHead({ cols }: { cols: ColSpec[] }) {
  return <div className="col-head" aria-hidden="true">{cols.map((col, index) => <span key={index} className={typeof col === 'object' && col.num ? 'col-head--num' : undefined}>{typeof col === 'string' ? col : col.label}</span>)}</div>
}
export function Empty({ children }: { children: ReactNode }) { return <div className="empty">{children}</div> }
export function Loading() { return <div className="skls" aria-busy="true"><div className="skl skl--row" /><div className="skl skl--row" /><div className="skl skl--row" /></div> }
let modalSequence = 0
const modalStack: number[] = []
export function Modal({ title, onClose, children, footer, wide = false }: { title: string; onClose: () => void; children: ReactNode; footer?: ReactNode; wide?: boolean }) {
  const sectionRef = useRef<HTMLElement>(null)
  const onClosed = useEffectEvent(onClose)
  useEffect(() => {
    const section = sectionRef.current
    const id = ++modalSequence
    const previous = document.activeElement as HTMLElement | null
    modalStack.push(id)
    document.body.classList.add('modal-lock')
    const focusable = () => Array.from(section?.querySelectorAll<HTMLElement>('a[href], button:not([disabled]), textarea, input, select, [tabindex]:not([tabindex="-1"])') ?? []).filter((el) => el.offsetParent !== null)
    if (modalStack.at(-1) === id && !section?.contains(document.activeElement)) {
      const bodyFirst = Array.from(section?.querySelectorAll<HTMLElement>('.modal__body button, .modal__body input, .modal__body select, .modal__body textarea, .modal__body a[href]') ?? []).find((el) => el.offsetParent !== null)
      ;(section?.querySelector<HTMLElement>('[autofocus]') ?? bodyFirst ?? focusable()[0] ?? section)?.focus()
    }
    const key = (event: KeyboardEvent) => {
      if (modalStack.at(-1) !== id) return
      if (event.key === 'Escape') { onClosed(); return }
      if (event.key === 'Tab') {
        const items = focusable()
        if (!items.length) { event.preventDefault(); return }
        const current = document.activeElement as HTMLElement
        const index = items.indexOf(current)
        if (event.shiftKey && (index <= 0 || index === -1)) { event.preventDefault(); items[items.length - 1].focus() }
        else if (!event.shiftKey && (index === items.length - 1 || index === -1)) { event.preventDefault(); items[0].focus() }
      }
    }
    window.addEventListener('keydown', key)
    return () => {
      window.removeEventListener('keydown', key)
      const index = modalStack.indexOf(id)
      if (index >= 0) modalStack.splice(index, 1)
      if (!modalStack.length) {
        document.body.classList.remove('modal-lock')
        if (previous && previous.isConnected && !section?.contains(previous)) previous.focus()
      }
    }
  }, [])
  return createPortal(<div className="modal__backdrop"><button type="button" className="modal__backdrop__hit" aria-label="Cerrar" tabIndex={-1} onMouseDown={onClose} /><section ref={sectionRef} tabIndex={-1} className={cx('modal', wide && 'modal--wide')} role="dialog" aria-modal="true" aria-label={title}><header className="modal__head"><h2 className="modal__title">{title}</h2><button type="button" className="iconbtn" aria-label="Cerrar" onClick={onClose}><X size={16} /></button></header><div className="modal__body">{children}</div>{footer && <footer className="modal__foot">{footer}</footer>}</section></div>, document.body)
}
export function ModalActions({ onClose, busy, label, formId }: { onClose: () => void; busy?: boolean; label: string; formId?: string }) { return <><button className="btn" type="button" onClick={onClose} disabled={busy}>Cancelar</button><button className="btn btn--primary" type="submit" form={formId} disabled={busy}>{busy ? 'Guardando…' : label}</button></> }
export function Toasts() { const toasts = useStore((s) => s.toasts); return createPortal(<div className="toasts" aria-live="polite">{toasts.map((toast) => <div className={cx('toast', `toast--${toast.tone}`)} key={toast.id}>{toast.msg}</div>)}</div>, document.body) }
export function tcValue(text: string): number | null { return tcFromText(text) }
