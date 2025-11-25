<script setup lang="ts">
import { onMounted, onBeforeUnmount, ref, nextTick } from 'vue'
import gsap from 'gsap'
import { ScrollTrigger } from 'gsap/ScrollTrigger'
import * as THREE from 'three'

gsap.registerPlugin(ScrollTrigger)

type HoverBinding = {
    element: Element
    enter: () => void
    leave: () => void
}

type NetworkInformationLike = {
    effectiveType?: string
    addEventListener?: (type: string, listener: () => void) => void
    removeEventListener?: (type: string, listener: () => void) => void
}

type NavigatorWithConnection = Navigator & {
    connection?: NetworkInformationLike
}

let moveHandler: ((event: MouseEvent) => void) | null = null
const hoverBindings: HoverBinding[] = []
let ctx: gsap.Context | null = null
let cursorDotRef: HTMLElement | null = null
let cursorOutlineRef: HTMLElement | null = null

const silkContainer = ref<HTMLDivElement | null>(null)
const isNavCompacted = ref(false)
const isReducedMotion = ref(false)
const shouldUseStaticSilk = ref(false)

type RibbonRuntimeConfig = {
    segments: number
    width: number
    length: number
    speed: number
    twistSpeed: number
    twistAmplitude: number
    flowFrequency: number
    baseColor: THREE.Color
    glowColor: THREE.Color
}

let scene: THREE.Scene | null = null
let camera: THREE.PerspectiveCamera | null = null
let renderer: THREE.WebGLRenderer | null = null
let ribbonMesh: THREE.Mesh<THREE.PlaneGeometry, THREE.MeshPhysicalMaterial> | null = null
let silkMaterial: THREE.MeshPhysicalMaterial | null = null
let animationFrameId: number | null = null
let resizeHandler: (() => void) | null = null
let time = 0
let scrollHandler: (() => void) | null = null
let motionQuery: MediaQueryList | null = null
let motionChangeHandler: ((event: MediaQueryListEvent) => void) | null = null
let networkInfo: NetworkInformationLike | null = null
let networkChangeHandler: (() => void) | null = null
let silkPointerHandler: ((event: PointerEvent) => void) | null = null

const heroScrollProgress = {
    target: 0,
    current: 0,
}

const pointerInfluence = {
    targetX: 0,
    targetY: 0,
    currentX: 0,
    currentY: 0,
}

const silkBasePosition = new THREE.Vector3(0, -1, 2)
const silkBaseRotation = { x: Math.PI / 4, y: 0, z: Math.PI / 2.8 }

const baseSilkColor = new THREE.Color(0xf6f0e9)
const highlightSilkColor = new THREE.Color(0xfdfaf2)
const baseGlowColor = new THREE.Color(0xfdf3e3)
const accentGlowColor = new THREE.Color(0xfff2d6)

const targetConfig: Omit<RibbonRuntimeConfig, 'segments' | 'width' | 'length'> = {
    speed: 0.18,
    twistSpeed: 0.06,
    twistAmplitude: 0.75,
    flowFrequency: 0.45,
    baseColor: baseSilkColor.clone(),
    glowColor: baseGlowColor.clone(),
}

const updatePointerTargets = (clientX: number, clientY: number) => {
    if (typeof window === 'undefined') {
        return
    }
    const width = window.innerWidth || 1
    const height = window.innerHeight || 1
    pointerInfluence.targetX = THREE.MathUtils.clamp((clientX / width) * 2 - 1, -1, 1)
    pointerInfluence.targetY = THREE.MathUtils.clamp((clientY / height) * 2 - 1, -1, 1)
}

const syncMotionPreference = (shouldReduce: boolean) => {
    isReducedMotion.value = shouldReduce
    if (shouldReduce || shouldUseStaticSilk.value) {
        disposeSilk()
        return
    }
    nextTick(() => {
        initSilkCanvas()
    })
}

const supportsWebGL = () => {
    if (typeof document === 'undefined') {
        return false
    }
    const canvas = document.createElement('canvas')
    const gl = canvas.getContext('webgl') || canvas.getContext('experimental-webgl')
    return Boolean(gl)
}

const evaluateSilkFallback = () => {
    if (typeof navigator === 'undefined') {
        return
    }
    const nav = navigator as NavigatorWithConnection
    const effectiveType = nav.connection?.effectiveType ?? ''
    const lowPowerCpu = typeof nav.hardwareConcurrency === 'number' && nav.hardwareConcurrency > 0 && nav.hardwareConcurrency <= 4
    const slowNetwork = effectiveType === 'slow-2g' || effectiveType === '2g'
    const lacksWebGL = !supportsWebGL()
    const shouldFallback = lowPowerCpu || slowNetwork || lacksWebGL

    if (shouldFallback === shouldUseStaticSilk.value) {
        return
    }

    shouldUseStaticSilk.value = shouldFallback

    if (shouldFallback) {
        disposeSilk()
    } else if (!isReducedMotion.value) {
        nextTick(() => {
            initSilkCanvas()
        })
    }
}

