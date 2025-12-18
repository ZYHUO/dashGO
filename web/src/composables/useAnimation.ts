import { ref, computed } from 'vue'

export type AnimationType = 'fadeIn' | 'slideUp' | 'scaleIn' | 'bounce' | 'slideDown' | 'slideLeft' | 'slideRight'

interface AnimationOptions {
  duration?: number
  delay?: number
  easing?: string
  fill?: 'none' | 'forwards' | 'backwards' | 'both'
}

interface AnimationSystem {
  animate: (element: HTMLElement, animation: AnimationType, options?: AnimationOptions) => Promise<void>
  transition: (from: string, to: string, duration?: number) => void
  isAnimating: ComputedRef<boolean>
  animationCount: ComputedRef<number>
}

// Animation keyframes definitions
const animationKeyframes: Record<AnimationType, Keyframe[]> = {
  fadeIn: [
    { opacity: 0 },
    { opacity: 1 }
  ],
  slideUp: [
    { transform: 'translateY(10px)', opacity: 0 },
    { transform: 'translateY(0)', opacity: 1 }
  ],
  slideDown: [
    { transform: 'translateY(-10px)', opacity: 0 },
    { transform: 'translateY(0)', opacity: 1 }
  ],
  slideLeft: [
    { transform: 'translateX(10px)', opacity: 0 },
    { transform: 'translateX(0)', opacity: 1 }
  ],
  slideRight: [
    { transform: 'translateX(-10px)', opacity: 0 },
    { transform: 'translateX(0)', opacity: 1 }
  ],
  scaleIn: [
    { transform: 'scale(0.95)', opacity: 0 },
    { transform: 'scale(1)', opacity: 1 }
  ],
  bounce: [
    { transform: 'scale(0.3)', opacity: 0 },
    { transform: 'scale(1.05)', opacity: 0.7, offset: 0.5 },
    { transform: 'scale(0.9)', opacity: 0.9, offset: 0.7 },
    { transform: 'scale(1)', opacity: 1 }
  ]
}

// Default animation options
const defaultOptions: Required<AnimationOptions> = {
  duration: 200,
  delay: 0,
  easing: 'cubic-bezier(0.4, 0, 0.2, 1)',
  fill: 'both'
}

export function useAnimation() {
  const activeAnimations = ref(new Set<Animation>())
  
  const animationCount = computed(() => activeAnimations.value.size)
  const isAnimating = computed(() => animationCount.value > 0)
  
  // Check if reduced motion is preferred
  const prefersReducedMotion = () => {
    return window.matchMedia('(prefers-reduced-motion: reduce)').matches
  }
  
  const animate = async (
    element: HTMLElement,
    animationType: AnimationType,
    options: AnimationOptions = {}
  ): Promise<void> => {
    // Respect user's motion preferences
    if (prefersReducedMotion()) {
      return Promise.resolve()
    }
    
    const finalOptions = { ...defaultOptions, ...options }
    const keyframes = animationKeyframes[animationType]
    
    if (!keyframes) {
      console.warn(`Animation type "${animationType}" not found`)
      return Promise.resolve()
    }
    
    return new Promise((resolve, reject) => {
      try {
        const animation = element.animate(keyframes, {
          duration: finalOptions.duration,
          delay: finalOptions.delay,
          easing: finalOptions.easing,
          fill: finalOptions.fill
        })
        
        // Track active animation
        activeAnimations.value.add(animation)
        
        animation.addEventListener('finish', () => {
          activeAnimations.value.delete(animation)
          resolve()
        })
        
        animation.addEventListener('cancel', () => {
          activeAnimations.value.delete(animation)
          resolve()
        })
        
        animation.addEventListener('error', (error) => {
          activeAnimations.value.delete(animation)
          reject(error)
        })
        
      } catch (error) {
        reject(error)
      }
    })
  }
  
  const transition = (
    from: string,
    to: string,
    duration: number = 200
  ): void => {
    if (prefersReducedMotion()) {
      return
    }
    
    // This is a placeholder for CSS transition management
    // In a real implementation, this would manage CSS transitions
    console.log(`Transitioning from ${from} to ${to} over ${duration}ms`)
  }
  
  // Cancel all active animations
  const cancelAll = (): void => {
    activeAnimations.value.forEach(animation => {
      animation.cancel()
    })
    activeAnimations.value.clear()
  }
  
  // Animate multiple elements in sequence
  const animateSequence = async (
    elements: HTMLElement[],
    animationType: AnimationType,
    options: AnimationOptions = {},
    stagger: number = 50
  ): Promise<void> => {
    const promises = elements.map((element, index) => {
      const delayedOptions = {
        ...options,
        delay: (options.delay || 0) + (index * stagger)
      }
      return animate(element, animationType, delayedOptions)
    })
    
    await Promise.all(promises)
  }
  
  // Animate elements in parallel
  const animateParallel = async (
    elements: HTMLElement[],
    animationType: AnimationType,
    options: AnimationOptions = {}
  ): Promise<void> => {
    const promises = elements.map(element => animate(element, animationType, options))
    await Promise.all(promises)
  }
  
  return {
    animate,
    transition,
    isAnimating,
    animationCount,
    cancelAll,
    animateSequence,
    animateParallel,
    prefersReducedMotion,
  }
}

// Utility function for one-off animations
export async function animateElement(
  element: HTMLElement,
  animationType: AnimationType,
  options?: AnimationOptions
): Promise<void> {
  const { animate } = useAnimation()
  return animate(element, animationType, options)
}

// CSS class-based animation helpers
export function addAnimationClass(element: HTMLElement, animationType: AnimationType): void {
  element.classList.add(`animate-${animationType}`)
  
  // Remove class after animation completes
  const handleAnimationEnd = () => {
    element.classList.remove(`animate-${animationType}`)
    element.removeEventListener('animationend', handleAnimationEnd)
  }
  
  element.addEventListener('animationend', handleAnimationEnd)
}