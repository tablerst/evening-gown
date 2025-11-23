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

let moveHandler: ((event: MouseEvent) => void) | null = null
const hoverBindings: HoverBinding[] = []
let ctx: gsap.Context | null = null
let cursorDotRef: HTMLElement | null = null
let cursorOutlineRef: HTMLElement | null = null

const silkContainer = ref<HTMLDivElement | null>(null)
const moodInput = ref('')
const aiResponse = ref('')
const isWeaving = ref(false)

type RibbonParams = {
    baseColor: string
    glowColor: string
    speed: number
    twistSpeed: number
    twistAmplitude: number
    flowFrequency: number
}

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

type GeminiResult = {
    params: RibbonParams
    poetic_desc: string
}

let scene: THREE.Scene | null = null
let camera: THREE.PerspectiveCamera | null = null
let renderer: THREE.WebGLRenderer | null = null
let ribbonMesh: THREE.Mesh<THREE.PlaneGeometry, THREE.MeshPhysicalMaterial> | null = null
let silkMaterial: THREE.MeshPhysicalMaterial | null = null
let animationFrameId: number | null = null
let resizeHandler: (() => void) | null = null
let time = 0

const targetConfig: Omit<RibbonRuntimeConfig, 'segments' | 'width' | 'length'> = {
    speed: 0.6,
    twistSpeed: 0.2,
    twistAmplitude: 1,
    flowFrequency: 1,
    baseColor: new THREE.Color(0x030305),
    glowColor: new THREE.Color(0x4b0082),
}

const config: RibbonRuntimeConfig = {
    segments: 400,
    width: 4,
    length: 25,
    speed: 0.6,
    twistSpeed: 0.2,
    twistAmplitude: 1,
    flowFrequency: 1,
    baseColor: new THREE.Color(0x030305),
    glowColor: new THREE.Color(0x4b0082),
}

const GEMINI_API_URL =
    'https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash-preview-09-2025:generateContent'
const GEMINI_API_KEY = import.meta.env.VITE_GEMINI_API_KEY

const initCursor = () => {
    cursorDotRef = document.querySelector('.cursor-dot') as HTMLElement | null
    cursorOutlineRef = document.querySelector('.cursor-outline') as HTMLElement | null

    if (!cursorDotRef || !cursorOutlineRef) {
        return
    }

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
            cursorOutlineRef!.style.borderColor = '#D4AF37'
            cursorOutlineRef!.style.background = 'rgba(75, 0, 130, 0.1)'
        }

        const onLeave = () => {
            cursorOutlineRef!.style.width = '40px'
            cursorOutlineRef!.style.height = '40px'
            cursorOutlineRef!.style.borderColor = 'rgba(255, 255, 255, 0.2)'
            cursorOutlineRef!.style.background = 'rgba(75, 0, 130, 0.15)'
        }

        element.addEventListener('mouseenter', onEnter)
        element.addEventListener('mouseleave', onLeave)

        hoverBindings.push({ element, enter: onEnter, leave: onLeave })
    })
}

const clamp = (value: number, min: number, max: number) => Math.min(max, Math.max(min, value))

const fallbackPresets: [GeminiResult, ...GeminiResult[]] = [
    {
        params: {
            baseColor: '#030308',
            glowColor: '#3f8cff',
            speed: 0.7,
            twistSpeed: 0.35,
            twistAmplitude: 1.1,
            flowFrequency: 1.3,
        },
        poetic_desc: '寒蓝光线在黑幕上折返，似深夜霓虹轻抚暗绸。',
    },
    {
        params: {
            baseColor: '#04010a',
            glowColor: '#a855f7',
            speed: 0.5,
            twistSpeed: 0.22,
            twistAmplitude: 1.4,
            flowFrequency: 0.9,
        },
        poetic_desc: '暮紫缠绕的丝束缓慢舒展，若舞台帷幕悄然苏醒。',
    },
    {
        params: {
            baseColor: '#050c12',
            glowColor: '#4de4c9',
            speed: 1.1,
            twistSpeed: 0.6,
            twistAmplitude: 1.8,
            flowFrequency: 2,
        },
        poetic_desc: '深海色的涟漪被青绿电光点亮，倏忽如星河。',
    },
]

const getFallbackResponse = (prompt: string): GeminiResult => {
    if (!prompt) {
        return fallbackPresets[0]
    }
    const index = Math.abs(prompt.length) % fallbackPresets.length
    return fallbackPresets[index]!
}

