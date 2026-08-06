type ExportFormat = 'csv' | 'json' | 'txt';

export function downloadBlob(blob: Blob, filename: string): void {
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = filename;
  document.body.appendChild(a);
  a.click();
  document.body.removeChild(a);
  URL.revokeObjectURL(url);
}

export function downloadText(content: string, filename: string, mimeType = 'text/plain;charset=utf-8'): void {
  const blob = new Blob([content], { type: mimeType });
  downloadBlob(blob, filename);
}

export function downloadCSV<T>(rows: T[], filename: string, columns?: { key: keyof T; header: string }[]): void {
  if (rows.length === 0) {
    downloadText('', filename, 'text/csv;charset=utf-8');
    return;
  }
  const cols = columns ?? (Object.keys(rows[0] as object) as (keyof T)[]).map((k) => ({ key: k, header: String(k) }));
  const header = cols.map((c) => escapeCSV(c.header)).join(',');
  const body = rows
    .map((row) =>
      cols
        .map((c) => {
          const v = row[c.key];
          return escapeCSV(v == null ? '' : String(v));
        })
        .join(','),
    )
    .join('\n');
  const content = `${header}\n${body}`;
  downloadText(content, filename, 'text/csv;charset=utf-8');
}

export function downloadJSON<T>(data: T, filename: string): void {
  const content = JSON.stringify(data, null, 2);
  downloadText(content, filename, 'application/json;charset=utf-8');
}

function escapeCSV(v: string): string {
  if (/[",\n\r]/.test(v)) {
    return `"${v.replace(/"/g, '""')}"`;
  }
  return v;
}

export function exportData<T>(data: T[], format: ExportFormat, filename: string, columns?: { key: keyof T; header: string }[]): void {
  switch (format) {
    case 'csv':
      downloadCSV(data, filename, columns);
      break;
    case 'json':
      downloadJSON(data, filename);
      break;
    case 'txt':
      downloadText(data.map((r) => JSON.stringify(r)).join('\n'), filename, 'text/plain;charset=utf-8');
      break;
  }
}
