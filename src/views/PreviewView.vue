<script setup lang="ts">
import { onMounted, onBeforeUnmount, ref } from 'vue'
import * as THREE from 'three'
import { CSS3DRenderer, CSS3DObject } from 'three/examples/jsm/renderers/CSS3DRenderer.js'
import { RoomEnvironment } from 'three/examples/jsm/environments/RoomEnvironment.js'
import { RoundedBoxGeometry } from 'three/examples/jsm/geometries/RoundedBoxGeometry.js'
import { RectAreaLightUniformsLib } from 'three/examples/jsm/lights/RectAreaLightUniformsLib.js'
import gsap from 'gsap'
import bgUrl from '@/assets/bg.webp'
import logoUrl from '@/assets/logo.webp'

// Logo tint via CSS filter (keeps original intrinsic size/aspect ratio).
// Black (emphasis): brightness(0) saturate(100%)
// If you want other colors later, we can provide preset filter strings.
const logoFilter = 'brightness(0) saturate(100%)'

const containerRef = ref<HTMLDivElement | null>(null)
const cardContentRef = ref<HTMLDivElement | null>(null)
const logoRef = ref<HTMLDivElement | null>(null)

const initScene = () => {
    if (!containerRef.value || !cardContentRef.value || !logoRef.value) return

    // --- 1. Setup Basic Scene ---
    const width = window.innerWidth
    const height = window.innerHeight

    const scene = new THREE.Scene()
    const cssScene = new THREE.Scene()

    // Camera setup
    const fov = 45
    const camera = new THREE.PerspectiveCamera(fov, width / height, 1, 5000)
    const zDist = (height / 2) / Math.tan((fov * Math.PI / 180) / 2)
    camera.position.set(0, 0, zDist)

    // --- 2. Renderers ---
    const webglRenderer = new THREE.WebGLRenderer({
        alpha: true,
        antialias: true,
        powerPreference: "high-performance"
    })
    webglRenderer.setSize(width, height)
    webglRenderer.setPixelRatio(Math.min(window.devicePixelRatio, 2))
    webglRenderer.outputColorSpace = THREE.SRGBColorSpace
    webglRenderer.shadowMap.enabled = true
    webglRenderer.shadowMap.type = THREE.PCFSoftShadowMap
    webglRenderer.toneMapping = THREE.ACESFilmicToneMapping
    webglRenderer.toneMappingExposure = 1.0
    webglRenderer.domElement.style.position = 'absolute'
    webglRenderer.domElement.style.top = '0'
    webglRenderer.domElement.style.zIndex = '1'
    webglRenderer.domElement.style.pointerEvents = 'none'
    containerRef.value.appendChild(webglRenderer.domElement)

    const cssRenderer = new CSS3DRenderer()
    cssRenderer.setSize(width, height)
    cssRenderer.domElement.style.position = 'absolute'
    cssRenderer.domElement.style.top = '0'
    cssRenderer.domElement.style.zIndex = '2'
    containerRef.value.appendChild(cssRenderer.domElement)

    // --- 2.1 Environment (critical for crystal reflections) ---
    RectAreaLightUniformsLib.init()
    const pmremGenerator = new THREE.PMREMGenerator(webglRenderer)
    pmremGenerator.compileEquirectangularShader()
    const envTexture = pmremGenerator.fromScene(new RoomEnvironment(), 0.04).texture
    scene.environment = envTexture

    // --- 3. Objects ---
    // NOTE: The same Object3D cannot belong to two scenes.
    // If we add a group to cssScene after scene.add(), it will be removed from the WebGL scene.
    // So we keep two groups and sync their transforms each frame.
    const webglGroup = new THREE.Group()
    const cssGroup = new THREE.Group()
    scene.add(webglGroup)
    cssScene.add(cssGroup)

    const cardWidth = 600
    const cardHeight = 400
    const cardThickness = 30

    // 3.0 Background Plane (for contrast; transmission still needs reflections to read as crystal)
    const textureLoader = new THREE.TextureLoader()
    const bgTexture = textureLoader.load(bgUrl)
    bgTexture.colorSpace = THREE.SRGBColorSpace
    // A touch of softness: rely on mipmaps + linear filtering for a subtle blur.
    bgTexture.generateMipmaps = true
    bgTexture.minFilter = THREE.LinearMipmapLinearFilter
    bgTexture.magFilter = THREE.LinearFilter
    bgTexture.anisotropy = 1
    bgTexture.needsUpdate = true

    const bgGeo = new THREE.PlaneGeometry(1, 1)
    const bgMat = new THREE.MeshBasicMaterial({ map: bgTexture })
    // Avoid background plane writing depth; it should never occlude the glass/card.
    bgMat.depthWrite = false
    const bgMesh = new THREE.Mesh(bgGeo, bgMat)
    const bgZ = -800
    // Keep background in world space to preserve depth parallax.
    bgMesh.position.set(0, 0, bgZ)
    scene.add(bgMesh)
    bgMesh.renderOrder = -1000
    bgMesh.frustumCulled = false

    // Background fitting helpers (avoid edges even when camera translates/rotates)
    const bgTarget = new THREE.Vector3(0, 0, 0)
    const bgDir = new THREE.Vector3()
    const bgCenter = new THREE.Vector3()
    const bgQuatInv = new THREE.Quaternion()
    const bgOffsetLocal = new THREE.Vector3()

    const updateBackground = () => {
        // Always face the camera so the plane remains a true “backdrop”.
        bgMesh.quaternion.copy(camera.quaternion)

        // Find where the view-center ray (camera -> target) hits z = bgZ.
        bgDir.subVectors(bgTarget, camera.position).normalize()
        // Guard: avoid division by ~0 (shouldn't happen in this scene).
        if (Math.abs(bgDir.z) < 1e-6) {
            return
        }
        const t = (bgZ - camera.position.z) / bgDir.z
        bgCenter.copy(camera.position).addScaledVector(bgDir, t)

        // Base size needed at that depth.
        const dist = camera.position.distanceTo(bgCenter)
        const vHeight = 2 * Math.tan(THREE.MathUtils.degToRad(fov / 2)) * dist
        const vWidth = vHeight * camera.aspect

        // If the plane's center isn't on the view-center hit point, extend size to cover.
        bgQuatInv.copy(bgMesh.quaternion).invert()
        bgOffsetLocal.copy(bgCenter).sub(bgMesh.position).applyQuaternion(bgQuatInv)
        const extraX = Math.abs(bgOffsetLocal.x) * 2
        const extraY = Math.abs(bgOffsetLocal.y) * 2

        // Extra margin to hide sub-pixel seams.
        const margin = Math.max(vWidth, vHeight) * 0.06
        bgMesh.scale.set(vWidth + extraX + margin, vHeight + extraY + margin, 1)
    }
    updateBackground()

    // 3.1 WebGL Card Body (Thick Crystal Block)
    // Using BoxGeometry for thickness
    const geometry = new RoundedBoxGeometry(cardWidth, cardHeight, cardThickness, 10, 10)

    // Crystal Material
    const material = new THREE.MeshPhysicalMaterial({
        color: 0xfffbf3,
        metalness: 0.0,
        roughness: 0.08,
        transmission: 1.0,
        thickness: cardThickness * 0.9,
        ior: 1.45,
        dispersion: 0.9,
        clearcoat: 1.0,
        clearcoatRoughness: 0.06,
        envMapIntensity: 1.4,
        specularIntensity: 1.0,
        specularColor: new THREE.Color(0xffffff),
        attenuationColor: new THREE.Color(0xfff2df),
        attenuationDistance: 650,
        transparent: true,
        opacity: 0.98
    })

    const mesh = new THREE.Mesh(geometry, material)
    mesh.castShadow = true
    mesh.receiveShadow = true
    webglGroup.add(mesh)

    // Edge highlight (gives the “crystal slab” read even when refraction is subtle)
    const edges = new THREE.EdgesGeometry(geometry, 35)
    const edgeLines = new THREE.LineSegments(
        edges,
        new THREE.LineBasicMaterial({
            color: 0xffffff,
            transparent: true,
            opacity: 0.22
        })
    )
    edgeLines.scale.set(1.002, 1.002, 1.002)
    webglGroup.add(edgeLines)

    // 3.2 CSS3D Object (Card Content)
    // Position it slightly in front of the glass block face
    const object = new CSS3DObject(cardContentRef.value)
    object.position.z = cardThickness / 2 + 1
    cssGroup.add(object)

    // 3.3 CSS3D Object (Logo)
    // Position logo inside or just above
    const logoObject = new CSS3DObject(logoRef.value)
    // Keep it near the top edge to avoid overlapping the title text.
    // Also keep Z close to content plane so it doesn't appear overly large.
    logoObject.position.z = cardThickness / 2 + 2
    logoObject.position.y = cardHeight / 2 - 60
    logoObject.scale.setScalar(0.92)
    cssGroup.add(logoObject)

    // --- 4. Lighting ---
    const ambientLight = new THREE.AmbientLight(0xffffff, 0.25)
    scene.add(ambientLight)

    // Main Directional Light (Sun)
    const dirLight = new THREE.DirectionalLight(0xffffff, 1.4)
    dirLight.position.set(500, 500, 1000)
    dirLight.castShadow = true
    dirLight.shadow.mapSize.width = 2048
    dirLight.shadow.mapSize.height = 2048
    scene.add(dirLight)

    // Rim Light (for edges)
    const spotLight = new THREE.SpotLight(0xffeebb, 4.0)
    spotLight.position.set(-500, 500, 0)
    spotLight.lookAt(0, 0, 0)
    spotLight.angle = Math.PI / 4
    spotLight.penumbra = 0.5
    scene.add(spotLight)

    // Bottom fill light
    const rectLight = new THREE.RectAreaLight(0xffffff, 2.0, cardWidth, cardHeight)
    rectLight.position.set(0, -200, 100)
    rectLight.lookAt(0, 0, 0)
    scene.add(rectLight)

    // --- 5. Interaction & Parallax ---
    let mouseX = 0
    let mouseY = 0

    const onMouseMove = (e: MouseEvent) => {
        mouseX = (e.clientX / window.innerWidth) * 2 - 1
        mouseY = -(e.clientY / window.innerHeight) * 2 + 1
    }
    window.addEventListener('mousemove', onMouseMove)

    // Hover Interaction
    const onHover = () => {
        gsap.to(webglGroup.rotation, {
            x: -Math.PI / 32, // Slight tilt up
            duration: 0.8,
            ease: 'power2.out'
        })
        gsap.to(webglGroup.position, {
            z: 20,
            duration: 0.8,
            ease: 'power2.out'
        })
    }

    const onLeave = () => {
        gsap.to(webglGroup.rotation, {
            x: 0,
            duration: 1.0,
            ease: 'power2.inOut'
        })
        gsap.to(webglGroup.position, {
            z: 0,
            duration: 1.0,
            ease: 'power2.inOut'
        })
    }

    if (cardContentRef.value) {
        cardContentRef.value.addEventListener('mouseenter', onHover)
        cardContentRef.value.addEventListener('mouseleave', onLeave)
    }

    // --- 6. Loop ---
    let rafId = 0
    const animate = () => {
        rafId = requestAnimationFrame(animate)

        // Smooth Mouse Parallax
        // Move camera slightly based on mouse to create depth
        camera.position.x += (mouseX * 50 - camera.position.x) * 0.05
        camera.position.y += (mouseY * 50 - camera.position.y) * 0.05
        camera.lookAt(scene.position)

        // Keep background perfectly covered while preserving parallax.
        updateBackground()

        // Keep CSS3D in perfect lockstep with WebGL transforms
        cssGroup.position.copy(webglGroup.position)
        cssGroup.quaternion.copy(webglGroup.quaternion)
        cssGroup.scale.copy(webglGroup.scale)

        webglRenderer.render(scene, camera)
        cssRenderer.render(cssScene, camera)
    }
    animate()

    const onResize = () => {
        const w = window.innerWidth
        const h = window.innerHeight
        camera.aspect = w / h
        camera.position.z = (h / 2) / Math.tan((fov * Math.PI / 180) / 2)
        camera.updateProjectionMatrix()
        webglRenderer.setSize(w, h)
        cssRenderer.setSize(w, h)
        updateBackground()
    }
    window.addEventListener('resize', onResize)

    return () => {
        window.removeEventListener('resize', onResize)
        window.removeEventListener('mousemove', onMouseMove)
        if (cardContentRef.value) {
            cardContentRef.value.removeEventListener('mouseenter', onHover)
            cardContentRef.value.removeEventListener('mouseleave', onLeave)
        }
        if (rafId) cancelAnimationFrame(rafId)
        containerRef.value?.removeChild(webglRenderer.domElement)
        containerRef.value?.removeChild(cssRenderer.domElement)
        envTexture.dispose()
        pmremGenerator.dispose()
        webglRenderer.dispose()
    }
}

