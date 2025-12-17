<script setup lang="ts">
import { onMounted, onBeforeUnmount, ref } from 'vue'
import * as THREE from 'three'
import { CSS3DRenderer, CSS3DObject } from 'three/examples/jsm/renderers/CSS3DRenderer.js'
import gsap from 'gsap'
import bgUrl from '@/assets/bg.webp'

const containerRef = ref<HTMLDivElement | null>(null)
const cardContentRef = ref<HTMLDivElement | null>(null)

const initScene = () => {
    if (!containerRef.value || !cardContentRef.value) return

    // --- 1. Setup Basic Scene ---
    const width = window.innerWidth
    const height = window.innerHeight

    const scene = new THREE.Scene()
    const cssScene = new THREE.Scene()

    // Camera setup to match pixels 1:1 at z=0
    const fov = 45
    const camera = new THREE.PerspectiveCamera(fov, width / height, 1, 5000)
    const zDist = (height / 2) / Math.tan((fov * Math.PI / 180) / 2)
    camera.position.set(0, 0, zDist)

    // --- 2. Renderers ---
    // WebGL Renderer (for the card body & shadows)
    const webglRenderer = new THREE.WebGLRenderer({
        alpha: true,
        antialias: true
    })
    webglRenderer.setSize(width, height)
    webglRenderer.setPixelRatio(window.devicePixelRatio)
    webglRenderer.shadowMap.enabled = true
    webglRenderer.shadowMap.type = THREE.PCFSoftShadowMap
    webglRenderer.domElement.style.position = 'absolute'
    webglRenderer.domElement.style.top = '0'
    webglRenderer.domElement.style.zIndex = '1'
    webglRenderer.domElement.style.pointerEvents = 'none' // Let events pass to CSS3D
    containerRef.value.appendChild(webglRenderer.domElement)

    // CSS3D Renderer (for the HTML content)
    const cssRenderer = new CSS3DRenderer()
    cssRenderer.setSize(width, height)
    cssRenderer.domElement.style.position = 'absolute'
    cssRenderer.domElement.style.top = '0'
    cssRenderer.domElement.style.zIndex = '2' // On top of WebGL
    containerRef.value.appendChild(cssRenderer.domElement)

    // --- 3. Objects ---

    // Group to move both WebGL and CSS objects together
    const cardGroup = new THREE.Group()
    scene.add(cardGroup)
    cssScene.add(cardGroup)

    // Card Dimensions (match the HTML div size)
    const cardWidth = 580
    const cardHeight = 380

    // 3.1 WebGL Card Body (The "Physical" Card)
    const geometry = new THREE.PlaneGeometry(cardWidth, cardHeight, 32, 32)
    // Glass/Paper Material
    const material = new THREE.MeshPhysicalMaterial({
        color: 0xffffff,
        metalness: 0.1,
        roughness: 0.4, // Frosted glass roughness
        transmission: 0.0, // Use opacity for transparency to let CSS blur show
        transparent: true,
        opacity: 0.1, // Very subtle glass body
        thickness: 2,
        clearcoat: 0.5,
        clearcoatRoughness: 0.1,
        side: THREE.DoubleSide
    })
    const mesh = new THREE.Mesh(geometry, material)
    mesh.castShadow = true
    mesh.receiveShadow = true
    cardGroup.add(mesh)

    // 3.2 CSS3D Object (The HTML Content)
    const object = new CSS3DObject(cardContentRef.value)
    // CSS3DObject is transparent by default, content is in the div
    cardGroup.add(object)

    // 3.3 Shadow Plane (Invisible plane to catch shadow)
    const shadowPlaneGeo = new THREE.PlaneGeometry(width * 2, height * 2)
    const shadowPlaneMat = new THREE.ShadowMaterial({
        opacity: 0.2, // Slightly darker to be visible on light bg
        color: 0x1a1a1a // Soft charcoal shadow
    })
    const shadowPlane = new THREE.Mesh(shadowPlaneGeo, shadowPlaneMat)
    shadowPlane.position.z = -50 // Behind the card
    shadowPlane.receiveShadow = true
    scene.add(shadowPlane)

    // --- 4. Lighting ---
    const ambientLight = new THREE.AmbientLight(0xffffff, 0.6)
    scene.add(ambientLight)

    const dirLight = new THREE.DirectionalLight(0xffffff, 0.8)
    dirLight.position.set(-200, 500, 500) // Top-left light source to match bg
    dirLight.castShadow = true
    dirLight.shadow.mapSize.width = 2048
    dirLight.shadow.mapSize.height = 2048
    dirLight.shadow.camera.near = 0.5
    dirLight.shadow.camera.far = 5000
    // Adjust shadow camera frustum to cover the scene
    const d = 1000
    dirLight.shadow.camera.left = -d
    dirLight.shadow.camera.right = d
    dirLight.shadow.camera.top = d
    dirLight.shadow.camera.bottom = -d
    scene.add(dirLight)

    // SpotLight for sheen
    const spotLight = new THREE.SpotLight(0xfff0dd, 2.0) // Increased intensity
    spotLight.position.set(-200, 200, 600)
    spotLight.angle = Math.PI / 6
    spotLight.penumbra = 0.5
    scene.add(spotLight)

    // --- 5. Initial State & Animation ---

    // Initial Tilt (30 degrees approx)
    // Using Euler angles. 30 deg = Math.PI / 6
    const initialRotX = -Math.PI / 12 // 15 deg tilt back
    const initialRotY = Math.PI / 12  // 15 deg tilt side
    // Total tilt visual approx 30 deg combined

    cardGroup.rotation.x = initialRotX
    cardGroup.rotation.y = initialRotY

    // Hover Interaction
    const onHover = () => {
        gsap.to(cardGroup.position, {
            y: 30, // Levitate up
            z: 50, // Move closer
            duration: 0.6,
            ease: 'power2.out'
        })
        gsap.to(cardGroup.rotation, {
            x: 0, // Straighten up
            y: 0,
            duration: 0.6,
            ease: 'power2.out'
        })
        // Lift shadow plane slightly or fade it? 
        // Actually shadow moves naturally with the object.
    }

    const onLeave = () => {
        gsap.to(cardGroup.position, {
            y: 0,
            z: 0,
            duration: 0.8,
            ease: 'power2.inOut'
        })
        gsap.to(cardGroup.rotation, {
            x: initialRotX,
            y: initialRotY,
            duration: 0.8,
            ease: 'power2.inOut'
        })
    }

    // Attach events to the DOM element (the card content div)
    // Note: CSS3DObject makes the div interactive.
    if (cardContentRef.value) {
        cardContentRef.value.addEventListener('mouseenter', onHover)
        cardContentRef.value.addEventListener('mouseleave', onLeave)
    }

    // --- 6. Loop ---
    const animate = () => {
        requestAnimationFrame(animate)
        webglRenderer.render(scene, camera)
        cssRenderer.render(cssScene, camera)
    }
    animate()

    // Resize
    const onResize = () => {
        const w = window.innerWidth
        const h = window.innerHeight
        camera.aspect = w / h
        // Recalculate Z to keep 1:1 pixel ratio
        camera.position.z = (h / 2) / Math.tan((fov * Math.PI / 180) / 2)
        camera.updateProjectionMatrix()
        webglRenderer.setSize(w, h)
        cssRenderer.setSize(w, h)
    }
    window.addEventListener('resize', onResize)

    return () => {
        window.removeEventListener('resize', onResize)
        if (cardContentRef.value) {
            cardContentRef.value.removeEventListener('mouseenter', onHover)
            cardContentRef.value.removeEventListener('mouseleave', onLeave)
        }
        containerRef.value?.removeChild(webglRenderer.domElement)
        containerRef.value?.removeChild(cssRenderer.domElement)
        webglRenderer.dispose()
    }
}

