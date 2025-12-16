import { describe, it, expect } from 'vitest'
import { cn, formatBytes } from './utils'

describe('utils', () => {
  describe('cn', () => {
    it('merges class names correctly', () => {
      const result = cn('foo', 'bar')
      expect(result).toBe('foo bar')
    })

    it('handles conditional classes', () => {
      const result = cn('foo', false && 'bar', 'baz')
      expect(result).toBe('foo baz')
    })

    it('merges tailwind classes correctly', () => {
      const result = cn('px-2 py-1', 'px-3')
      expect(result).toBe('py-1 px-3')
    })
  })

  describe('formatBytes', () => {
    it('formats 0 bytes', () => {
      expect(formatBytes(0)).toBe('0 Bytes')
    })

    it('formats bytes correctly', () => {
      expect(formatBytes(500)).toBe('500 Bytes')
    })

    it('formats KB correctly', () => {
      expect(formatBytes(1024)).toBe('1 KB')
      expect(formatBytes(2048)).toBe('2 KB')
    })

    it('formats MB correctly', () => {
      expect(formatBytes(1024 * 1024)).toBe('1 MB')
      expect(formatBytes(1024 * 1024 * 5)).toBe('5 MB')
    })

    it('formats GB correctly', () => {
      expect(formatBytes(1024 * 1024 * 1024)).toBe('1 GB')
      expect(formatBytes(1024 * 1024 * 1024 * 2.5)).toBe('2.5 GB')
    })

    it('formats TB correctly', () => {
      expect(formatBytes(1024 * 1024 * 1024 * 1024)).toBe('1 TB')
    })

    it('respects decimals parameter', () => {
      expect(formatBytes(1536, 0)).toBe('2 KB')
      expect(formatBytes(1536, 1)).toBe('1.5 KB')
      expect(formatBytes(1536, 3)).toBe('1.5 KB')
    })

    it('handles large numbers', () => {
      const largeNumber = 1024 * 1024 * 1024 * 500 // 500 GB
      const result = formatBytes(largeNumber)
      expect(result).toContain('GB')
      expect(result).toContain('500')
    })
  })
})
