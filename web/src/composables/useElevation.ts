import { ref, computed, watch } from 'vue'

export type ElevationLevel = 0 | 1 | 2 | 3 | 4 | 5 | 6

export interface ElevationSystem {
  elevation: ComputedRef<number>
  setElevation: (level: ElevationLevel) => void
  elevationClasses: ComputedRef<string[]>
  shadowClass: ComputedRef<string>
  isElevated: ComputedRef<boolean>
  zIndex: ComputedRef<number>
  elevate: () => void
  lower: () => void
  reset: () => void
}

export interface ElevationTransition {
  from: ElevationLevel
  to: ElevationLevel
  duration: number
  easing: string
}

// Elevation level mappings
const elevationShadows: Record<ElevationLevel, string> = {
  0: 'shadow-none',
  1: 'shadow-xs',
  2: 'shadow-sm',
  3: 'shadow-md',
  4: 'shadow-lg',
  5: 'shadow-xl',
  6: 'shadow-2xl',
}

const elevationZIndex: Record<ElevationLevel, number> = {
  0: 0,
  1: 10,
  2: 20,
  3: 30,
  4: 40,
  5: 50,
  6: 60,
}

export function useElevation(initialLevel: ElevationLevel = 0, options?: {
  interactive?: boolean
  hoverElevation?: ElevationLevel
  activeElevation?: ElevationLevel
}) {
  const currentElevation = ref<ElevationLevel>(initialLevel)
  const isHovered = ref(false)
  const isActive = ref(false)
  const isTransitioning = ref(false)
  
  const { interactive = false, hoverElevation, activeElevation } = options || {}
  
  const elevation = computed(() => currentElevation.value)
  
  // Calculate effective elevation based on interaction state
  const effectiveElevation = computed(() => {
    if (!interactive) return currentElevation.value
    
    if (isActive.value && activeElevation !== undefined) {
      return activeElevation
    }
    
    if (isHovered.value && hoverElevation !== undefined) {
      return hoverElevation
    }
    
    return currentElevation.value
  })
  
  const setElevation = (level: ElevationLevel, animate = true) => {
    if (animate && level !== currentElevation.value) {
      isTransitioning.value = true
      setTimeout(() => {
        isTransitioning.value = false
      }, 200)
    }
    currentElevation.value = level
  }
  
  const shadowClass = computed(() => elevationShadows[effectiveElevation.value])
  
  const elevationClasses = computed(() => {
    const classes = [`elevation-${effectiveElevation.value}`]
    
    if (effectiveElevation.value > 0) {
      classes.push(shadowClass.value)
      classes.push('relative')
    }
    
    if (interactive) {
      classes.push('elevation-interactive')
    }
    
    if (isTransitioning.value) {
      classes.push('elevation-transitioning')
    }
    
    return classes
  })
  
  const isElevated = computed(() => effectiveElevation.value > 0)
  
  const zIndex = computed(() => elevationZIndex[effectiveElevation.value])
  
  // Helper methods for common elevation changes
  const elevate = () => {
    if (currentElevation.value < 6) {
      currentElevation.value = (currentElevation.value + 1) as ElevationLevel
    }
  }
  
  const lower = () => {
    if (currentElevation.value > 0) {
      currentElevation.value = (currentElevation.value - 1) as ElevationLevel
    }
  }
  
  const reset = () => {
    currentElevation.value = 0
  }
  
  // Interactive state management
  const setHovered = (hovered: boolean) => {
    isHovered.value = hovered
  }
  
  const setActive = (active: boolean) => {
    isActive.value = active
  }
  
  // Animation helpers
  const animateToElevation = async (targetLevel: ElevationLevel, duration = 200) => {
    return new Promise<void>((resolve) => {
      isTransitioning.value = true
      currentElevation.value = targetLevel
      
      setTimeout(() => {
        isTransitioning.value = false
        resolve()
      }, duration)
    })
  }
  
  // Preset elevation changes
  const floatUp = () => animateToElevation(Math.min(6, currentElevation.value + 2) as ElevationLevel)
  const settle = () => animateToElevation(Math.max(0, currentElevation.value - 1) as ElevationLevel)
  
  return {
    elevation,
    effectiveElevation,
    setElevation,
    elevationClasses,
    shadowClass,
    isElevated,
    zIndex,
    elevate,
    lower,
    reset,
    setHovered,
    setActive,
    animateToElevation,
    floatUp,
    settle,
    isTransitioning: computed(() => isTransitioning.value),
    isHovered: computed(() => isHovered.value),
    isActive: computed(() => isActive.value),
  }
}