const config: RibbonRuntimeConfig = {
    segments: 260,
    width: 6,
    length: 26,
    speed: 0.16,
    twistSpeed: 0.06,
    twistAmplitude: 0.8,
    flowFrequency: 0.5,
    baseColor: baseSilkColor.clone(),
    glowColor: baseGlowColor.clone(),
}

const initCursor = () => {
    cursorDotRef = document.querySelector('.cursor-dot') as HTMLElement | null
    cursorOutlineRef = document.querySelector('.cursor-outline') as HTMLElement | null

    if (!cursorDotRef || !cursorOutlineRef) {
        return
    }

    const styles = getComputedStyle(document.documentElement)
    const accentGold = styles.getPropertyValue('--accent-gold').trim() || '#D4AF37'
    const neutralDot = 'rgba(26, 26, 26, 0.45)'
    const veilGlow = 'rgba(255, 255, 255, 0.25)'
    const accentAura = 'rgba(212, 175, 55, 0.35)'
    const idleBorder = 'rgba(26, 26, 26, 0.18)'

    document.body.classList.add('has-custom-cursor')
    cursorDotRef.style.opacity = '1'
    cursorOutlineRef.style.opacity = '1'
    cursorDotRef.style.backgroundColor = neutralDot
    cursorOutlineRef.style.borderColor = idleBorder
    cursorOutlineRef.style.background = veilGlow
    cursorOutlineRef.style.boxShadow = '0 0 15px rgba(255, 255, 255, 0.25)'

    moveHandler = (event: MouseEvent) => {
        const { clientX, clientY } = event
        cursorDotRef!.style.left = `${clientX}px`
        cursorDotRef!.style.top = `${clientY}px`
        cursorOutlineRef!.animate(
            {
                left: `${clientX}px`,
                top: `${clientY}px`,
            },
            { duration: 400, fill: 'forwards' }
        )
    }

    window.addEventListener('mousemove', moveHandler)

    const interactiveElements = Array.from(
        document.querySelectorAll<HTMLElement>(
            'a, button, textarea, .project-item, .hero-cta button, .nav-link, .nav-ghost, .bento-card'
        )
    )

    interactiveElements.forEach((element) => {
        const onEnter = () => {
            cursorOutlineRef!.style.width = '50px'
            cursorOutlineRef!.style.height = '50px'
            cursorOutlineRef!.style.borderColor = accentGold
            cursorOutlineRef!.style.background = 'radial-gradient(circle, rgba(255, 255, 255, 0.45) 0%, transparent 70%)'
            cursorOutlineRef!.style.boxShadow = `0 0 35px ${accentAura}`
            cursorDotRef!.style.backgroundColor = accentGold
        }

        const onLeave = () => {
            cursorOutlineRef!.style.width = '40px'
            cursorOutlineRef!.style.height = '40px'
            cursorOutlineRef!.style.borderColor = idleBorder
            cursorOutlineRef!.style.background = veilGlow
            cursorOutlineRef!.style.boxShadow = '0 0 15px rgba(255, 255, 255, 0.25)'
            cursorDotRef!.style.backgroundColor = neutralDot
        }

        element.addEventListener('mouseenter', onEnter)
        element.addEventListener('mouseleave', onLeave)

        hoverBindings.push({ element, enter: onEnter, leave: onLeave })
    })
}

