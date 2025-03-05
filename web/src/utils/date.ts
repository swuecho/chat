import { DateTime } from 'luxon';

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

export function formatYearMonth(date: Date): string {
  const year = date.getFullYear().toString()
  const month = (date.getMonth() + 1).toString().padStart(2, '0')
  return `${year}-${month}`
}