const applyAIParams = (params: RibbonParams) => {
    try {
        if (params.baseColor) {
            targetConfig.baseColor = new THREE.Color(params.baseColor)
        }
        if (params.glowColor) {
            targetConfig.glowColor = new THREE.Color(params.glowColor)
        }
    } catch (error) {
        console.warn('Invalid color from parameters', error)
    }

    targetConfig.speed = clamp(params.speed ?? targetConfig.speed, 0.2, 2)
    targetConfig.twistSpeed = clamp(params.twistSpeed ?? targetConfig.twistSpeed, 0.1, 1.5)
    targetConfig.twistAmplitude = clamp(params.twistAmplitude ?? targetConfig.twistAmplitude, 0.5, 2.5)
    targetConfig.flowFrequency = clamp(params.flowFrequency ?? targetConfig.flowFrequency, 0.5, 3)
}

const updateRibbon = () => {
    if (!ribbonMesh) {
        return
    }

    const geometry = ribbonMesh.geometry
    const positions = geometry.attributes.position as THREE.BufferAttribute
    const colors = geometry.attributes.color as THREE.BufferAttribute

    const widthSegments = config.segments
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

        let waveZ = Math.sin(x * 0.4 + time) * 1.2
        waveZ += Math.sin(x * 1.5 + time * 1.5) * 0.3

        const centerY = Math.sin(x * 0.2 + time * 0.5) * 0.8
        const twist = Math.sin(x * 0.3 + time * config.twistSpeed) * config.twistAmplitude

        const idxTop = col
        const idxBot = col + verticesPerRow
        const halfWidth = config.width / 2

        const topY = centerY + halfWidth * Math.cos(twist)
        const topZ = waveZ + halfWidth * Math.sin(twist)
        const botY = centerY - halfWidth * Math.cos(twist)
        const botZ = waveZ - halfWidth * Math.sin(twist)

        positions.setY(idxTop, topY)
        positions.setZ(idxTop, topZ)
        positions.setY(idxBot, botY)
        positions.setZ(idxBot, botZ)

        const flowPhase = ratio * 5 * config.flowFrequency - time * 2
        let glowFactor = Math.sin(flowPhase)
        glowFactor = Math.pow((glowFactor + 1) / 2, 8)

        const twistHighlight = Math.abs(Math.sin(twist))
        const mixRatio = Math.min(glowFactor * 1.5 + twistHighlight * 0.2, 1)

        const r = baseR + (glowR - baseR) * mixRatio
        const g = baseG + (glowG - baseG) * mixRatio
        const b = baseB + (glowB - baseB) * mixRatio

        colors.setXYZ(idxTop, r, g, b)
        colors.setXYZ(idxBot, r, g, b)
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
    if (!silkContainer.value || renderer) {
        return
    }

    const container = silkContainer.value
    scene = new THREE.Scene()
    scene.fog = new THREE.FogExp2(0x030305, 0.04)

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

    const geometry = new THREE.PlaneGeometry(config.length, config.width, config.segments, 2)
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
        emissive: 0x000000,
        metalness: 0.6,
        roughness: 0.2,
        clearcoat: 1,
        clearcoatRoughness: 0.1,
        side: THREE.DoubleSide,
        flatShading: false,
    })

    ribbonMesh = new THREE.Mesh(geometry, silkMaterial)
    ribbonMesh.rotation.z = Math.PI / 4
    ribbonMesh.rotation.x = Math.PI / 8
    scene.add(ribbonMesh)

    const ambientLight = new THREE.AmbientLight(0x222222, 1)
    scene.add(ambientLight)

    const rimLight = new THREE.DirectionalLight(0xffffff, 1.5)
    rimLight.position.set(0, 10, -5)
    scene.add(rimLight)

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

const callGemini = async (prompt: string): Promise<GeminiResult | null> => {
    if (!GEMINI_API_KEY) {
        console.warn('VITE_GEMINI_API_KEY is missing; using fallback presets.')
        return getFallbackResponse(prompt)
    }

    const systemPrompt = `You are a visual generative artist creating high-end, abstract silk visualizations. Translate user inputs into render parameters.

Return JSON only:
1. "params": Numerical/Color values.
2. "poetic_desc": A short, elegant Chinese sentence describing the visual.

Params:
- baseColor: Hex (e.g., "#000510").
- glowColor: Hex.
- speed: Float 0.2 to 2.0.
- twistSpeed: Float 0.1 to 1.5.
- twistAmplitude: Float 0.5 to 2.5.
- flowFrequency: Float 0.5 to 3.0.`

    const payload = {
        contents: [{ parts: [{ text: `User input: "${prompt}"` }] }],
        systemInstruction: { parts: [{ text: systemPrompt }] },
        generationConfig: { responseMimeType: 'application/json' },
    }

    try {
        const response = await fetch(`${GEMINI_API_URL}?key=${GEMINI_API_KEY}`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(payload),
        })

        if (!response.ok) {
            throw new Error(`API Error: ${response.status}`)
        }

        const data = await response.json()
        const rawText = data?.candidates?.[0]?.content?.parts?.[0]?.text
        if (!rawText) {
            throw new Error('Gemini response missing text payload')
        }

        return JSON.parse(rawText) as GeminiResult
    } catch (error) {
        console.error('Gemini API Failed', error)
        return getFallbackResponse(prompt)
    }
}