const updateRibbon = () => {
    if (!ribbonMesh) {
        return
    }

    const geometry = ribbonMesh.geometry
    const positions = geometry.attributes.position as THREE.BufferAttribute
    const colors = geometry.attributes.color as THREE.BufferAttribute

    const widthSegments = config.segments
    const heightSegments = 20 // 对应 initSilkCanvas 中的设置
    const verticesPerRow = widthSegments + 1

    const baseR = config.baseColor.r
    const baseG = config.baseColor.g
    const baseB = config.baseColor.b
    const glowR = config.glowColor.r
    const glowG = config.glowColor.g
    const glowB = config.glowColor.b

    const pointerEnergy = 1 + Math.min(Math.hypot(pointerInfluence.currentX, pointerInfluence.currentY), 1) * 0.65
    const scrollEnvelope = 0.4 + heroScrollProgress.current * 0.9

    for (let col = 0; col < verticesPerRow; col += 1) {
        const ratio = col / widthSegments
        const x = ratio * config.length - config.length / 2

        // 基础波浪 - 更平滑的 S 形
        let waveZ = Math.sin(x * 0.35 + time * 0.8) * (1.6 + scrollEnvelope * 1.1)
        waveZ += Math.sin(x * 1.1 - time * 0.6) * 0.55 * pointerEnergy

        // 中心线 Y 偏移 - 模拟飘动
        const centerY = Math.sin(x * 0.18 + time * 0.35) * (1.2 + scrollEnvelope * 0.5)

        // 扭曲角度 - 更加柔和
        const twist = Math.sin(x * 0.22 + time * config.twistSpeed) * config.twistAmplitude

        // 颜色计算因子
        const flowPhase = ratio * 4 * config.flowFrequency - time * 1.5 + pointerInfluence.currentX * 0.65
        let glowFactor = Math.sin(flowPhase)
        glowFactor = Math.pow((glowFactor + 1) / 2, 6) // 锐化高光

        const twistHighlight = Math.abs(Math.sin(twist + time * 0.5))
        const mixRatio = Math.min(glowFactor * 1.2 + twistHighlight * 0.3, 1)

        const r = baseR + (glowR - baseR) * mixRatio
        const g = baseG + (glowG - baseG) * mixRatio
        const b = baseB + (glowB - baseB) * mixRatio

        // 遍历每一行（宽度方向）
        for (let row = 0; row <= heightSegments; row++) {
            const idx = row * verticesPerRow + col

            // 计算归一化宽度坐标 (0 到 1)
            const v = row / heightSegments
            // 相对于中心的偏移量 (-width/2 到 +width/2)
            const offset = (v - 0.5) * config.width

            // 根据扭曲角度计算最终位置
            // 绕 X 轴旋转 offset
            const y = centerY + offset * Math.cos(twist)
            const z = waveZ + offset * Math.sin(twist)

            positions.setY(idx, y)
            positions.setZ(idx, z)

            // 设置颜色
            colors.setXYZ(idx, r, g, b)
        }
    }

    positions.needsUpdate = true
    colors.needsUpdate = true
    geometry.computeVertexNormals()
}

const animateSilk = () => {
    animationFrameId = requestAnimationFrame(animateSilk)

    pointerInfluence.currentX += (pointerInfluence.targetX - pointerInfluence.currentX) * 0.04
    pointerInfluence.currentY += (pointerInfluence.targetY - pointerInfluence.currentY) * 0.05
    heroScrollProgress.current += (heroScrollProgress.target - heroScrollProgress.current) * 0.08

    const pointerEnergy = Math.min(Math.hypot(pointerInfluence.currentX, pointerInfluence.currentY), 1)
    const scrollEnergy = heroScrollProgress.current
    const energyMix = Math.min(pointerEnergy * 0.6 + scrollEnergy * 0.8, 1)

    targetConfig.speed = 0.18 + scrollEnergy * 0.12
    targetConfig.twistSpeed = 0.05 + pointerEnergy * 0.04
    targetConfig.twistAmplitude = 0.65 + scrollEnergy * 0.4 + pointerEnergy * 0.2
    targetConfig.flowFrequency = 0.4 + scrollEnergy * 0.35
    targetConfig.baseColor.copy(baseSilkColor).lerp(highlightSilkColor, energyMix * 0.6)
    targetConfig.glowColor.copy(baseGlowColor).lerp(accentGlowColor, 0.35 + energyMix * 0.4)

    config.speed += (targetConfig.speed - config.speed) * 0.05
    config.twistSpeed += (targetConfig.twistSpeed - config.twistSpeed) * 0.05
    config.twistAmplitude += (targetConfig.twistAmplitude - config.twistAmplitude) * 0.05
    config.flowFrequency += (targetConfig.flowFrequency - config.flowFrequency) * 0.05
    config.baseColor.lerp(targetConfig.baseColor, 0.05)
    config.glowColor.lerp(targetConfig.glowColor, 0.05)

    time += 0.01 * config.speed
    updateRibbon()

    if (camera) {
        const depth = typeof window !== 'undefined' && window.innerWidth < 768 ? 34 : 28
        camera.position.x = pointerInfluence.currentX * 3.8
        camera.position.y = pointerInfluence.currentY * -2.5 + 0.4
        camera.position.z = depth
        camera.lookAt(0, 0, 0)
    }

    if (ribbonMesh) {
        ribbonMesh.rotation.x = silkBaseRotation.x + pointerInfluence.currentY * 0.12
        ribbonMesh.rotation.y = silkBaseRotation.y + pointerInfluence.currentX * 0.08
        ribbonMesh.rotation.z = silkBaseRotation.z + pointerInfluence.currentX * 0.05
        ribbonMesh.position.x = silkBasePosition.x + pointerInfluence.currentX * 0.85
        ribbonMesh.position.y = silkBasePosition.y + pointerInfluence.currentY * 0.65
    }

    if (renderer && scene && camera) {
        renderer.render(scene, camera)
    }
}

