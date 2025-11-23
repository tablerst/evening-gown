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

const targetConfig: Omit<RibbonRuntimeConfig, 'segments' | 'width' | 'length'> = {
    speed: 0.4,
    twistSpeed: 0.1,
    twistAmplitude: 1.5,
    flowFrequency: 0.8,
    baseColor: new THREE.Color(0x04140f),
    glowColor: new THREE.Color(0x0b5d3a),
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
    segments: 400,
    width: 5,
    length: 30,
    speed: 0.4,
    twistSpeed: 0.1,
    twistAmplitude: 1.5,
    flowFrequency: 0.8,
    baseColor: new THREE.Color(0x04140f),
    glowColor: new THREE.Color(0x0b5d3a),
}

const initCursor = () => {
    cursorDotRef = document.querySelector('.cursor-dot') as HTMLElement | null
    cursorOutlineRef = document.querySelector('.cursor-outline') as HTMLElement | null

    if (!cursorDotRef || !cursorOutlineRef) {
        return
    }

    const styles = getComputedStyle(document.documentElement)
    const accentGold = styles.getPropertyValue('--accent-gold').trim() || '#D4AF37'
    const emeraldOverlay = 'rgba(5, 110, 65, 0.12)'
    const emeraldIdle = 'rgba(4, 61, 44, 0.18)'
    const idleBorder = 'rgba(232, 214, 179, 0.25)'

    document.body.classList.add('has-custom-cursor')
    cursorDotRef.style.opacity = '1'
    cursorOutlineRef.style.opacity = '1'

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
        document.querySelectorAll<HTMLElement>('a, button, textarea, .project-item, .hero-cta span, .nav-link')
    )

    interactiveElements.forEach((element) => {
        const onEnter = () => {
            cursorOutlineRef!.style.width = '50px'
            cursorOutlineRef!.style.height = '50px'
            cursorOutlineRef!.style.borderColor = accentGold
            cursorOutlineRef!.style.background = emeraldOverlay
        }

        const onLeave = () => {
            cursorOutlineRef!.style.width = '40px'
            cursorOutlineRef!.style.height = '40px'
            cursorOutlineRef!.style.borderColor = idleBorder
            cursorOutlineRef!.style.background = emeraldIdle
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

    for (let col = 0; col < verticesPerRow; col += 1) {
        const ratio = col / widthSegments
        const x = ratio * config.length - config.length / 2

        // 基础波浪
        let waveZ = Math.sin(x * 0.4 + time) * 1.2
        waveZ += Math.sin(x * 1.5 + time * 1.5) * 0.3

        // 中心线 Y 偏移
        const centerY = Math.sin(x * 0.2 + time * 0.5) * 0.8
        // 扭曲角度
        const twist = Math.sin(x * 0.3 + time * config.twistSpeed) * config.twistAmplitude

        // 颜色计算因子
        const flowPhase = ratio * 5 * config.flowFrequency - time * 2
        let glowFactor = Math.sin(flowPhase)
        glowFactor = Math.pow((glowFactor + 1) / 2, 8)
        const twistHighlight = Math.abs(Math.sin(twist))
        const mixRatio = Math.min(glowFactor * 1.5 + twistHighlight * 0.2, 1)

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

    config.speed += (targetConfig.speed - config.speed) * 0.05
    config.twistSpeed += (targetConfig.twistSpeed - config.twistSpeed) * 0.05
    config.twistAmplitude += (targetConfig.twistAmplitude - config.twistAmplitude) * 0.05
    config.flowFrequency += (targetConfig.flowFrequency - config.flowFrequency) * 0.05
    config.baseColor.lerp(targetConfig.baseColor, 0.05)
    config.glowColor.lerp(targetConfig.glowColor, 0.05)

    time += 0.01 * config.speed
    updateRibbon()

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

    camera = new THREE.PerspectiveCamera(35, width / height, 0.1, 1000)
    camera.position.set(0, 0, 20)

    renderer = new THREE.WebGLRenderer({ antialias: true, alpha: true })
    renderer.setPixelRatio(window.devicePixelRatio)
    renderer.setSize(width, height)
    renderer.setClearAlpha(0)
    renderer.toneMapping = THREE.ACESFilmicToneMapping
    renderer.toneMappingExposure = 1.2
    renderer.outputColorSpace = THREE.SRGBColorSpace
    container.appendChild(renderer.domElement)

    // 增加宽度方向的分段数，从 2 增加到 20，以解决光照伪影
    const geometry = new THREE.PlaneGeometry(config.length, config.width, config.segments, 20)
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
        emissive: 0x050510,
        metalness: 0.8,
        roughness: 0.3,
        clearcoat: 1,
        clearcoatRoughness: 0.2,
        side: THREE.DoubleSide,
        flatShading: false,
    })

    ribbonMesh = new THREE.Mesh(geometry, silkMaterial)
    ribbonMesh.rotation.z = Math.PI / 3
    ribbonMesh.rotation.x = Math.PI / 6
    ribbonMesh.position.x = 2
    scene.add(ribbonMesh)

    // 模拟月光环境
    const ambientLight = new THREE.AmbientLight(0x1c392d, 1.1)
    scene.add(ambientLight)

    // 主轮廓光
    const mainLight = new THREE.DirectionalLight(0xf6e7c8, 2.4)
    mainLight.position.set(5, 5, 5)
    scene.add(mainLight)

    // 底部补光，增加层次
    const fillLight = new THREE.DirectionalLight(0x0a3225, 1.3)
    fillLight.position.set(-5, -5, 0)
    scene.add(fillLight)

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
    }

    window.addEventListener('resize', resizeHandler)
    animateSilk()
}

