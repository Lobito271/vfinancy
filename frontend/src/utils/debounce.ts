import { useEffect, useState } from 'react';

export function useDebounce<T>(value: T, delay = 300): T {
  const [debounced, setDebounced] = useState(value);
  useEffect(() => {
    const id = setTimeout(() => setDebounced(value), delay);
    return () => clearTimeout(id);
  }, [value, delay]);
  return debounced;
}

export function useDebouncedCallback<T extends (...args: never[]) => unknown>(
  fn: T,
  delay = 300,
): (...args: Parameters<T>) => void {
  const [timer, setTimer] = useState<ReturnType<typeof setTimeout> | null>(null);
  useEffect(() => () => {
    if (timer) clearTimeout(timer);
  }, [timer]);
  return (...args: Parameters<T>) => {
    if (timer) clearTimeout(timer);
    setTimer(setTimeout(() => fn(...args), delay));
  };
}

export function debounce<T extends (...args: never[]) => unknown>(fn: T, delay = 300): T {
  let timer: ReturnType<typeof setTimeout> | null = null;
  return ((...args: Parameters<T>) => {
    if (timer) clearTimeout(timer);
    timer = setTimeout(() => fn(...args), delay);
  }) as T;
}

export function throttle<T extends (...args: never[]) => unknown>(fn: T, ms = 200): T {
  let last = 0;
  return ((...args: Parameters<T>) => {
    const now = Date.now();
    if (now - last >= ms) {
      last = now;
      fn(...args);
    }
  }) as T;
}
