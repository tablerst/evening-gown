const STYLE_NO_MAX_LEN = 64

// Keep this aligned with backend NormalizeStyleNo:
// - trim
// - upper
// - allow A-Z0-9 and hyphen groups (no leading/trailing hyphen, no consecutive hyphens)
const STYLE_NO_RE = /^[A-Z0-9]+(?:-[A-Z0-9]+)*$/
const DIGITS_RE = /^[0-9]+$/

export const normalizeStyleNo = (raw: unknown): string => String(raw ?? '').trim().toUpperCase()

export const isValidStyleNo = (raw: unknown): boolean => {
    const s = normalizeStyleNo(raw)
    if (!s) return false
    if (s.length > STYLE_NO_MAX_LEN) return false
    return STYLE_NO_RE.test(s)
}

const compareNumericStrings = (a: string, b: string): number => {
    // Both are digits only.
    if (a.length !== b.length) return a.length - b.length
    if (a === b) return 0
    return a < b ? -1 : 1
}

export const compareStyleNo = (aRaw: unknown, bRaw: unknown): number => {
    const a = normalizeStyleNo(aRaw)
    const b = normalizeStyleNo(bRaw)

    if (!a && !b) return 0
    if (!a) return 1
    if (!b) return -1

    const aDigits = DIGITS_RE.test(a)
    const bDigits = DIGITS_RE.test(b)
    if (aDigits && bDigits) return compareNumericStrings(a, b)

    // Keep digit-only styleNos grouped together for nicer ordering.
    if (aDigits && !bDigits) return -1
    if (!aDigits && bDigits) return 1

    return a.localeCompare(b, 'en', { sensitivity: 'base' })
}
