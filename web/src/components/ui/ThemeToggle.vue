<template>
  <div class="theme-toggle">
    <button
      :class="buttonClasses"
      @click="handleToggle"
      :aria-label="`Switch to ${getNextThemeLabel()} theme`"
      :title="`Current: ${getThemeLabel()}, Click to switch to ${getNextThemeLabel()}`"
    >
      <span class="theme-icon" :class="iconClasses">
        {{ getThemeIcon() }}
      </span>
      <span v-if="showLabel" class="theme-label">
        {{ getThemeLabel() }}
      </span>
    </button>
    
    <!-- Dropdown for advanced theme selection -->
    <div v-if="showDropdown && isDropdownOpen" class="theme-dropdown">
      <button
        v-for="themeOption in themeOptions"
        :key="themeOption.value"
        :class="getDropdownItemClasses(themeOption.value)"
        @click="selectTheme(themeOption.value)"
      >
        <span class="dropdown-icon">{{ themeOption.icon }}</span>
        <span class="dropdown-label">{{ themeOption.label }}</span>
        <span v-if="theme === themeOption.value" class="dropdown-check">✓</span>
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useTheme, type ThemeMode } from '../../composables/useTheme'

interface Props {
  variant?: 'button' | 'icon' | 'dropdown'
  size?: 'sm' | 'md' | 'lg'
  showLabel?: boolean
  showDropdown?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  variant: 'icon',
  size: 'md',
  showLabel: false,
  showDropdown: false,
})

const {
  theme,
  effectiveTheme,
  setTheme,
  toggleTheme,
  getThemeIcon,
  getThemeLabel,
  isLight,
  isDark,
  isAuto,
} = useTheme()

const isDropdownOpen = ref(false)

const themeOptions = [
  { value: 'light' as ThemeMode, label: 'Light', icon: '☀️' },
  { value: 'dark' as ThemeMode, label: 'Dark', icon: '🌙' },
  { value: 'auto' as ThemeMode, label: 'Auto', icon: '🔄' },
]

const buttonClasses = computed(() => {
  const classes = ['theme-toggle-button']
  
  // Variant classes
  classes.push(`theme-toggle-${props.variant}`)
  
  // Size classes
  const sizeClasses = {
    sm: 'p-1.5 text-sm',
    md: 'p-2 text-base',
    lg: 'p-3 text-lg',
  }
  classes.push(sizeClasses[props.size])
  
  // State classes
  classes.push('interactive')
  classes.push('focus-ring')
  
  return classes.join(' ')
})

const iconClasses = computed(() => {
  const classes = ['theme-icon-inner']
  
  // Add rotation animation class
  if (effectiveTheme.value === 'dark') {
    classes.push('rotate-180')
  }
  
  return classes.join(' ')
})

const getNextThemeLabel = () => {
  const currentIndex = themeOptions.findIndex(option => option.value === theme.value)
  const nextIndex = (currentIndex + 1) % themeOptions.length
  return themeOptions[nextIndex].label
}

const handleToggle = () => {
  if (props.showDropdown) {
    isDropdownOpen.value = !isDropdownOpen.value
  } else {
    toggleTheme()
  }
}

const selectTheme = (selectedTheme: ThemeMode) => {
  setTheme(selectedTheme)
  isDropdownOpen.value = false
}

const getDropdownItemClasses = (themeValue: ThemeMode) => {
  const classes = ['dropdown-item']
  
  if (theme.value === themeValue) {
    classes.push('dropdown-item-active')
  }
  
  return classes.join(' ')
}

// Close dropdown when clicking outside
const handleClickOutside = (event: Event) => {
  const target = event.target as HTMLElement
  if (!target.closest('.theme-toggle')) {
    isDropdownOpen.value = false
  }
}

// Setup click outside listener
if (props.showDropdown) {
  document.addEventListener('click', handleClickOutside)
}
</script>

<style scoped>
.theme-toggle {
  @apply relative inline-block;
}

.theme-toggle-button {
  @apply rounded-md border border-primary transition-all duration-200;
  @apply bg-surface-primary text-text-primary;
  @apply hover:bg-surface-secondary hover:shadow-md;
  @apply active:scale-95;
}

.theme-toggle-icon {
  @apply flex items-center justify-center;
}

.theme-toggle-button {
  @apply flex items-center gap-2;
}

.theme-toggle-dropdown {
  @apply cursor-pointer;
}

.theme-icon {
  @apply inline-block transition-transform duration-300;
}

.theme-icon-inner {
  @apply transition-transform duration-300 ease-bounce;
}

.theme-label {
  @apply font-medium;
}

.theme-dropdown {
  @apply absolute top-full left-0 mt-2 py-2 min-w-32;
  @apply bg-surface-elevated border border-primary rounded-md shadow-lg;
  @apply z-50 animate-slide-up;
}

.dropdown-item {
  @apply w-full px-3 py-2 text-left flex items-center gap-2;
  @apply text-text-primary hover:bg-surface-secondary;
  @apply transition-colors duration-150;
}

.dropdown-item-active {
  @apply bg-surface-secondary text-primary-600;
}

.dropdown-icon {
  @apply text-sm;
}

.dropdown-label {
  @apply flex-1 font-medium;
}

.dropdown-check {
  @apply text-primary-500 font-bold;
}

/* Dark theme specific styles */
[data-theme="dark"] .theme-toggle-button {
  @apply border-border-primary;
}

/* Animation for theme transitions */
.theme-transitioning .theme-toggle-button {
  @apply transition-all duration-200;
}

/* Accessibility improvements */
.theme-toggle-button:focus-visible {
  @apply ring-2 ring-primary-500 ring-offset-2;
}

/* Reduced motion support */
@media (prefers-reduced-motion: reduce) {
  .theme-icon-inner {
    @apply transition-none;
  }
  
  .theme-dropdown {
    @apply animate-none;
  }
}
</style>