const generateRibbon = async () => {
    const prompt = moodInput.value.trim()
    if (!prompt || isWeaving.value) {
        return
    }

    isWeaving.value = true
    aiResponse.value = ''

    const result = await callGemini(prompt)

    if (result?.params) {
        applyAIParams(result.params)
        aiResponse.value = result.poetic_desc
    } else {
        aiResponse.value = '未能获取灵感参数，请稍后重试。'
    }

    isWeaving.value = false
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
            .to(
                '.mood-panel-fixed',
                {
                    opacity: 1,
                    y: 0,
                    duration: 1.2,
                    ease: 'power3.out',
                },
                '-=0.8'
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

        gsap.to('.mood-panel-fixed', {
            scrollTrigger: {
                trigger: 'header',
                start: 'top top',
                end: 'bottom top',
                scrub: true,
            },
            opacity: 0,
            y: -40,
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
    if (window.matchMedia('(pointer: fine)').matches) {
        initCursor()
    }

    initAnimations()
    nextTick(() => {
        initSilkCanvas()
    })
})

onBeforeUnmount(() => {
    if (moveHandler) {
        window.removeEventListener('mousemove', moveHandler)
    }

    hoverBindings.forEach(({ element, enter, leave }) => {
        element.removeEventListener('mouseenter', enter)
        element.removeEventListener('mouseleave', leave)
    })
    hoverBindings.length = 0

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
    <div class="min-h-screen bg-[var(--void-black)] text-[var(--text-body)]">
        <div class="cursor-dot" aria-hidden="true"></div>
        <div class="cursor-outline" aria-hidden="true"></div>

        <nav class="nav-glass fixed top-0 w-full z-50 px-6 md:px-8 py-4 flex justify-between items-center">
            <div
                class="text-lg md:text-xl font-serif text-white tracking-[0.2em] font-bold flex items-center gap-2 nav-link">
                <span class="text-[#D4AF37]">✦</span>
                NOIR & ÉCLAT
            </div>
            <div class="hidden md:flex space-x-12 text-xs tracking-[0.2em] text-muted">
                <a href="#" class="hover:text-white transition-colors duration-300 nav-link">COLLECTIONS</a>
                <a href="#" class="hover:text-white transition-colors duration-300 nav-link">RUNWAY</a>
                <a href="#" class="hover:text-white transition-colors duration-300 nav-link">ATELIER</a>
                <a href="#" class="hover:text-white transition-colors duration-300 nav-link">CONTACT</a>
            </div>
            <button
                class="border border-[#D4AF37] text-[#D4AF37] px-6 md:px-8 py-2 text-[10px] tracking-[0.3em] uppercase hover:bg-[#D4AF37] hover:text-black transition-all duration-500 nav-link">
                Private View
            </button>
        </nav>

        <header class="relative w-full h-screen flex items-center justify-center overflow-hidden bg-black">
            <div class="absolute inset-0 w-full h-[120%] -top-[10%]" id="hero-bg">
                <div ref="silkContainer" aria-hidden="true" class="absolute inset-0 pointer-events-none overflow-hidden"
                    id="silk-canvas"></div>
                <div
                    class="absolute inset-0 bg-gradient-to-br from-[rgba(25,25,112,0.35)] via-black/80 to-[rgba(75,0,130,0.25)] pointer-events-none">
                </div>
                <div
                    class="absolute inset-0 bg-gradient-to-t from-[var(--void-black)] via-transparent to-black/50 pointer-events-none">
                </div>
            </div>

            <div
                class="mood-panel-fixed hidden lg:flex flex-col gap-4 absolute top-24 right-10 w-[300px] bg-[rgba(10,10,15,0.65)] border border-white/10 rounded-2xl backdrop-blur-xl p-6 text-white shadow-[0_15px_45px_rgba(0,0,0,0.55)] opacity-0 translate-y-4 z-20">
                <div class="flex items-center gap-3 text-[11px] uppercase tracking-[0.3em] text-muted">
                    <span class="text-[#D4AF37]">✦</span>
                    Mood Weaver
                </div>
                <textarea v-model="moodInput" aria-label="输入意境" placeholder="输入意境… 例如：银河倾泻"
                    class="w-full h-24 bg-white/5 border border-white/15 rounded-lg text-sm text-[var(--text-body)] px-3 py-2 focus:outline-none focus:border-white/40 focus:bg-white/10 placeholder:text-white/30 transition-colors">
                </textarea>
                <button type="button" @click="generateRibbon" :disabled="isWeaving || !moodInput.trim()" :class="[
                    'flex items-center justify-center gap-3 text-[11px] tracking-[0.3em] uppercase border rounded-lg py-3 transition-all duration-300',
                    isWeaving || !moodInput.trim()
                        ? 'border-white/10 text-white/40 bg-white/5'
                        : 'border-white/20 text-white hover:bg-white/10'
                ]">
                    <div v-if="isWeaving"
                        class="w-3 h-3 border border-white border-t-transparent rounded-full animate-spin"></div>
                    <span>{{ isWeaving ? 'WEAVING' : 'GENERATE' }}</span>
                </button>
                <p aria-live="polite"
                    class="text-sm font-serif italic text-white/80 leading-relaxed min-h-[48px] transition-opacity duration-700"
                    :class="aiResponse ? 'opacity-100' : 'opacity-70'">
                    {{ aiResponse || '用一句意境唤醒丝绸。' }}
                </p>
            </div>

            <div class="relative z-10 text-center flex flex-col items-center px-4">
                <p class="italic-serif text-hero-sub text-base md:text-xl mb-8 tracking-[0.3em] opacity-0 hero-sub">
                    The <span class="text-[#D4AF37]">2025</span> Midnight Series
                </p>
                <h1 class="flex flex-col items-center justify-center gap-2 md:gap-4 scale-y-110">
                    <span class="skew-title text-5xl md:text-8xl lg:text-9xl tracking-tighter"
                        id="hero-text-1">OBSIDIAN</span>
                    <span class="skew-title text-5xl md:text-8xl lg:text-9xl md:ml-12 tracking-tighter"
                        id="hero-text-2">DREAMS</span>
                </h1>
                <div class="mt-20 opacity-0 hero-cta flex flex-col items-center">
                    <div class="h-16 w-[1px] bg-gradient-to-b from-[#D4AF37] to-transparent mb-4 opacity-50"></div>
                    <span class="eyebrow nav-link hover:text-white transition-colors">Discover the Essence</span>
                </div>
            </div>

            <div
                class="absolute bottom-12 left-6 md:left-12 text-muted font-serif italic text-lg hidden md:block opacity-0 hero-deco">
                Vol. II
            </div>
            <div
                class="absolute bottom-12 right-6 md:right-12 flex flex-col gap-6 hidden md:flex opacity-0 hero-deco items-center">
                <div class="gold-line-vertical h-12"></div>
                <span class="writing-vertical text-[10px] tracking-[0.5em] text-caption">SCROLL</span>
            </div>
        </header>

        <section class="relative w-full min-h-screen py-24 md:py-32 px-6 md:px-20 bg-[var(--void-black)]">
            <div class="mb-20 md:mb-32 flex flex-col md:flex-row items-start md:items-end justify-between gap-8">
                <div class="relative pl-8">
                    <div class="absolute left-0 top-2 bottom-2 w-[1px] bg-[#D4AF37]"></div>
                    <span class="eyebrow block mb-3">Masterpieces</span>
                    <h2 class="text-4xl md:text-5xl font-serif text-white">The Collection</h2>
                </div>
                <p class="text-body-copy text-sm leading-relaxed font-light md:w-1/3 md:text-right">
                    Where darkness meets luminescence.<br />
                    <span class="italic text-[#D4AF37]">Silk, velvet, and the weight of the night.</span>
                </p>
            </div>

            <div class="grid grid-cols-1 md:grid-cols-12 gap-8 md:gap-y-40">
                <div class="col-span-1 md:col-span-7 flex flex-col gap-6 group project-item">
                    <div
                        class="image-wrapper aspect-[3/4] w-full relative cursor-none-important grayscale-[30%] hover:grayscale-0 transition-all duration-700">
                        <img src="https://images.unsplash.com/photo-1595777457583-95e059d581b8?q=80&w=1983&auto=format&fit=crop"
                            alt="Moonlit Velvet" class="w-full h-full object-cover" />
                        <div
                            class="absolute inset-0 border border-white/5 group-hover:border-white/20 transition-colors duration-500">
                        </div>
                    </div>
                    <div class="flex justify-between items-start mt-2 px-2">
                        <div>
                            <h3
                                class="text-3xl font-serif italic text-white group-hover:text-[#D4AF37] transition-colors duration-500">
                                Moonlit Velvet
                            </h3>
                            <div class="flex items-center gap-3 mt-2">
                                <span class="w-8 h-px bg-gray-700"></span>
                                <p class="text-[10px] text-caption uppercase tracking-[0.3em]">Silver Thread / 900 Hours
                                </p>
                            </div>
                        </div>
                    </div>
                </div>

                <div class="col-span-1 md:col-span-4 md:col-start-9 md:mt-32 flex flex-col gap-6 group project-item">
                    <div class="image-wrapper aspect-[3/4] w-full relative">
                        <img src="https://images.unsplash.com/photo-1566174053879-31528523f8ae?q=80&w=2548&auto=format&fit=crop"
                            alt="Nebula Gown" class="w-full h-full object-cover" />
                    </div>
                    <div class="flex justify-between items-start mt-2 px-2">
                        <div>
                            <h3
                                class="text-2xl font-serif italic text-white group-hover:text-[#a5b4fc] transition-colors duration-500">
                                Nebula Gown
                            </h3>
                            <div class="flex items-center gap-3 mt-2">
                                <span class="w-8 h-px bg-gray-700"></span>
                                <p class="text-[10px] text-caption uppercase tracking-[0.3em]">Deep Purple Chiffon</p>
                            </div>
                        </div>
                    </div>
                </div>

                <div
                    class="col-span-1 md:col-span-10 md:col-start-2 mt-12 md:mt-20 flex flex-col gap-6 group project-item">
                    <div class="image-wrapper aspect-[16/9] w-full relative">
                        <img src="https://images.unsplash.com/photo-1539008835657-9e8e9680c956?q=80&w=1887&auto=format&fit=crop"
                            alt="Ethereal" class="w-full h-full object-cover object-top" />
                        <div
                            class="absolute inset-0 bg-black/50 group-hover:bg-black/20 transition-colors duration-500">
                        </div>
                        <div class="absolute inset-0 flex items-center justify-center mix-blend-overlay">
                            <h2
                                class="skew-title text-4xl md:text-8xl z-10 opacity-0 project-text-reveal text-white/90">
                                ETHEREAL</h2>
                        </div>
                    </div>
                    <div
                        class="flex flex-col md:flex-row md:items-center justify-between mt-4 px-4 border-t border-white/10 pt-4 gap-4">
                        <h3 class="text-2xl font-serif italic text-white">The Royal Silhouette</h3>
                        <span
                            class="text-[#D4AF37] text-xs tracking-[0.2em] border border-[#D4AF37] px-4 py-2 rounded-full hover:bg-[#D4AF37] hover:text-black transition-all cursor-pointer nav-link">
                            INQUIRE
                        </span>
                    </div>
                </div>
            </div>
        </section>

        <footer class="bg-black py-24 md:py-32 text-center relative overflow-hidden border-t border-white/5">
            <div
                class="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[80vw] h-[50vh] bg-[radial-gradient(ellipse_at_center,_rgba(75,0,130,0.15),_transparent_70%)] opacity-50 pointer-events-none">
            </div>
            <p class="text-[#D4AF37] text-xs tracking-[0.5em] mb-6 uppercase">By Appointment Only</p>
            <h2 class="skew-title text-4xl md:text-6xl mb-12 opacity-100" style="opacity: 1 !important;">YOUR LEGACY
            </h2>
            <div class="flex justify-center gap-8 text-[10px] tracking-[0.2em] text-muted uppercase relative z-10">
                <a href="#" class="hover:text-white transition-colors nav-link">Instagram</a>
                <span class="text-[#D4AF37]">•</span>
                <a href="#" class="hover:text-white transition-colors nav-link">WeChat</a>
                <span class="text-[#D4AF37]">•</span>
                <a href="#" class="hover:text-white transition-colors nav-link">Email</a>
            </div>
            <p class="text-caption text-[10px] mt-24 tracking-[0.5em] relative z-10">© 2025 NOIR & ÉCLAT. PARIS /
                SHANGHAI.</p>
        </footer>
    </div>
</template>
