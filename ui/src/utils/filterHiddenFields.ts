import { ALWAYS_HIDDEN_FIELDS } from '../constants/config';

export function filterHiddenFields(obj: any, options?: { showStatus?: boolean }): any {
  if (typeof obj !== 'object' || obj === null) return obj;
  if (Array.isArray(obj)) return obj.map(item => filterHiddenFields(item, options));

  const filtered: any = {};
  for (const [key, value] of Object.entries(obj)) {
    if ((ALWAYS_HIDDEN_FIELDS as readonly string[]).includes(key)) continue;
    if (options && !options.showStatus && key === 'status') continue;
    filtered[key] = filterHiddenFields(value, options);
  }
  return filtered;
}