onMounted(() => {
    const cleanup = initScene()
    onBeforeUnmount(() => cleanup && cleanup())
})
</script>

<template>
    <div class="min-h-screen flex items-center justify-center relative overflow-hidden bg-stone-100">

        <!-- 3D Container -->
        <div ref="containerRef" class="absolute inset-0 z-10"></div>

        <!-- Hidden Content for CSS3D -->
        <div class="hidden">
            <!-- 1. Logo Layer (Floating) -->
            <div ref="logoRef" class="text-center select-none pointer-events-none flex flex-col items-center">
                <img :src="logoUrl" alt="Logo" class="w-48 h-auto opacity-95" :style="{ filter: logoFilter }" />
            </div>

            <!-- 2. Card Body Layer (Content Only) -->
            <!-- Note: No background, no border, just text. The glass block provides the body. -->
            <div ref="cardContentRef"
                class="w-[600px] h-[400px] flex flex-col items-center justify-center text-center cursor-pointer p-12 select-none">

                <div class="flex flex-col items-center gap-8 mt-16">
                    <div class="space-y-3">
                        <h2 class="font-serif italic text-4xl text-[#1A1A1A] tracking-wide">
                            The Atelier
                        </h2>
                        <div class="flex flex-col gap-1">
                            <p class="font-sans text-[11px] text-[#595959] uppercase tracking-[0.2em]">
                                Site Under Construction
                            </p>
                            <p class="font-sans text-[10px] text-[#999] tracking-[0.15em]">
                                官网正在开发中
                            </p>
                        </div>
                    </div>

                    <a class="group relative inline-flex items-center justify-center px-8 py-2.5 overflow-hidden transition-all border border-[#1A1A1A] bg-transparent hover:bg-gradient-to-br hover:from-white hover:to-[#F2F0EA] hover:border-white/0 text-[#1A1A1A]"
                        href="/" aria-label="Enter Gallery">
                        <span class="relative text-[10px] tracking-[0.25em] uppercase font-medium">
                            Enter Gallery
                        </span>
                    </a>
                </div>
            </div>
        </div>
    </div>
</template>

<style scoped>
.hidden {
    display: none;
}
</style>