const initAnimations = () => {
    ctx = gsap.context(() => {
        const tl = gsap.timeline()

        tl.to('.hero-sub', {
            opacity: 1,
            y: 0,
            duration: 1.5,
            delay: 0.5,
            ease: 'power3.out',
        })
            .to(
                '#hero-text-1',
                {
                    opacity: 1,
                    y: 0,
                    duration: 1.8,
                    skewX: -15,
                    ease: 'power4.out',
                },
                '-=1'
            )
            .from(
                '#hero-text-1',
                {
                    y: 120,
                    filter: 'blur(15px)',
                },
                '<'
            )
            .to(
                '#hero-text-2',
                {
                    opacity: 1,
                    y: 0,
                    duration: 1.8,
                    skewX: -15,
                    ease: 'power4.out',
                },
                '-=1.5'
            )
            .from(
                '#hero-text-2',
                {
                    y: 160,
                    filter: 'blur(15px)',
                },
                '<'
            )
            .to(
                '.hero-cta',
                {
                    opacity: 1,
                    y: 0,
                    duration: 1,
                },
                '-=0.5'
            )
            .to(
                '.hero-deco',
                {
                    opacity: 1,
                    duration: 1,
                },
                '<'
            )

        gsap.to('#hero-bg', {
            scrollTrigger: {
                trigger: 'header',
                start: 'top top',
                end: 'bottom top',
                scrub: true,
            },
            yPercent: 15,
            scale: 1.05,
        })

        const items = document.querySelectorAll<HTMLElement>('.project-item')
        items.forEach((item, index) => {
            gsap.from(item, {
                scrollTrigger: {
                    trigger: item,
                    start: 'top 90%',
                    toggleActions: 'play none none reverse',
                },
                y: 80,
                opacity: 0,
                duration: 1.5,
                ease: 'power2.out',
                delay: index * 0.15,
            })
        })

        gsap.to('.project-text-reveal', {
            scrollTrigger: {
                trigger: '.project-text-reveal',
                start: 'top 75%',
            },
            opacity: 1,
            letterSpacing: '0.15em',
            duration: 2.5,
            ease: 'power2.out',
        })
    })
}

