import { DateTime } from 'luxon'

export function nowISO(): string {
  return DateTime.now().toISO() || ''
}

export function getCurrentDate() {
  const now = DateTime.now()
  const formattedDate = now.toFormat('yyyy-MM-dd-HHmm-ss')
  return formattedDate
}

// 2025-03-05T12:48:11.990824Z 2025-02-26T08:58:48Z
export function displayLocaleDate(ts: string) {
  const dateObj = DateTime.fromISO(ts)

  const dateString = dateObj.toFormat('D t')

  return dateString
}

// Relative "last used" label for a model timestamp, e.g. "3 days ago".
export function formatLastUsed(timestamp?: string, now: DateTime = DateTime.now()): string {
  if (!timestamp)
    return 'Never used'

  const date = DateTime.fromISO(timestamp)
  // The backend sends Go's zero time (0001-01-01T00:00:00Z) for unused models.
  if (!date.isValid || date.toUTC().year <= 1)
    return 'Never used'

  if (now.diff(date, 'days').days > 30)
    return 'a month ago'

  return date.toRelative({ base: now }) ?? 'Never used'
}

export function formatYearMonth(date: Date): string {
  const year = date.getFullYear().toString()
  const month = (date.getMonth() + 1).toString().padStart(2, '0')
  return `${year}-${month}`
}
