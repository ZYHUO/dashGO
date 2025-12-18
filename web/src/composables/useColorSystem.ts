import { computed, ref } from 'vue'
import { 
  getContrastRatio, 
  meetsWCAGAA, 
  getTextColorForBackground,
  isPureBlack,
  isTooBlackForLargeAreas,
  withOpacity,
  createGradient
} from '../utils/colorUtils'

export interface ColorScale {
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
}

export interface SemanticColors {
  success: ColorScale
  warning: ColorScale
  error: ColorScale
  info: ColorScale
}

export interface ColorSystemState {
  primary: ColorScale
  secondary: ColorScale
  neutral: ColorScale
  semantic: SemanticColors
}

// Default color system (matches our design tokens)
const defaultColorSystem: ColorSystemState = {
  primary: {
    50: '#EFF6FF',
    100: '#DBEAFE',
    200: '#BFDBFE',
    300: '#93C5FD',
    400: '#60A5FA',
    500: '#3B82F6',
    600: '#2563EB',
    700: '#1D4ED8',
    800: '#1E40AF',
    900: '#1E3A8A',
    950: '#172554',
  },
  secondary: {
    50: '#FAF5FF',
    100: '#F3E8FF',
    200: '#E9D5FF',
    300: '#D8B4FE',
    400: '#C084FC',
    500: '#A855F7',
    600: '#9333EA',
    700: '#7C3AED',
    800: '#6B21A8',
    900: '#581C87',
    950: '#3B0764',
  },
  neutral: {
    50: '#FAFAF9',
    100: '#F5F5F4',
    200: '#E7E5E4',
    300: '#D6D3D1',
    400: '#A8A29E',
    500: '#78716C',
    600: '#57534E',
    700: '#44403C',
    800: '#292524',
    900: '#1C1917',
    950: '#0C0A09',
  },
  semantic: {
    success: {
      50: '#F0FDF4',
      100: '#DCFCE7',
      200: '#BBF7D0',
      300: '#86EFAC',
      400: '#4ADE80',
      500: '#22C55E',
      600: '#16A34A',
      700: '#15803D',
      800: '#166534',
      900: '#14532D',
      950: '#052e16',
    },
    warning: {
      50: '#FFFBEB',
      100: '#FEF3C7',
      200: '#FDE68A',
      300: '#FCD34D',
      400: '#FBBF24',
      500: '#F59E0B',
      600: '#D97706',
      700: '#B45309',
      800: '#92400E',
      900: '#78350F',
      950: '#451a03',
    },
    error: {
      50: '#FEF2F2',
      100: '#FEE2E2',
      200: '#FECACA',
      300: '#FCA5A5',
      400: '#F87171',
      500: '#EF4444',
      600: '#DC2626',
      700: '#B91C1C',
      800: '#991B1B',
      900: '#7F1D1D',
      950: '#450a0a',
    },
    info: {
      50: '#F0F9FF',
      100: '#E0F2FE',
      200: '#BAE6FD',
      300: '#7DD3FC',
      400: '#38BDF8',
      500: '#0EA5E9',
      600: '#0284C7',
      700: '#0369A1',
      800: '#075985',
      900: '#0C4A6E',
      950: '#082f49',
    },
  },
}

