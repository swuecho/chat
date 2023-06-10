import { DateTime } from 'luxon';

export function nowISO(): string {
  return DateTime.now().toISO() || ''
}

export function getCurrentDate() {
  const now = DateTime.now()
  const formattedDate = now.toFormat('YYYY-MM-DD HH:mm:ss')
  return formattedDate
}

export function displayLocaleDate(ts: string) {
  // const timestampFromDb = "2023-04-16 11:10:48.000"; // assume this is the value retrieved from the database

  const dateObj = DateTime.fromISO(ts)

  const dateString = dateObj.toFormat('D t')


  return dateString
}

export function formatYearMonth(date: Date): string {
  const year = date.getFullYear().toString()
  const month = (date.getMonth() + 1).toString().padStart(2, '0')
  return `${year}-${month}`
}