// Predefined elevation presets for common components
export const elevationPresets = {
  flat: 0 as ElevationLevel,
  raised: 1 as ElevationLevel,
  floating: 2 as ElevationLevel,
  overlay: 3 as ElevationLevel,
  modal: 4 as ElevationLevel,
  popover: 5 as ElevationLevel,
  tooltip: 6 as ElevationLevel,
} as const

// Helper function to get elevation classes without composable
export function getElevationClasses(level: ElevationLevel): string[] {
  const classes = [`elevation-${level}`]
  
  if (level > 0) {
    classes.push(elevationShadows[level])
    classes.push('relative')
  }
  
  return classes
}
// Advanced elevation utilities
export function createElevationGroup(elements: { level: ElevationLevel; interactive?: boolean }[]) {
  const elevationInstances = elements.map(({ level, interactive }) => 
    useElevation(level, { interactive })
  )
  
  const setGroupElevation = (level: ElevationLevel) => {
    elevationInstances.forEach(instance => instance.setElevation(level))
  }
  
  const elevateGroup = () => {
    elevationInstances.forEach(instance => instance.elevate())
  }
  
  const lowerGroup = () => {
    elevationInstances.forEach(instance => instance.lower())
  }
  
  const resetGroup = () => {
    elevationInstances.forEach(instance => instance.reset())
  }
  
  return {
    instances: elevationInstances,
    setGroupElevation,
    elevateGroup,
    lowerGroup,
    resetGroup,
  }
}

// Elevation context for nested components
export function useElevationContext(baseLevel: ElevationLevel = 0) {
  const contextLevel = ref(baseLevel)
  
  const getChildElevation = (childLevel: ElevationLevel): ElevationLevel => {
    const combined = contextLevel.value + childLevel
    return Math.min(6, combined) as ElevationLevel
  }
  
  const setContextLevel = (level: ElevationLevel) => {
    contextLevel.value = level
  }
  
  return {
    contextLevel: computed(() => contextLevel.value),
    getChildElevation,
    setContextLevel,
  }
}

// Elevation-aware component wrapper
export function withElevation<T extends Record<string, any>>(
  component: T,
  defaultElevation: ElevationLevel = 0
) {
  return {
    ...component,
    props: {
      ...component.props,
      elevation: {
        type: Number,
        default: defaultElevation,
        validator: (value: number) => value >= 0 && value <= 6 && Number.isInteger(value)
      },
      interactive: {
        type: Boolean,
        default: false
      },
      hoverElevation: {
        type: Number,
        validator: (value: number) => value >= 0 && value <= 6 && Number.isInteger(value)
      },
      activeElevation: {
        type: Number,
        validator: (value: number) => value >= 0 && value <= 6 && Number.isInteger(value)
      }
    }
  }
}

// Elevation transition utilities
export function createElevationTransition(
  from: ElevationLevel,
  to: ElevationLevel,
  options?: {
    duration?: number
    easing?: string
    delay?: number
  }
): ElevationTransition {
  const { duration = 200, easing = 'cubic-bezier(0.4, 0, 0.2, 1)', delay = 0 } = options || {}
  
  return {
    from,
    to,
    duration: duration + delay,
    easing
  }
}

// Elevation performance monitoring
export function useElevationPerformance() {
  const elevationChanges = ref(0)
  const lastChangeTime = ref(0)
  
  const trackElevationChange = () => {
    elevationChanges.value++
    lastChangeTime.value = performance.now()
  }
  
  const getPerformanceMetrics = () => ({
    totalChanges: elevationChanges.value,
    lastChange: lastChangeTime.value,
    averageFrequency: elevationChanges.value / (performance.now() / 1000)
  })
  
  const resetMetrics = () => {
    elevationChanges.value = 0
    lastChangeTime.value = 0
  }
  
  return {
    trackElevationChange,
    getPerformanceMetrics,
    resetMetrics,
    metrics: computed(() => getPerformanceMetrics())
  }
}