export function useColorSystem() {
  const colorSystem = ref<ColorSystemState>(defaultColorSystem)
  
  // Computed properties for easy access
  const primary = computed(() => colorSystem.value.primary)
  const secondary = computed(() => colorSystem.value.secondary)
  const neutral = computed(() => colorSystem.value.neutral)
  const semantic = computed(() => colorSystem.value.semantic)
  
  // Utility functions
  const getColor = (scale: ColorScale, shade: keyof ColorScale) => {
    return scale[shade]
  }
  
  const getPrimaryColor = (shade: keyof ColorScale = 500) => {
    return getColor(primary.value, shade)
  }
  
  const getSecondaryColor = (shade: keyof ColorScale = 500) => {
    return getColor(secondary.value, shade)
  }
  
  const getNeutralColor = (shade: keyof ColorScale = 500) => {
    return getColor(neutral.value, shade)
  }
  
  const getSemanticColor = (type: keyof SemanticColors, shade: keyof ColorScale = 500) => {
    return getColor(semantic.value[type], shade)
  }
  
  // Accessibility helpers
  const checkContrast = (foreground: string, background: string) => {
    const ratio = getContrastRatio(foreground, background)
    return {
      ratio,
      meetsAA: meetsWCAGAA(foreground, background),
      meetsAALarge: meetsWCAGAA(foreground, background, true),
    }
  }
  
  const getAccessibleTextColor = (backgroundColor: string) => {
    const textType = getTextColorForBackground(backgroundColor)
    return textType === 'light' ? neutral.value[50] : neutral.value[900]
  }
  
  // Validation functions
  const validateColorSystem = () => {
    const issues: string[] = []
    
    // Check for pure black usage
    Object.entries(colorSystem.value).forEach(([scaleName, scale]) => {
      if (typeof scale === 'object' && 'value' in scale) return // Skip computed refs
      
      Object.entries(scale as ColorScale | SemanticColors).forEach(([shadeName, color]) => {
        if (typeof color === 'string') {
          if (isPureBlack(color)) {
            issues.push(`Pure black detected in ${scaleName}.${shadeName}`)
          }
          if (isTooBlackForLargeAreas(color) && shadeName !== '950') {
            issues.push(`Very dark color in ${scaleName}.${shadeName} may not be suitable for large areas`)
          }
        } else if (typeof color === 'object') {
          // Handle semantic colors
          Object.entries(color).forEach(([subShade, subColor]) => {
            if (isPureBlack(subColor)) {
              issues.push(`Pure black detected in ${scaleName}.${shadeName}.${subShade}`)
            }
          })
        }
      })
    })
    
    return {
      isValid: issues.length === 0,
      issues,
    }
  }
  
  // Generate color variants
  const generateColorVariants = (baseColor: string) => {
    return {
      default: baseColor,
      hover: withOpacity(baseColor, 0.8),
      active: withOpacity(baseColor, 0.9),
      disabled: withOpacity(baseColor, 0.5),
      subtle: withOpacity(baseColor, 0.1),
      muted: withOpacity(baseColor, 0.6),
    }
  }
  
  // Create gradients
  const createPrimaryGradient = (direction = '135deg') => {
    return createGradient(primary.value[500], primary.value[600], direction)
  }
  
  const createSecondaryGradient = (direction = '135deg') => {
    return createGradient(secondary.value[500], secondary.value[600], direction)
  }
  
  const createSurfaceGradient = (direction = '135deg') => {
    return createGradient(neutral.value[50], neutral.value[100], direction)
  }
  
  // Get color for current theme
  const getThemeAwareColor = (lightColor: string, darkColor: string) => {
    // This would integrate with the theme system
    // For now, return light color as default
    return lightColor
  }
  
  // Update color system
  const updateColorSystem = (newColorSystem: Partial<ColorSystemState>) => {
    colorSystem.value = { ...colorSystem.value, ...newColorSystem }
  }
  
  return {
    // State
    colorSystem: computed(() => colorSystem.value),
    primary,
    secondary,
    neutral,
    semantic,
    
    // Getters
    getColor,
    getPrimaryColor,
    getSecondaryColor,
    getNeutralColor,
    getSemanticColor,
    
    // Accessibility
    checkContrast,
    getAccessibleTextColor,
    
    // Validation
    validateColorSystem,
    
    // Utilities
    generateColorVariants,
    createPrimaryGradient,
    createSecondaryGradient,
    createSurfaceGradient,
    getThemeAwareColor,
    
    // Updates
    updateColorSystem,
  }
}