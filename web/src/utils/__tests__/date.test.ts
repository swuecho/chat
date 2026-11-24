import { beforeAll, describe, expect, it } from 'vitest'
import { DateTime } from 'luxon'
import { displayLocaleDate, formatLastUsed } from '../date'

describe('displayLocaleDate', () => {
  beforeAll(() => {
    // Set a fixed timezone for all tests
    process.env.TZ = 'Asia/Shanghai'
  })

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

describe('formatLastUsed', () => {
  const now = DateTime.fromISO('2025-03-10T12:00:00Z')

  it('returns "Never used" for missing, invalid, and Go zero timestamps', () => {
    expect(formatLastUsed(undefined, now)).toBe('Never used')
    expect(formatLastUsed('not-a-date', now)).toBe('Never used')
    expect(formatLastUsed('0001-01-01T00:00:00Z', now)).toBe('Never used')
  })

  it('formats recent timestamps relative to now', () => {
    expect(formatLastUsed('2025-03-07T12:00:00Z', now)).toBe('3 days ago')
  })

  it('caps old timestamps at "a month ago"', () => {
    expect(formatLastUsed('2024-12-01T00:00:00Z', now)).toBe('a month ago')
  })
})