const disposeSilk = () => {
    if (animationFrameId) {
        cancelAnimationFrame(animationFrameId)
        animationFrameId = null
    }

    if (resizeHandler) {
        window.removeEventListener('resize', resizeHandler)
        resizeHandler = null
    }

    if (scene && ribbonMesh) {
        scene.remove(ribbonMesh)
        ribbonMesh.geometry.dispose()
    }

    silkMaterial?.dispose()
    silkMaterial = null
    ribbonMesh = null

    renderer?.dispose()
    if (renderer?.domElement && silkContainer.value?.contains(renderer.domElement)) {
        silkContainer.value.removeChild(renderer.domElement)
    }
    renderer = null
    scene = null
    camera = null
}

const initSilkCanvas = () => {
    if (isReducedMotion.value || shouldUseStaticSilk.value || !silkContainer.value || renderer) {
        return
    }

    const container = silkContainer.value
    scene = new THREE.Scene()
    // scene.fog = new THREE.FogExp2(0x030305, 0.04)

    const rect = container.getBoundingClientRect()
    const width = rect.width || window.innerWidth
    const height = rect.height || window.innerHeight

    camera = new THREE.PerspectiveCamera(45, width / height, 0.1, 1000)
    camera.position.set(0, 0, 30)

    renderer = new THREE.WebGLRenderer({ antialias: true, alpha: true })
    renderer.setPixelRatio(Math.min(window.devicePixelRatio, 2))
    renderer.setSize(width, height)
    renderer.setClearAlpha(0)
    renderer.toneMapping = THREE.ACESFilmicToneMapping
    renderer.toneMappingExposure = 1.0
    renderer.outputColorSpace = THREE.SRGBColorSpace
    container.appendChild(renderer.domElement)

    // 增加宽度方向的分段数，从 2 增加到 20，以解决光照伪影
    const geometry = new THREE.PlaneGeometry(config.length * 1.45, config.width * 1.1, config.segments, 30)
    const positionAttr = geometry.attributes.position as THREE.BufferAttribute | undefined
    if (!positionAttr) {
        console.error('PlaneGeometry is missing position attribute')
        return
    }
    const colorAttr = new THREE.BufferAttribute(new Float32Array(positionAttr.count * 3), 3)
    geometry.setAttribute('color', colorAttr)

    silkMaterial = new THREE.MeshPhysicalMaterial({
        color: 0xffffff,
        vertexColors: true,
        emissive: 0xfaf3e2,
        emissiveIntensity: 0.22,
        metalness: 0.28,
        roughness: 0.18,
        clearcoat: 0.98,
        clearcoatRoughness: 0.15,
        transmission: 0.24,
        thickness: 1.4,
        sheen: 1,
        sheenColor: new THREE.Color(0xfff5df),
        sheenRoughness: 0.55,
        iridescence: 0.28,
        iridescenceIOR: 1.2,
        iridescenceThicknessRange: [120, 320],
        envMapIntensity: 0.4,
        side: THREE.DoubleSide,
        flatShading: false,
    })

    ribbonMesh = new THREE.Mesh(geometry, silkMaterial)

    const updateMeshPosition = () => {
        if (!ribbonMesh) return

        const viewportWidth = window.innerWidth
        const isTablet = viewportWidth < 1280
        const isMobile = viewportWidth < 768

        if (isMobile) {
            silkBaseRotation.x = Math.PI / 2.3
            silkBaseRotation.y = 0.12
            silkBaseRotation.z = Math.PI / 4.6
            silkBasePosition.set(0.2, 0.8, -2)
        } else if (isTablet) {
            silkBaseRotation.x = Math.PI / 3
            silkBaseRotation.y = 0.02
            silkBaseRotation.z = Math.PI / 2.9
            silkBasePosition.set(0.6, 0.1, 0.6)
        } else {
            silkBaseRotation.x = Math.PI / 3.4
            silkBaseRotation.y = -0.08
            silkBaseRotation.z = Math.PI / 2.45
            silkBasePosition.set(1.2, -0.4, 1.8)
        }

        ribbonMesh.rotation.set(silkBaseRotation.x, silkBaseRotation.y, silkBaseRotation.z)
        ribbonMesh.position.copy(silkBasePosition)
    }
    updateMeshPosition()
    scene.add(ribbonMesh)

    // 柔和顶光
    const ambientLight = new THREE.AmbientLight(0xfdf8ef, 0.85)
    scene.add(ambientLight)

    // 主轮廓光 - 暖调
    const mainLight = new THREE.DirectionalLight(0xfff1dc, 2.2)
    mainLight.position.set(12, 14, 8)
    scene.add(mainLight)

    // 冷色补光
    const fillLight = new THREE.DirectionalLight(0xdfe9ff, 1.2)
    fillLight.position.set(-6, -8, 4)
    scene.add(fillLight)

    // 背光提升质感
    const backLight = new THREE.SpotLight(0xf3e6ff, 2.5)
    backLight.position.set(0, 10, -6)
    backLight.lookAt(0, 0, 0)
    scene.add(backLight)

    resizeHandler = () => {
        if (!renderer || !camera || !silkContainer.value) {
            return
        }
        const bounds = silkContainer.value.getBoundingClientRect()
        const newWidth = bounds.width || window.innerWidth
        const newHeight = bounds.height || window.innerHeight
        camera.aspect = newWidth / newHeight
        camera.updateProjectionMatrix()
        renderer.setSize(newWidth, newHeight)
        updateMeshPosition()
    }

    window.addEventListener('resize', resizeHandler)
    animateSilk()
}

