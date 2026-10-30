export const MillimetersPerInch = 25.4;
export const PointsPerInch = 72.0;
export const DefaultScreenDPI = 96.0;

export function mmToPt(mm: number): number {
  return (mm * PointsPerInch) / MillimetersPerInch;
}

export function ptToMm(pt: number): number {
  return (pt * MillimetersPerInch) / PointsPerInch;
}

export function mmToPx(mm: number, dpi: number = DefaultScreenDPI): number {
  return (mm * dpi) / MillimetersPerInch;
}

export function pxToMm(px: number, dpi: number = DefaultScreenDPI): number {
  return (px * MillimetersPerInch) / dpi;
}

export function replacePlaceholders(text: string, variables: Record<string, string>): string {
  // Normalize lookup keys: lowercase and strip underscores/hyphens
  const normMap: Record<string, string> = {};
  for (const [k, v] of Object.entries(variables)) {
    normMap[k.toLowerCase().replace(/[-_]/g, '')] = v;
  }

  return text.replace(/\{\{([a-zA-Z0-9_-]+)\}\}/g, (match, key) => {
    if (variables[key] !== undefined) {
      return variables[key];
    }
    const normKey = key.toLowerCase().replace(/[-_]/g, '');
    if (normMap[normKey] !== undefined) {
      return normMap[normKey];
    }
    return match;
  });
}

export function toTitleCase(s: string): string {
  const trimmed = s.trim();
  if (!trimmed) return '';

  let capitalizeNext = true;
  let result = '';

  for (let i = 0; i < trimmed.length; i++) {
    const char = trimmed[i];
    const isLetter = /\p{L}/u.test(char);

    if (isLetter) {
      if (capitalizeNext) {
        result += char.toUpperCase();
        capitalizeNext = false;
      } else {
        if (i > 1 && trimmed[i - 1] === '.') {
          result += char.toUpperCase();
        } else {
          result += char.toLowerCase();
        }
      }
    } else {
      result += char;
      if (/\s/.test(char) || char === '-' || char === '/' || char === '(' || char === '[' || char === '.') {
        capitalizeNext = true;
      }
    }
  }

  return result;
}

export function applyTextTransform(
  text: string,
  transform?: 'none' | 'uppercase' | 'capitalize' | 'lowercase'
): string {
  if (!transform || transform === 'none') return text;
  switch (transform) {
    case 'uppercase':
      return text.toUpperCase();
    case 'lowercase':
      return text.toLowerCase();
    case 'capitalize':
      return toTitleCase(text);
    default:
      return text;
  }
}
