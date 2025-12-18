import { ref, computed, watch, onMounted } from 'vue'

export type ThemeMode = 'light' | 'dark' | 'auto'

const THEME_STORAGE_KEY = 'dashgo-theme'

// Theme state
const currentTheme = ref<ThemeMode>('auto')
const systemTheme = ref<'light' | 'dark'>('light')

// Get system theme preference
const getSystemTheme = (): 'light' | 'dark' => {
  if (typeof window !== 'undefined' && window.matchMedia) {
    return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
  }
  return 'light'
}

// Get effective theme (resolves 'auto' to actual theme)
const getEffectiveTheme = (theme: ThemeMode): 'light' | 'dark' => {
  if (theme === 'auto') {
    return systemTheme.value
  }
  return theme
}

// Initialize theme from localStorage or system preference
const initTheme = () => {
  // Get system theme
  systemTheme.value = getSystemTheme()
  
  // Get stored theme preference
  const stored = localStorage.getItem(THEME_STORAGE_KEY) as ThemeMode | null
  
  if (stored && ['light', 'dark', 'auto'].includes(stored)) {
    currentTheme.value = stored
  } else {
    currentTheme.value = 'auto'
  }
  
  applyTheme(currentTheme.value)
}

// Apply theme to document with smooth transition
const applyTheme = (theme: ThemeMode) => {
  const effectiveTheme = getEffectiveTheme(theme)
  
  // Add transition class for smooth theme switching
  document.documentElement.classList.add('theme-transitioning')
  
  // Apply theme attribute
  document.documentElement.setAttribute('data-theme', theme)
  
  // Update body class for additional styling hooks
  document.body.className = document.body.className.replace(/theme-\w+/g, '')
  document.body.classList.add(`theme-${effectiveTheme}`)
  
  // Update meta theme-color for mobile browsers
  const metaThemeColor = document.querySelector('meta[name="theme-color"]')
  if (metaThemeColor) {
    const themeColors = {
      light: '#FAFAF9',
      dark: '#0C0A09'
    }
    metaThemeColor.setAttribute('content', themeColors[effectiveTheme])
  }
  
  // Update CSS custom properties for immediate theme application
  const root = document.documentElement
  if (effectiveTheme === 'dark') {
    root.style.setProperty('--theme-transition-duration', '200ms')
  } else {
    root.style.setProperty('--theme-transition-duration', '200ms')
  }
  
  // Dispatch theme change event for other components to listen
  const themeChangeEvent = new CustomEvent('themechange', {
    detail: { theme, effectiveTheme }
  })
  document.dispatchEvent(themeChangeEvent)
  
  // Remove transition class after animation completes
  setTimeout(() => {
    document.documentElement.classList.remove('theme-transitioning')
  }, 200)
}

// Listen for system theme changes
const setupSystemThemeListener = () => {
  if (typeof window !== 'undefined' && window.matchMedia) {
    const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)')
    
    const handleSystemThemeChange = (e: MediaQueryListEvent) => {
      systemTheme.value = e.matches ? 'dark' : 'light'
      
      // If current theme is auto, reapply to reflect system change
      if (currentTheme.value === 'auto') {
        applyTheme(currentTheme.value)
      }
    }
    
    // Modern browsers
    if (mediaQuery.addEventListener) {
      mediaQuery.addEventListener('change', handleSystemThemeChange)
    } else {
      // Legacy browsers
      mediaQuery.addListener(handleSystemThemeChange)
    }
    
    return () => {
      if (mediaQuery.removeEventListener) {
        mediaQuery.removeEventListener('change', handleSystemThemeChange)
      } else {
        mediaQuery.removeListener(handleSystemThemeChange)
      }
    }
  }
}

// Watch for theme changes
watch(currentTheme, (newTheme) => {
  localStorage.setItem(THEME_STORAGE_KEY, newTheme)
  applyTheme(newTheme)
})

export function useTheme() {
  // Initialize on first use
  if (!document.documentElement.hasAttribute('data-theme')) {
    initTheme()
  }
  
  // Setup system theme listener
  onMounted(() => {
    setupSystemThemeListener()
  })
  
  const theme = computed(() => currentTheme.value)
  const effectiveTheme = computed(() => getEffectiveTheme(currentTheme.value))
  
  const setTheme = (newTheme: ThemeMode) => {
    currentTheme.value = newTheme
  }
  
  const toggleTheme = () => {
    const themes: ThemeMode[] = ['light', 'dark', 'auto']
    const currentIndex = themes.indexOf(currentTheme.value)
    const nextIndex = (currentIndex + 1) % themes.length
    currentTheme.value = themes[nextIndex]
  }
  
  // Quick toggle between light and dark (ignoring auto)
  const toggleLightDark = () => {
    if (effectiveTheme.value === 'light') {
      currentTheme.value = 'dark'
    } else {
      currentTheme.value = 'light'
    }
  }
  
  const isLight = computed(() => effectiveTheme.value === 'light')
  const isDark = computed(() => effectiveTheme.value === 'dark')
  const isAuto = computed(() => currentTheme.value === 'auto')
  
  // Theme preference helpers
  const getThemeIcon = () => {
    switch (currentTheme.value) {
      case 'light': return '☀️'
      case 'dark': return '🌙'
      case 'auto': return '🔄'
      default: return '☀️'
    }
  }
  
  const getThemeLabel = () => {
    switch (currentTheme.value) {
      case 'light': return 'Light'
      case 'dark': return 'Dark'
      case 'auto': return 'Auto'
      default: return 'Light'
    }
  }
  
  // Advanced theme utilities
  const getThemeAwareValue = <T>(lightValue: T, darkValue: T): T => {
    return effectiveTheme.value === 'light' ? lightValue : darkValue
  }
  
  const watchThemeChange = (callback: (theme: ThemeMode, effectiveTheme: 'light' | 'dark') => void) => {
    const handleThemeChange = (event: CustomEvent) => {
      callback(event.detail.theme, event.detail.effectiveTheme)
    }
    
    document.addEventListener('themechange', handleThemeChange as EventListener)
    
    // Return cleanup function
    return () => {
      document.removeEventListener('themechange', handleThemeChange as EventListener)
    }
  }
  
  // Persistence utilities
  const exportThemePreference = () => {
    return {
      theme: currentTheme.value,
      timestamp: Date.now(),
      systemTheme: systemTheme.value
    }
  }
  
  const importThemePreference = (preference: { theme: ThemeMode }) => {
    if (['light', 'dark', 'auto'].includes(preference.theme)) {
      setTheme(preference.theme)
    }
  }
  
  // Reset to system default
  const resetToSystem = () => {
    setTheme('auto')
  }
  
  return {
    // Core state
    theme,
    effectiveTheme,
    systemTheme: computed(() => systemTheme.value),
    
    // Theme controls
    setTheme,
    toggleTheme,
    toggleLightDark,
    resetToSystem,
    
    // State checks
    isLight,
    isDark,
    isAuto,
    
    // UI helpers
    getThemeIcon,
    getThemeLabel,
    
    // Advanced utilities
    getThemeAwareValue,
    watchThemeChange,
    
    // Persistence
    exportThemePreference,
    importThemePreference,
  }
}
