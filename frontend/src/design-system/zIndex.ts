export const ZIndex = {
  hide: -1,
  auto: 'auto',
  base: 0,
  dropdown: 1000,
  sticky: 1100,
  fixed: 1200,
  topbar: 1300,
  sidebarBackdrop: 1400,
  drawer: 1500,
  modalBackdrop: 1600,
  modal: 1700,
  popover: 1800,
  tooltip: 1900,
  toast: 2000,
  notification: 2100,
  max: 9999,
} as const;

export type ZIndexKey = keyof typeof ZIndex;