const initAnimations = () => {
    ctx = gsap.context(() => {
        const tl = gsap.timeline({ defaults: { ease: 'power3.out' } })

        tl.fromTo(
            '.hero-sub',
            { y: 32, opacity: 0 },
            { y: 0, opacity: 1, duration: 1.2, delay: 0.25 }
        )
            .to(
                '#hero-text-1',
                {
                    opacity: 1,
                    y: 0,
                    duration: 1.4,
                    skewX: -8,
                    filter: 'blur(0px)',
                },
                '-=0.9'
            )
            .from(
                '#hero-text-1',
                {
                    y: 110,
                    filter: 'blur(16px)',
                },
                '<'
            )
            .to(
                '#hero-text-2',
                {
                    opacity: 1,
                    y: 0,
                    duration: 1.4,
                    skewX: -8,
                    filter: 'blur(0px)',
                },
                '-=1.2'
            )
            .from(
                '#hero-text-2',
                {
                    y: 140,
                    filter: 'blur(18px)',
                },
                '<'
            )
            .from(
                '.hero-cta',
                {
                    opacity: 0,
                    y: 20,
                    duration: 0.9,
                },
                '-=0.6'
            )
            .from(
                '.hero-pill',
                {
                    opacity: 0,
                    y: 24,
                    stagger: 0.08,
                    duration: 0.8,
                },
                '-=0.6'
            )
            .from(
                '.hero-note',
                {
                    opacity: 0,
                    y: 40,
                    duration: 1,
                },
                '-=0.5'
            )

        gsap.to('.hero-backdrop__grid', {
            scrollTrigger: {
                trigger: 'header',
                start: 'top top',
                end: 'bottom top',
                scrub: true,
            },
            yPercent: 18,
            scale: 1.08,
        })

        gsap.utils.toArray<HTMLElement>('.project-item').forEach((item) => {
            gsap.from(item, {
                scrollTrigger: {
                    trigger: item,
                    start: 'top 90%',
                    toggleActions: 'play none none reverse',
                },
                y: 70,
                opacity: 0,
                duration: 1.2,
                ease: 'power2.out',
            })
        })

        gsap.utils.toArray<HTMLElement>('.project-text-reveal').forEach((el) => {
            gsap.fromTo(
                el,
                { opacity: 0, letterSpacing: '0.4em' },
                {
                    opacity: 1,
                    letterSpacing: '0.15em',
                    duration: 1.4,
                    ease: 'power2.out',
                    scrollTrigger: {
                        trigger: el,
                        start: 'top 80%',
                        toggleActions: 'play none none reverse',
                    },
                }
            )
        })

        ScrollTrigger.create({
            trigger: '.hero-section',
            start: 'top top',
            end: 'bottom top',
            onUpdate: (self) => {
                heroScrollProgress.target = self.progress
            },
        })
    })
}

onMounted(() => {
    if (typeof window !== 'undefined' && window.matchMedia('(pointer: fine)').matches) {
        initCursor()
    }

    if (typeof window !== 'undefined') {
        silkPointerHandler = (event: PointerEvent) => {
            updatePointerTargets(event.clientX, event.clientY)
        }
        window.addEventListener('pointermove', silkPointerHandler, { passive: true })
    }

    evaluateSilkFallback()

    if (typeof window !== 'undefined' && 'matchMedia' in window) {
        motionQuery = window.matchMedia('(prefers-reduced-motion: reduce)')
        syncMotionPreference(motionQuery.matches)
        motionChangeHandler = (event: MediaQueryListEvent) => {
            syncMotionPreference(event.matches)
        }
        motionQuery.addEventListener('change', motionChangeHandler)
    } else {
        syncMotionPreference(false)
    }

    if (typeof navigator !== 'undefined') {
        const nav = navigator as NavigatorWithConnection
        networkInfo = nav.connection ?? null
        if (networkInfo?.addEventListener) {
            networkChangeHandler = () => {
                evaluateSilkFallback()
            }
            networkInfo.addEventListener('change', networkChangeHandler)
        }
    }

    initAnimations()

    if (typeof window !== 'undefined') {
        scrollHandler = () => {
            isNavCompacted.value = window.scrollY > 30
        }
        window.addEventListener('scroll', scrollHandler, { passive: true })
    }
})

