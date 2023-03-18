import moment from 'moment'

export function getCurrentDate() {
  const now = moment()
  const formattedDate = now.format('YYYY-MM-DD HH:mm:ss')
  return formattedDate
}
