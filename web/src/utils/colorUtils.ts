// Color utility functions for accessibility and design system

/**
 * Convert hex color to RGB values
 */
export function hexToRgb(hex: string): { r: number; g: number; b: number } | null {
  const result = /^#?([a-f\d]{2})([a-f\d]{2})([a-f\d]{2})$/i.exec(hex)
  return result ? {
    r: parseInt(result[1], 16),
    g: parseInt(result[2], 16),
    b: parseInt(result[3], 16)
  } : null
}

/**
 * Calculate relative luminance of a color
 * Based on WCAG 2.1 guidelines
 */
export function getLuminance(r: number, g: number, b: number): number {
  const [rs, gs, bs] = [r, g, b].map(c => {
    c = c / 255
    return c <= 0.03928 ? c / 12.92 : Math.pow((c + 0.055) / 1.055, 2.4)
  })
  return 0.2126 * rs + 0.7152 * gs + 0.0722 * bs
}

/**
 * Calculate contrast ratio between two colors
 * Returns a value between 1 and 21
 */
export function getContrastRatio(color1: string, color2: string): number {
  const rgb1 = hexToRgb(color1)
  const rgb2 = hexToRgb(color2)
  
  if (!rgb1 || !rgb2) return 1
  
  const lum1 = getLuminance(rgb1.r, rgb1.g, rgb1.b)
  const lum2 = getLuminance(rgb2.r, rgb2.g, rgb2.b)
  
  const brightest = Math.max(lum1, lum2)
  const darkest = Math.min(lum1, lum2)
  
  return (brightest + 0.05) / (darkest + 0.05)
}

/**
 * Check if color combination meets WCAG AA standards
 */
export function meetsWCAGAA(foreground: string, background: string, isLargeText = false): boolean {
  const ratio = getContrastRatio(foreground, background)
  return isLargeText ? ratio >= 3 : ratio >= 4.5
}

/**
 * Check if color combination meets WCAG AAA standards
 */
export function meetsWCAGAAA(foreground: string, background: string, isLargeText = false): boolean {
  const ratio = getContrastRatio(foreground, background)
  return isLargeText ? ratio >= 4.5 : ratio >= 7
}

/**
 * Get the appropriate text color (light or dark) for a background
 */
export function getTextColorForBackground(backgroundColor: string): 'light' | 'dark' {
  const rgb = hexToRgb(backgroundColor)
  if (!rgb) return 'dark'
  
  const luminance = getLuminance(rgb.r, rgb.g, rgb.b)
  return luminance > 0.5 ? 'dark' : 'light'
}

/**
 * Generate a color palette with proper contrast ratios
 */
export function generateAccessiblePalette(baseColor: string): {
  50: string
  100: string
  200: string
  300: string
  400: string
  500: string
  600: string
  700: string
  800: string
  900: string
  950: string
} {
  // This is a simplified implementation
  // In a real application, you'd use a more sophisticated color generation algorithm
  const rgb = hexToRgb(baseColor)
  if (!rgb) throw new Error('Invalid base color')
  
  const { r, g, b } = rgb
  
  return {
    50: `rgb(${Math.min(255, r + 200)}, ${Math.min(255, g + 200)}, ${Math.min(255, b + 200)})`,
    100: `rgb(${Math.min(255, r + 150)}, ${Math.min(255, g + 150)}, ${Math.min(255, b + 150)})`,
    200: `rgb(${Math.min(255, r + 100)}, ${Math.min(255, g + 100)}, ${Math.min(255, b + 100)})`,
    300: `rgb(${Math.min(255, r + 50)}, ${Math.min(255, g + 50)}, ${Math.min(255, b + 50)})`,
    400: `rgb(${Math.min(255, r + 25)}, ${Math.min(255, g + 25)}, ${Math.min(255, b + 25)})`,
    500: baseColor,
    600: `rgb(${Math.max(0, r - 25)}, ${Math.max(0, g - 25)}, ${Math.max(0, b - 25)})`,
    700: `rgb(${Math.max(0, r - 50)}, ${Math.max(0, g - 50)}, ${Math.max(0, b - 50)})`,
    800: `rgb(${Math.max(0, r - 100)}, ${Math.max(0, g - 100)}, ${Math.max(0, b - 100)})`,
    900: `rgb(${Math.max(0, r - 150)}, ${Math.max(0, g - 150)}, ${Math.max(0, b - 150)})`,
    950: `rgb(${Math.max(0, r - 200)}, ${Math.max(0, g - 200)}, ${Math.max(0, b - 200)})`,
  }
}

/**
 * Validate that a color is not pure black (#000000)
 */
export function isPureBlack(color: string): boolean {
  const normalized = color.toLowerCase().replace('#', '')
  return normalized === '000000' || normalized === '000'
}

/**
 * Check if a color is too dark for large areas (to avoid pure black backgrounds)
 */
export function isTooBlackForLargeAreas(color: string): boolean {
  const rgb = hexToRgb(color)
  if (!rgb) return false
  
  const luminance = getLuminance(rgb.r, rgb.g, rgb.b)
  return luminance < 0.05 // Very dark colors
}

/**
 * Get semantic color variants for different states
 */
export function getSemanticColorVariants(baseColor: string) {
  const rgb = hexToRgb(baseColor)
  if (!rgb) throw new Error('Invalid base color')
  
  return {
    default: baseColor,
    hover: `rgb(${Math.max(0, rgb.r - 20)}, ${Math.max(0, rgb.g - 20)}, ${Math.max(0, rgb.b - 20)})`,
    active: `rgb(${Math.max(0, rgb.r - 40)}, ${Math.max(0, rgb.g - 40)}, ${Math.max(0, rgb.b - 40)})`,
    disabled: `rgb(${Math.min(255, rgb.r + 100)}, ${Math.min(255, rgb.g + 100)}, ${Math.min(255, rgb.b + 100)})`,
  }
}

/**
 * Create gradient CSS string
 */
export function createGradient(
  color1: string,
  color2: string,
  direction: string = '135deg'
): string {
  return `linear-gradient(${direction}, ${color1} 0%, ${color2} 100%)`
}

/**
 * Create color with opacity
 */
export function withOpacity(color: string, opacity: number): string {
  const rgb = hexToRgb(color)
  if (!rgb) return color
  
  return `rgba(${rgb.r}, ${rgb.g}, ${rgb.b}, ${opacity})`
}