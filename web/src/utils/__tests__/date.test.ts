import { expect, describe, it } from 'vitest'
import { displayLocaleDate } from '../date'


describe('displayLocaleDate', () => {
  it('should format ISO date string to local date and time', () => {
    const isoDate = '2025-03-05T12:48:11.990824Z'
    const result = displayLocaleDate(isoDate)
    expect(result).toBe('3/5/2025 8:48 PM')
  })

  it('should handle date without milliseconds', () => {
    const isoDate = '2025-02-26T08:58:48Z'
    const result = displayLocaleDate(isoDate)
    expect(result).toBe('2/26/2025 4:58 PM')
  })

  it('should handle invalid date string', () => {
    const invalidDate = 'invalid-date'
    const result = displayLocaleDate(invalidDate)
    expect(result).toBe('Invalid DateTime')
  })
})
