import { ALWAYS_HIDDEN_FIELDS } from '../constants/config';

export function filterHiddenFields(obj: Record<string, unknown> | unknown[], options?: { showStatus?: boolean }): Record<string, unknown> | unknown[] | unknown {
  if (typeof obj !== 'object' || obj === null) return obj;
  if (Array.isArray(obj)) return obj.map(item => filterHiddenFields(item as Record<string, unknown>, options));

  const filtered: Record<string, unknown> = {};
  for (const [key, value] of Object.entries(obj)) {
    if ((ALWAYS_HIDDEN_FIELDS as readonly string[]).includes(key)) continue;
    if (options && !options.showStatus && key === 'status') continue;
    filtered[key] = filterHiddenFields(value as Record<string, unknown> | unknown[], options);
  }
  return filtered;
}
