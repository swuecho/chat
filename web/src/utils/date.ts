import moment from 'moment'

export function getCurrentDate() {
  const now = moment()
  const formattedDate = now.format('YYYY-MM-DD HH:mm:ss')
  return formattedDate
}

export function displayLocaleDate(ts: string) {
  // const timestampFromDb = "2023-04-16 11:10:48.000"; // assume this is the value retrieved from the database
  const dateObj = moment(ts)

  const dateString = `${dateObj.format('LLL')}`

  return dateString
}

export function formatYearMonth(date: Date): string {
  const year = date.getFullYear().toString()
  const month = (date.getMonth() + 1).toString().padStart(2, '0')
  return `${year}-${month}`
}