onMounted(() => {
    if (typeof window !== 'undefined' && window.matchMedia('(pointer: fine)').matches) {
        initCursor()
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
            isNavCompacted.value = window.scrollY > 40
        }
        window.addEventListener('scroll', scrollHandler, { passive: true })
    }
})

onBeforeUnmount(() => {
    if (moveHandler) {
        window.removeEventListener('mousemove', moveHandler)
    }

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
    <div class="min-h-screen bg-obsidian text-platinum">
        <div class="cursor-dot" aria-hidden="true"></div>
        <div class="cursor-outline" aria-hidden="true"></div>

        <nav :class="[
            'nav-glass fixed top-0 w-full z-50 px-6 md:px-12 flex justify-between items-center transition-all duration-500',
            isNavCompacted ? 'py-3 nav-glass--compact' : 'py-5'
        ]">
            <div
                class="text-lg md:text-xl font-serif text-champagne tracking-[0.25em] font-bold flex items-center gap-3 nav-link">
                <span class="text-accent-gold text-base">✦</span>
                NOIR & ÉCLAT
            </div>
            <div class="hidden md:flex space-x-10 text-[11px] tracking-[0.35em] text-muted">
                <a href="#featured" class="hover:text-champagne transition-colors duration-300 nav-link">COLLECTIONS</a>
                <a href="#best" class="hover:text-champagne transition-colors duration-300 nav-link">BEST SELLERS</a>
                <a href="#couture" class="hover:text-champagne transition-colors duration-300 nav-link">COUTURE</a>
                <a href="#contact" class="hover:text-champagne transition-colors duration-300 nav-link">CONTACT</a>
            </div>
            <button
                class="nav-cta border border-accent-gold text-accent-gold px-5 md:px-7 py-2 text-[10px] tracking-[0.35em] uppercase hover:bg-accent-gold hover:text-obsidian transition-all duration-500 nav-link">
                PRIVATE VIEWING
            </button>
        </nav>

        <header class="hero-section relative overflow-hidden bg-surface" aria-labelledby="hero-title">
            <div class="hero-backdrop absolute inset-0" id="hero-bg" aria-hidden="true">
                <div class="hero-backdrop__veil"></div>
                <div class="hero-backdrop__gradient"></div>
                <div class="hero-backdrop__noise"></div>
            </div>

            <div class="hero-shell page-shell">
                <div class="hero-grid relative z-10 grid gap-12 lg:gap-16">
                    <div class="hero-left col-span-12 lg:col-span-5">
                        <div class="hero-left__veil" aria-hidden="true"></div>
                        <p class="hero-eyebrow hero-sub opacity-0">
                            THE COLLECTION · <span class="text-accent-gold">THE EMERALD NIGHT</span>
                        </p>
                        <h1 class="hero-title" id="hero-title">
                            <span id="hero-text-1">OBSIDIAN</span>
                            <span id="hero-text-2">DREAMS</span>
                        </h1>
                        <p class="hero-lede text-body-copy">
                            黑曜石般的背景衬托祖母绿的丝绸流光，
                            低声讲述 Atelier 手工缝制的仪式感。14 道工序、900 小时，让夜色在你指尖流动。
                        </p>
                        <div class="hero-cta flex flex-col sm:flex-row gap-4 mt-12 opacity-0">
                            <button class="hero-button hero-button--primary nav-link" type="button">
                                探索系列
                            </button>
                            <button class="hero-button hero-button--secondary nav-link" type="button">
                                预约试穿
                            </button>
                        </div>
                        <div class="hero-pill-group">
                            <div class="hero-pill">
                                <span class="hero-pill__label">COUTURE SALON</span>
                                <span class="hero-pill__value">14 ateliers worldwide</span>
                            </div>
                            <div class="hero-pill">
                                <span class="hero-pill__label">HAND-CRAFTED</span>
                                <span class="hero-pill__value">900 hours / gown</span>
                            </div>
                        </div>
                    </div>

                    <div class="hero-middle col-span-12 lg:col-span-3">
                        <div class="hero-card">
                            <p class="hero-card__title">SILK LUMINESCENCE</p>
                            <p class="hero-card__body">
                                单条丝绸缓慢呼吸，S 形弧线从右上垂落到左下，
                                高光保持克制，仿佛剧场灯光在布料上游走。
                            </p>
                            <div class="hero-card__meta">
                                <span>12s breathing cycle</span>
                                <span>WCAG AA</span>
                            </div>
                        </div>
                        <div class="hero-divider"></div>
                        <div class="hero-note">
                            <span class="hero-note__eyebrow">APPOINTMENT</span>
                            <p>72 小时内专属顾问回复，
                                <span class="text-accent-gold">定制流程一次完成。</span>
                            </p>
                        </div>
                    </div>

                    <div class="hero-right col-span-12 lg:col-span-4">
                        <div class="hero-silk-shell" aria-hidden="true">
                            <div ref="silkContainer" class="hero-silk-canvas"></div>
                            <div class="hero-silk-fallback" v-if="isReducedMotion || shouldUseStaticSilk">
                                <div class="hero-silk-fallback__glow"></div>
                            </div>
                        </div>
                        <div class="hero-right__caption">
                            <span>VOL. II · EMERALD NIGHT</span>
                            <span>Scroll</span>
                        </div>
                    </div>
                </div>
            </div>
        </header>

        <main class="bg-obsidian">
            <section id="featured" class="section-block">
                <div class="page-shell">
                    <div class="section-heading">
                        <div class="section-heading__line"></div>
                        <div>
                            <p class="eyebrow">FEATURED COLLECTION</p>
                            <h2>精选系列 · Emerald Reverie</h2>
                        </div>
                        <p class="section-heading__lede">
                            12 列栅格下的丝绒留白，让礼服成为空间主角。
                            每张卡片都以 3:4 比例呈现裙摆垂坠感。
                        </p>
                    </div>

                    <div class="collection-grid">
                        <article class="collection-card project-item">
                            <div class="collection-card__media">
                                <img src="https://images.unsplash.com/photo-1595777457583-95e059d581b8?q=80&w=1983&auto=format&fit=crop"
                                    alt="Moonlit Velvet" loading="lazy" />
                            </div>
                            <div class="collection-card__content">
                                <h3>Moonlit Velvet</h3>
                                <p>Silver Thread Atelier · 900 Hours</p>
                            </div>
                        </article>

                        <article class="collection-card project-item">
                            <div class="collection-card__media">
                                <img src="https://images.unsplash.com/photo-1566174053879-31528523f8ae?q=80&w=1983&auto=format&fit=crop"
                                    alt="Nebula Gown" loading="lazy" />
                            </div>
                            <div class="collection-card__content">
                                <h3>Nebula Gown</h3>
                                <p>Deep Purple Chiffon · Atelier 04</p>
                            </div>
                        </article>

                        <article class="collection-card collection-card--highlight project-item">
                            <div class="collection-card__media">
                                <img src="https://images.unsplash.com/photo-1539008835657-9e8e9680c956?q=80&w=1887&auto=format&fit=crop"
                                    alt="Ethereal" loading="lazy" />
                                <div class="collection-card__overlay">
                                    <span class="eyebrow">ETHEREAL</span>
                                </div>
                            </div>
                            <div class="collection-card__content">
                                <h3>The Royal Silhouette</h3>
                                <p>Obsidian Tulle · Limited 12</p>
                                <button class="tag-chip nav-link" type="button">INQUIRE</button>
                            </div>
                        </article>
                    </div>
                </div>
            </section>

            <section id="best" class="section-block section-block--dense">
                <div class="page-shell">
                    <div class="section-heading section-heading--compact">
                        <p class="eyebrow">BEST SELLERS</p>
                        <h2>热门单品 · Best Sellers</h2>
                    </div>
                    <div class="best-grid">
                        <article class="best-card">
                            <div class="best-card__media">
                                <img src="https://images.unsplash.com/photo-1514996937319-344454492b37?q=80&w=900&auto=format&fit=crop"
                                    alt="Asteria" loading="lazy" />
                            </div>
                            <div class="best-card__body">
                                <div>
                                    <h3>Asteria Veil</h3>
                                    <p>Hand-beaded constellation lace</p>
                                </div>
                                <div class="best-card__footer">
                                    <span class="price">¥ 128,000</span>
                                    <button class="hero-button hero-button--ghost nav-link" type="button">加入心愿单</button>
                                </div>
                            </div>
                        </article>

                        <article class="best-card">
                            <div class="best-card__media">
                                <img src="https://images.unsplash.com/photo-1475180098004-ca77a66827be?q=80&w=900&auto=format&fit=crop"
                                    alt="Seraphine" loading="lazy" />
                            </div>
                            <div class="best-card__body">
                                <div>
                                    <h3>Seraphine Column</h3>
                                    <p>Bias-cut satin · Atelier Milano</p>
                                </div>
                                <div class="best-card__footer">
                                    <span class="price">¥ 96,000</span>
                                    <button class="hero-button hero-button--ghost nav-link" type="button">预约试穿</button>
                                </div>
                            </div>
                        </article>

                        <article class="best-card">
                            <div class="best-card__media">
                                <img src="https://images.unsplash.com/photo-1521572163474-6864f9cf17ab?q=80&w=900&auto=format&fit=crop"
                                    alt="Nocturne" loading="lazy" />
                            </div>
                            <div class="best-card__body">
                                <div>
                                    <h3>Nocturne Cape</h3>
                                    <p>Velvet gradient · Detachable cape</p>
                                </div>
                                <div class="best-card__footer">
                                    <span class="price">¥ 138,000</span>
                                    <button class="hero-button hero-button--ghost nav-link" type="button">查看细节</button>
                                </div>
                            </div>
                        </article>
                    </div>
                </div>
            </section>

            <section id="couture" class="section-block section-couture">
                <div class="page-shell">
                    <div class="section-heading">
                        <div class="section-heading__line"></div>
                        <div>
                            <p class="eyebrow">COUTURE & CUSTOM</p>
                            <h2>定制服务 · Couture & Custom</h2>
                        </div>
                    </div>
                    <div class="couture-grid">
                        <div class="couture-text">
                            <p>
                                Atelier 团队根据体温、肤色与出席场合定制配色，
                                每次量体都在玻璃拟态的静谧空间完成。
                            </p>
                            <ul>
                                <li>一对一造型顾问 · 线下 / 远程</li>
                                <li>丝绒面料触感档案 &amp; 香氛礼包</li>
                                <li>72 小时出具草图，14 天完成初版试衣</li>
                            </ul>
                        </div>
                        <div class="couture-panel" id="contact">
                            <div class="couture-panel__badge">By Appointment Only</div>
                            <h3>预约专属试穿</h3>
                            <p>留下你的城市与日期，我们将安排最近的 Emerald Night Salon。</p>
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
            <h2 class="skew-title text-4xl md:text-6xl">YOUR LEGACY</h2>
            <div class="site-footer__links">
                <a href="#" class="nav-link">Instagram</a>
                <span class="text-accent-gold">•</span>
                <a href="#" class="nav-link">WeChat</a>
                <span class="text-accent-gold">•</span>
                <a href="#" class="nav-link">Email</a>
            </div>
            <p class="site-footer__legal">© 2025 NOIR & ÉCLAT · PARIS / SHANGHAI</p>
        </footer>
    </div>
</template>