onMounted(() => {
    const cleanup = initScene()
    onBeforeUnmount(() => cleanup && cleanup())
})
</script>

<template>
    <div class="min-h-screen flex items-center justify-center relative overflow-hidden"
        :style="{ backgroundImage: `url(${bgUrl})`, backgroundSize: 'cover', backgroundPosition: 'center' }">

        <!-- 3D Container -->
        <div ref="containerRef" class="absolute inset-0 z-10"></div>

        <!-- Card Content (Hidden from normal flow, used by CSS3D) -->
        <div class="hidden">
            <div ref="cardContentRef"
                class="w-[580px] h-[380px] bg-white/60 backdrop-blur-2xl border border-white/50 rounded-3xl p-12 flex flex-col items-center justify-center text-center shadow-[0_20px_40px_-10px_rgba(200,200,200,0.4)] select-none cursor-pointer">

                <p class="eyebrow mb-8 tracking-[0.4em] text-stone/80">PREVIEW MODE</p>

                <h1 class="font-display text-6xl italic mb-4 text-charcoal">White Phantom</h1>
                <h2 class="font-serif text-2xl text-stone mb-8">Site Under Construction</h2>

                <p class="text-stone leading-relaxed mb-10 font-light max-w-md">
                    当前站点处于开发预览模式。<br />
                    悬浮卡片以解锁 Linear Luxury 的极致体验。
                </p>

                <div class="flex gap-4">
                    <a class="hero-button hero-button--primary nav-link inline-flex items-center justify-center px-10 py-3 text-sm"
                        href="/" aria-label="刷新页面">
                        刷新页面
                    </a>
                </div>
            </div>
        </div>
    </div>
</template>

<style scoped>
/* Ensure the hidden content is still renderable by CSS3D */
.hidden {
    display: none;
}
</style>
