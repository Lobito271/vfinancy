import { useEffect, useRef, useState } from 'react'
import type { ReactNode } from 'react'
import { MoreHorizontal } from 'lucide-react'
import { cx } from '../lib/format'

export interface MenuItem {
  label: string
  icon?: ReactNode
  onSelect?: () => void
  danger?: boolean
  disabled?: boolean
}

export function RowMenu({ items, label = 'Acciones' }: { items: MenuItem[]; label?: string }) {
  const [open, setOpen] = useState(false)
  const root = useRef<HTMLDivElement>(null)
  const trigger = useRef<HTMLButtonElement>(null)
  const close = () => { setOpen(false); trigger.current?.focus() }

  useEffect(() => {
    if (!open) return
    const onKeyDown = (event: KeyboardEvent) => {
      if (event.key === 'Escape') { event.preventDefault(); close(); return }
      if (event.key !== 'ArrowDown' && event.key !== 'ArrowUp') return
      event.preventDefault()
      const options = Array.from(root.current?.querySelectorAll<HTMLButtonElement>('.menu__item') ?? []).filter((item) => !item.disabled)
      if (!options.length) return
      const current = options.indexOf(document.activeElement as HTMLButtonElement)
      const next = event.key === 'ArrowDown' ? (current + 1) % options.length : (current - 1 + options.length) % options.length
      options[next]?.focus()
    }
    const closeOutside = (event: MouseEvent) => {
      if (!root.current?.contains(event.target as Node)) close()
    }
    document.addEventListener('keydown', onKeyDown)
    document.addEventListener('mousedown', closeOutside)
    return () => {
      document.removeEventListener('keydown', onKeyDown)
      document.removeEventListener('mousedown', closeOutside)
    }
  }, [open])

  return <div className={cx('menu', open && 'menu--open')} ref={root}>
    <button ref={trigger} type="button" className={cx('iconbtn', open && 'is-active')} aria-label={label} aria-haspopup="menu" aria-expanded={open} onClick={() => setOpen((value) => !value)}><MoreHorizontal size={15} /></button>
    {open && <div className="menu__pop" role="menu">{items.map((item) => <button key={item.label} type="button" role="menuitem" disabled={item.disabled} className={cx('menu__item', item.danger && 'menu__item--danger')} onClick={() => { close(); item.onSelect?.() }}>{item.icon}{item.label}</button>)}</div>}
  </div>
}
