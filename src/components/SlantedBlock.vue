<script setup lang="ts">
import { computed } from 'vue'

interface Props {
  direction?: 'left' | 'right'
  bgColor?: string // CSS variable or color value
  overlayColor?: string
  image?: string
  height?: string
  slantedMask?: boolean // Enable gradient mask mode
  maskAngle?: string // Custom angle for gradient mask (e.g. "45deg")
  maskColor?: string // Color for the gradient mask
}

const props = withDefaults(defineProps<Props>(), {
  direction: 'right',
  bgColor: 'var(--color-bg-hero)',
  height: 'auto',
  slantedMask: false,
  maskAngle: '45deg',
  maskColor: '#ffffff'
})

const clipPath = computed(() => {
  if (props.slantedMask) return 'none' // No physical clip in mask mode

  if (props.direction === 'right') {
    return 'polygon(0 0, 100% 0, 100% 85%, 0 100%)'
  } else {
    return 'polygon(0 0, 100% 0, 100% 100%, 0 85%)'
  }
})

const style = computed(() => ({
  backgroundColor: props.bgColor,
  clipPath: clipPath.value,
  minHeight: props.height,
  position: 'relative' as const,
}))

const overlayStyle = computed(() => {
  if (props.slantedMask) {
    // Gradient mask logic
    const angle = props.maskAngle
    
    // Use color-mix to create a transparent version of the mask color
    // This ensures the gradient fades to transparent of the SAME color (avoiding gray/blackish fade)
    const startColor = props.maskColor
    const endColor = `color-mix(in srgb, ${startColor}, transparent)`
    
    return {
        background: `linear-gradient(${angle}, ${startColor} 30%, ${endColor} 70%)`,
        position: 'absolute' as const,
        top: 0,
        left: 0,
        width: '100%',
        height: '100%',
        zIndex: 1,
    }
  }

  return {
    backgroundColor: props.overlayColor,
    position: 'absolute' as const,
    top: 0,
    left: 0,
    width: '100%',
    height: '100%',
    zIndex: 1,
  }
})
</script>

<template>
  <div class="slanted-block" :style="style">
    <div
      v-if="image"
      class="slanted-block__bg-image"
      :style="{ backgroundImage: `url(${image})` }"
    ></div>
    <div v-if="overlayColor" class="slanted-block__overlay" :style="overlayStyle"></div>
    <div class="slanted-block__content">
      <slot></slot>
    </div>
  </div>
</template>

<style scoped lang="scss">
.slanted-block {
  width: 100%;
  overflow: hidden;

  &__bg-image {
    position: absolute;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    background-size: cover;
    background-position: center;
    z-index: 0;
  }

  &__content {
    position: relative;
    z-index: 2;
    height: 100%;
  }
}
</style>
