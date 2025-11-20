<script setup lang="ts">
import { onMounted, onUnmounted, ref } from 'vue'
import * as THREE from 'three'

const containerRef = ref<HTMLElement | null>(null)
let scene: THREE.Scene
let camera: THREE.PerspectiveCamera
let renderer: THREE.WebGLRenderer
let particles: THREE.Points
let animationId: number

// Colors from design system
const COLOR_PURPLE = 0xa78bbf
const COLOR_GOLD = 0xc8a667

const initThree = () => {
  if (!containerRef.value) return

  // 1. Scene
  scene = new THREE.Scene()
  // Transparent background to blend with CSS background
  scene.background = null

  // 2. Camera
  const width = containerRef.value.clientWidth
  const height = containerRef.value.clientHeight
  camera = new THREE.PerspectiveCamera(75, width / height, 0.1, 1000)
  camera.position.z = 30

  // 3. Renderer
  renderer = new THREE.WebGLRenderer({ alpha: true, antialias: true })
  renderer.setSize(width, height)
  renderer.setPixelRatio(Math.min(window.devicePixelRatio, 2))
  containerRef.value.appendChild(renderer.domElement)

  // 4. Particles
  const geometry = new THREE.BufferGeometry()
  const count = 1500
  const positions = new Float32Array(count * 3)
  const colors = new Float32Array(count * 3)
  const sizes = new Float32Array(count)

  const color1 = new THREE.Color(COLOR_PURPLE)
  const color2 = new THREE.Color(COLOR_GOLD)

  for (let i = 0; i < count; i++) {
    // Sphere distribution
    const r = 15 + Math.random() * 10
    const theta = Math.random() * Math.PI * 2
    const phi = Math.acos(2 * Math.random() - 1)

    positions[i * 3] = r * Math.sin(phi) * Math.cos(theta)
    positions[i * 3 + 1] = r * Math.sin(phi) * Math.sin(theta)
    positions[i * 3 + 2] = r * Math.cos(phi)

    // Mix colors
    const mixedColor = Math.random() > 0.5 ? color1 : color2
    colors[i * 3] = mixedColor.r
    colors[i * 3 + 1] = mixedColor.g
    colors[i * 3 + 2] = mixedColor.b

    sizes[i] = Math.random() * 0.5
  }

  geometry.setAttribute('position', new THREE.BufferAttribute(positions, 3))
  geometry.setAttribute('color', new THREE.BufferAttribute(colors, 3))
  // geometry.setAttribute('size', new THREE.BufferAttribute(sizes, 1)) // Custom shader needed for size attribute usually, or use PointsMaterial size

  const material = new THREE.PointsMaterial({
    size: 0.15,
    vertexColors: true,
    transparent: true,
    opacity: 0.8,
    sizeAttenuation: true,
  })

  particles = new THREE.Points(geometry, material)
  scene.add(particles)

  // 5. Lights (Optional for Points but good for scene)
  const ambientLight = new THREE.AmbientLight(0xffffff, 0.5)
  scene.add(ambientLight)
}

const animate = () => {
  animationId = requestAnimationFrame(animate)

  if (particles) {
    particles.rotation.y += 0.001
    particles.rotation.x += 0.0005

    // Gentle wave effect could be added here by updating positions
  }

  renderer.render(scene, camera)
}

const handleResize = () => {
  if (!containerRef.value || !camera || !renderer) return

  const width = containerRef.value.clientWidth
  const height = containerRef.value.clientHeight

  camera.aspect = width / height
  camera.updateProjectionMatrix()
  renderer.setSize(width, height)
}

onMounted(() => {
  initThree()
  animate()
  window.addEventListener('resize', handleResize)
})

onUnmounted(() => {
  window.removeEventListener('resize', handleResize)
  cancelAnimationFrame(animationId)
  if (renderer) {
    renderer.dispose()
    containerRef.value?.removeChild(renderer.domElement)
  }
  if (particles) {
    particles.geometry.dispose()
    ;(particles.material as THREE.Material).dispose()
  }
})
</script>

<template>
  <div ref="containerRef" class="three-container"></div>
</template>

<style scoped>
.three-container {
  width: 100%;
  height: 100%;
  position: absolute;
  top: 0;
  left: 0;
  pointer-events: none; /* Let clicks pass through to underlying elements if needed */
  z-index: 1;
}
</style>