onBeforeUnmount(() => {
    if (moveHandler) {
        window.removeEventListener('mousemove', moveHandler)
    }

    if (typeof window !== 'undefined' && silkPointerHandler) {
        window.removeEventListener('pointermove', silkPointerHandler)
    }
    silkPointerHandler = null

    if (typeof window !== 'undefined' && scrollHandler) {
        window.removeEventListener('scroll', scrollHandler)
    }
    scrollHandler = null

    if (motionQuery && motionChangeHandler) {
        motionQuery.removeEventListener('change', motionChangeHandler)
    }
    motionQuery = null
    motionChangeHandler = null

    hoverBindings.forEach(({ element, enter, leave }) => {
        element.removeEventListener('mouseenter', enter)
        element.removeEventListener('mouseleave', leave)
    })
    hoverBindings.length = 0

    if (networkInfo && networkChangeHandler && networkInfo.removeEventListener) {
        networkInfo.removeEventListener('change', networkChangeHandler)
    }
    networkInfo = null
    networkChangeHandler = null

    document.body.classList.remove('has-custom-cursor')
    if (cursorDotRef) {
        cursorDotRef.style.opacity = '0'
    }
    if (cursorOutlineRef) {
        cursorOutlineRef.style.opacity = '0'
    }

    disposeSilk()
    ctx?.revert()
    ScrollTrigger.killAll()
})
</script>

<template>
    <div class="site-root min-h-screen bg-atelier text-charcoal">
        <div class="cursor-dot" aria-hidden="true"></div>
        <div class="cursor-outline" aria-hidden="true"></div>

        <nav :class="[
            'lens-nav fixed top-0 w-full z-50 px-6 md:px-10 flex justify-between items-center transition-all duration-300',
            isNavCompacted ? 'py-3 lens-nav--compact' : 'py-5'
        ]">
            <div class="lens-nav__brand nav-link">
                NOIR & ÉCLAT
                <span>WHITE PHANTOM</span>
            </div>
            <div class="lens-nav__links">
                <a href="#gallery" class="nav-link">Gallery</a>
                <a href="#atelier" class="nav-link">Atelier</a>
                <a href="#couture" class="nav-link">Couture</a>
                <a href="#contact" class="nav-link">Contact</a>
            </div>
            <button class="nav-ghost nav-link" type="button">预约私享厅</button>
        </nav>

        <header class="hero-section pt-28 md:pt-32 relative overflow-hidden" aria-labelledby="hero-title">
            <div class="hero-backdrop absolute inset-0" id="hero-bg" aria-hidden="true">
                <div class="hero-backdrop__wash hero-deco"></div>
                <div class="hero-backdrop__grid hero-deco"></div>
                <div class="hero-backdrop__halo hero-deco"></div>
            </div>

            <div class="hero-shell page-shell">
                <div class="hero-grid">
                    <div class="hero-left">
                        <p class="hero-sub eyebrow opacity-0 translate-y-4">
                            THE GALLERY · WHITE PHANTOM
                        </p>
                        <h1 class="hero-title" id="hero-title">
                            <span id="hero-text-1">WHITE</span>
                            <span id="hero-text-2">PHANTOM</span>
                        </h1>
                        <p class="hero-lede">
                            以白中之白的光感重塑晚礼服展厅：玻璃雾面、珠光丝绸与线性 UI
                            律动交织，宛如步入被自然光包裹的 Vernissage 预展。
                            每一次滚动，都解锁不同材质的光亮与肌理。
                        </p>
                        <div class="hero-cta opacity-0">
                            <button class="hero-button hero-button--primary nav-link" type="button">
                                预约预展
                            </button>
                            <button class="hero-button hero-button--secondary nav-link" type="button">
                                下载 Lookbook
                            </button>
                        </div>
                        <div class="hero-pill-group">
                            <div class="hero-pill">
                                <span class="hero-pill__label">atelier cadence</span>
                                <span class="hero-pill__value">900 小时 / 礼服</span>
                            </div>
                            <div class="hero-pill">
                                <span class="hero-pill__label">materials</span>
                                <span class="hero-pill__value">塔夫绸 · 真丝欧根纱</span>
                            </div>
                            <div class="hero-pill">
                                <span class="hero-pill__label">global salons</span>
                                <span class="hero-pill__value">14 城巡回</span>
                            </div>
                        </div>
                    </div>

                    <div class="hero-right">
                        <div class="hero-note project-item">
                            <div class="hero-note__badge">THE GALLERY</div>
                            <p>
                                通过 Three.js 绘制的白色丝绸在视野中缓慢弯折，模仿厚重面料的
                                菲涅尔色移。GSAP 则负责文字与微交互动效，让 Linear 风格的界面更
                                像一场展览流程表。
                            </p>
                            <div class="hero-note__meta">
                                <span>VOLUME II</span>
                                <span>WHITE PHANTOM</span>
                                <span>2025</span>
                            </div>
                        </div>
                        <div class="hero-visual relative mt-10 lg:mt-12 min-h-[320px] lg:min-h-[520px]">
                            <div class="hero-visual__halo" aria-hidden="true"></div>
                            <div class="hero-visual__frame" aria-hidden="true"></div>
                            <div class="hero-visual__grain" aria-hidden="true"></div>
                            <div class="hero-silk-shell" aria-hidden="true">
                                <div ref="silkContainer" class="hero-silk-canvas"></div>
                                <div class="hero-silk-fallback" v-if="isReducedMotion || shouldUseStaticSilk">
                                    <div class="hero-silk-fallback__glow"></div>
                                </div>
                            </div>
                            <div class="hero-scroll project-text-reveal">
                                <span>Scroll</span>
                                <span>White Phantom</span>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </header>

        <main class="bg-atelier">
            <section id="gallery" class="section-block gallery-section">
                <div class="page-shell">
                    <div class="section-heading">
                        <div>
                            <p class="eyebrow">THE GALLERY</p>
                            <h2 class="text-2xl md:text-4xl font-serif">Vernissage · 白晕展柜</h2>
                        </div>
                        <p class="section-heading__lede">
                            非对称 Bento Grid 将礼服、高定配饰与材质特写拆分展示。
                            每张卡片以 1px 线性边框悬浮在“工坊迷雾”上，呈现高调摄影的高光。
                        </p>
                    </div>

                    <div class="bento-grid">
                        <article class="bento-card bento-card--xl project-item">
                            <div class="bento-card__media" aria-hidden="true">
                                <img
                                    src="https://images.unsplash.com/photo-1524504388940-b1c1722653e1?q=80&w=1980&auto=format&fit=crop"
                                    alt="Architected halo gown" loading="lazy" />
                            </div>
                            <div class="bento-card__meta">
                                <p class="bento-card__caption">LOOK 01 · STRUCTURE</p>
                                <h3 class="bento-card__title">Architected Halo</h3>
                                <p class="bento-card__caption">3D Corsetry · Hidden boning</p>
                            </div>
                        </article>

                        <article class="bento-card bento-card--tall project-item">
                            <div class="bento-card__media" aria-hidden="true">
                                <img
                                    src="https://images.unsplash.com/photo-1429257413823-8a01dd0e9701?q=80&w=1200&auto=format&fit=crop"
                                    alt="Hand pleated silk" loading="lazy" />
                            </div>
                            <div class="bento-card__meta">
                                <p class="bento-card__caption">MATERIAL LAB</p>
                                <h3 class="bento-card__title">Hand Pleated Silk</h3>
                                <p class="bento-card__caption">Perforated voile</p>
                            </div>
                        </article>

                        <article class="bento-card bento-card--wide project-item">
                            <div class="bento-card__meta">
                                <p class="bento-card__caption">VERNISSAGE NOTES</p>
                                <h3 class="bento-card__title">Gallery Flow</h3>
                                <p class="text-sm text-stone leading-relaxed">
                                    入口即见中央主礼服，右侧为材质试验室，左翼陈列编辑精选 Lookbook。
                                    全场音景与灯光均以 72 分贝以内的呼吸节奏铺陈。
                                </p>
                                <span class="bento-card__tag nav-link">LINEAR LUXURY</span>
                            </div>
                        </article>

                        <article class="bento-card project-item">
                            <div class="bento-card__media" aria-hidden="true">
                                <img
                                    src="https://images.unsplash.com/photo-1518544801958-efcbf8a7ec10?q=80&w=1200&auto=format&fit=crop"
                                    alt="Pearl embroidery" loading="lazy" />
                            </div>
                            <div class="bento-card__meta">
                                <p class="bento-card__caption">DETAIL</p>
                                <h3 class="bento-card__title">Paillette Clouds</h3>
                            </div>
                        </article>
                    </div>
                </div>
            </section>

            <section id="atelier" class="section-block atelier-section">
                <div class="page-shell">
                    <div class="section-heading">
                        <div>
                            <p class="eyebrow">THE ATELIER</p>
                            <h2 class="text-2xl md:text-4xl font-serif">工坊流程 · Linear Precision</h2>
                        </div>
                        <p class="section-heading__lede">
                            以界面化的方式呈现高定流程，让客户像阅读产品规格一样理解面料克重、缝制工序与排期。
                        </p>
                    </div>

                    <div class="atelier-grid">
                        <article class="atelier-panel project-item">
                            <h3 class="text-xl font-serif mb-6">Routine Timeline</h3>
                            <ul class="atelier-list">
                                <li>
                                    <strong>DAY 01 · PATTERN SKETCH</strong>
                                    <span>72h 线稿与立体裁剪，使用透明塔夫绸校准版型</span>
                                </li>
                                <li>
                                    <strong>DAY 04 · MATERIAL CURATION</strong>
                                    <span>香槟白塔夫绸 × 雾面珠光纱，记录温湿度与肌理</span>
                                </li>
                                <li>
                                    <strong>DAY 09 · FITTING</strong>
                                    <span>三轮试衣 + 激光刺绣定位，保留每次调整的 Data Log</span>
                                </li>
                            </ul>
                            <div class="atelier-meta">
                                <div class="atelier-meta__item">透光测试</div>
                                <div class="atelier-meta__item">香调匹配</div>
                                <div class="atelier-meta__item">触感档案</div>
                            </div>
                        </article>

                        <article class="atelier-panel project-item">
                            <h3 class="text-xl font-serif mb-4">Material Lab</h3>
                            <p class="text-sm text-stone leading-relaxed">
                                采用高调摄影的曝光策略记录每次用料，
                                并以数字孪生方式同步到 WebGL 丝绸的着色器参数，确保线上线下色温一致。
                            </p>
                            <div class="hero-pill-group mt-6">
                                <div class="hero-pill">
                                    <span class="hero-pill__label">reflectance</span>
                                    <span class="hero-pill__value">0.78 · pearl</span>
                                </div>
                                <div class="hero-pill">
                                    <span class="hero-pill__label">roughness</span>
                                    <span class="hero-pill__value">0.32 · satin</span>
                                </div>
                            </div>
                            <div class="hero-pill-group mt-4">
                                <div class="hero-pill">
                                    <span class="hero-pill__label">color drift</span>
                                    <span class="hero-pill__value">冷白 ↔ 暖金</span>
                                </div>
                                <div class="hero-pill">
                                    <span class="hero-pill__label">wind profile</span>
                                    <span class="hero-pill__value">Perlin 0.6</span>
                                </div>
                            </div>
                        </article>
                    </div>
                </div>
            </section>

            <section id="couture" class="section-block couture-section">
                <div class="page-shell">
                    <div class="section-heading">
                        <div>
                            <p class="eyebrow">COUTURE &amp; CUSTOM</p>
                            <h2 class="text-2xl md:text-4xl font-serif">预约 White Phantom 试穿</h2>
                        </div>
                    </div>
                    <div class="couture-grid">
                        <div class="couture-text project-text-reveal">
                            <p>
                                每一次量体都在磨砂玻璃围合的静谧空间进行，Ambient 光模拟拂晓日光。
                                造型顾问会根据肤色与场合输出 Look Sheet，并同步香氛气味与面料纹理档案。
                            </p>
                            <ul>
                                <li>远程体感会议 · 45 分钟</li>
                                <li>面料触感档案 + 香氛礼包寄送</li>
                                <li>72 小时内交付草图，14 天完成首版试衣</li>
                            </ul>
                        </div>
                        <div class="couture-panel project-item" id="contact">
                            <div class="couture-panel__badge">By Appointment Only</div>
                            <h3 class="text-xl font-serif mb-3">预约专属试穿</h3>
                            <p class="text-sm text-stone leading-relaxed">
                                告诉我们你的城市与理想日期，Atelier 将为你安排最近的 White Phantom Salon。
                            </p>
                            <form class="couture-form" aria-label="预约试穿">
                                <label>
                                    <span>城市</span>
                                    <input type="text" name="city" placeholder="Shanghai / Paris" />
                                </label>
                                <label>
                                    <span>日期</span>
                                    <input type="date" name="date" />
                                </label>
                                <button class="hero-button hero-button--primary w-full nav-link" type="button">
                                    提交预约
                                </button>
                            </form>
                        </div>
                    </div>
                </div>
            </section>
        </main>

        <footer class="site-footer">
            <div class="site-footer__glow" aria-hidden="true"></div>
            <p class="eyebrow">BY APPOINTMENT ONLY</p>
            <h2 class="text-3xl md:text-5xl font-serif tracking-[0.3em] mt-4">WHITE PHANTOM</h2>
            <div class="site-footer__links mt-6">
                <a href="#" class="nav-link">Instagram</a>
                <span>•</span>
                <a href="#" class="nav-link">WeChat</a>
                <span>•</span>
                <a href="#" class="nav-link">Email</a>
            </div>
            <p class="site-footer__legal">© 2025 NOIR & ÉCLAT · PARIS / SHANGHAI</p>
        </footer>
    </div>
</template>
