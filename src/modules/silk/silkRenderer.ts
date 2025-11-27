import { nextTick } from 'vue'
import type { Ref } from 'vue'
import * as THREE from 'three'

export type NetworkInformationLike = {
    effectiveType?: string
    addEventListener?: (type: string, listener: () => void) => void
    removeEventListener?: (type: string, listener: () => void) => void
}

export type NavigatorWithConnection = Navigator & {
    connection?: NetworkInformationLike
}

export type HeroScrollProgress = {
    target: number
    current: number
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

type pointerInfluenceState = {
    targetX: number
    targetY: number
    currentX: number
    currentY: number
}

type BreezeState = {
    phaseX: number
    phaseY: number
    driftX: number
    driftY: number
    gustDuration: number
    gustElapsed: number
    gustStrength: number
    timeUntilNextGust: number
}

export type SilkRendererOptions = {
    containerRef: Ref<HTMLDivElement | null>
    isReducedMotion: Ref<boolean>
    shouldUseStaticSilk: Ref<boolean>
}

export type SilkRendererController = {
    heroScrollProgress: HeroScrollProgress
    evaluateFallback: () => void
    syncMotionPreference: (shouldReduce: boolean) => void
    dispose: () => void
}

export const createSilkRenderer = ({
    containerRef,
    isReducedMotion,
    shouldUseStaticSilk,
}: SilkRendererOptions): SilkRendererController => {
    const heroScrollProgress: HeroScrollProgress = {
        target: 0,
        current: 0,
    }

    const pointerInfluence: pointerInfluenceState = {
        targetX: 0,
        targetY: 0,
        currentX: 0,
        currentY: 0,
    }

    const breezeState: BreezeState = {
        phaseX: Math.random() * Math.PI * 2,
        phaseY: Math.random() * Math.PI * 2,
        driftX: THREE.MathUtils.randFloatSpread(0.6),
        driftY: THREE.MathUtils.randFloatSpread(0.6),
        gustDuration: 0,
        gustElapsed: 0,
        gustStrength: 0,
        timeUntilNextGust: 3 + Math.random() * 3,
    }

    const silkBasePosition = new THREE.Vector3(-2, -5, -2)
    const silkBaseRotation = { x: Math.PI / 4, y: 0, z: Math.PI / 2.8 }
    const silkVerticalSpin = THREE.MathUtils.degToRad(0)

    const baseSilkColor = new THREE.Color(0xf6f0e9)
    const highlightSilkColor = new THREE.Color(0xfdfaf2)
    const baseGlowColor = new THREE.Color(0xfdf3e3)
    const accentGlowColor = new THREE.Color(0xfff2d6)

    const targetConfig: Omit<RibbonRuntimeConfig, 'segments' | 'width' | 'length'> = {
        speed: 0.12,
        twistSpeed: 0.05,
        twistAmplitude: 0.65,
        flowFrequency: 0.34,
        baseColor: baseSilkColor.clone(),
        glowColor: baseGlowColor.clone(),
    }

    const config: RibbonRuntimeConfig = {
        segments: 260,
        width: 6,
        length: 32,
        speed: 0.11,
        twistSpeed: 0.05,
        twistAmplitude: 0.72,
        flowFrequency: 0.38,
        baseColor: baseSilkColor.clone(),
        glowColor: baseGlowColor.clone(),
    }

    let scene: THREE.Scene | null = null
    let camera: THREE.PerspectiveCamera | null = null
    let renderer: THREE.WebGLRenderer | null = null
    let ribbonMesh: THREE.Mesh<THREE.PlaneGeometry, THREE.MeshPhysicalMaterial> | null = null
    let silkMaterial: THREE.MeshPhysicalMaterial | null = null
    let animationFrameId: number | null = null
    let resizeHandler: (() => void) | null = null
    let time = 0
    let lastFrameTime = 0

    const resetBreezeState = () => {
        breezeState.phaseX = Math.random() * Math.PI * 2
        breezeState.phaseY = Math.random() * Math.PI * 2
        breezeState.driftX = THREE.MathUtils.randFloatSpread(0.6)
        breezeState.driftY = THREE.MathUtils.randFloatSpread(0.6)
        breezeState.gustDuration = 0
        breezeState.gustElapsed = 0
        breezeState.gustStrength = 0
        breezeState.timeUntilNextGust = 3 + Math.random() * 3
    }

    const applyAutonomousBreeze = (delta: number) => {
        breezeState.phaseX += delta * 0.32
        breezeState.phaseY += delta * 0.26

        breezeState.timeUntilNextGust -= delta
        if (breezeState.timeUntilNextGust <= 0) {
            breezeState.gustDuration = 2.2 + Math.random() * 1.8
            breezeState.gustElapsed = 0
            breezeState.gustStrength = 0.18 + Math.random() * 0.35
            breezeState.timeUntilNextGust = 4.5 + Math.random() * 3.5
            breezeState.driftX = THREE.MathUtils.randFloatSpread(0.7)
            breezeState.driftY = THREE.MathUtils.randFloatSpread(0.7)
        }

        if (breezeState.gustElapsed < breezeState.gustDuration) {
            breezeState.gustElapsed += delta
        }

        const gustProgress = breezeState.gustDuration
            ? Math.min(breezeState.gustElapsed / breezeState.gustDuration, 1)
            : 0
        const gustEnvelope = Math.sin(gustProgress * Math.PI) * breezeState.gustStrength

        const baseX = Math.sin(breezeState.phaseX) * 0.48
        const layeredX = Math.sin(breezeState.phaseX * 0.55 + 1.4) * 0.22
        const baseY = Math.cos(breezeState.phaseY) * 0.35
        const layeredY = Math.sin(breezeState.phaseY * 0.38 - 0.8) * 0.18

        const targetX = baseX + layeredX + gustEnvelope * 0.55 + breezeState.driftX * 0.35
        const targetY = baseY + layeredY + gustEnvelope * 0.8 + breezeState.driftY * 0.45

        pointerInfluence.targetX = THREE.MathUtils.clamp(targetX, -1, 1)
        pointerInfluence.targetY = THREE.MathUtils.clamp(targetY, -1, 1)
    }

    const updateRibbon = () => {
        if (!ribbonMesh) {
            return
        }

        const geometry = ribbonMesh.geometry
        const positions = geometry.attributes.position as THREE.BufferAttribute
        const colors = geometry.attributes.color as THREE.BufferAttribute

        const widthSegments = config.segments
        const heightSegments = 20
        const verticesPerRow = widthSegments + 1

        const baseR = config.baseColor.r
        const baseG = config.baseColor.g
        const baseB = config.baseColor.b
        const glowR = config.glowColor.r
        const glowG = config.glowColor.g
        const glowB = config.glowColor.b

        const pointerEnergy = 1 + Math.min(Math.hypot(pointerInfluence.currentX, pointerInfluence.currentY), 1) * 0.65
        const scrollEnvelope = 0.4 + heroScrollProgress.current * 0.9

        const travel = time * 1.35
        const breathing = Math.sin(time * 0.35) * 0.3

        for (let col = 0; col < verticesPerRow; col += 1) {
            const ratio = col / widthSegments
            const x = ratio * config.length - config.length / 2
            const flowPhase = ratio * (Math.PI * 2) * config.flowFrequency - travel
            const crossPhase = ratio * 10 - time * 0.65

            const breezePush = pointerInfluence.currentX * 0.75
            const lift = pointerInfluence.currentY * 0.9

            let waveZ = Math.sin(flowPhase) * (1 + scrollEnvelope * 0.8)
            waveZ += Math.sin(flowPhase * 1.5 - time * 0.4) * (0.3 + scrollEnvelope * 0.18) * (0.5 + pointerEnergy * 0.4)
            waveZ += Math.sin(crossPhase) * (0.2 + scrollEnvelope * 0.12)
            waveZ += breezePush * 0.55
            waveZ += breathing * 0.18

            const meander = Math.sin(flowPhase * 0.32 - time * 0.18) * 0.35
            const centerY = Math.cos(flowPhase * 0.64 + crossPhase * 0.2) * (0.9 + scrollEnvelope * 0.45) + meander + lift * 0.85

            const twistEnvelope = 0.65 + scrollEnvelope * 0.5
            const twist = Math.sin(flowPhase * 0.9 + crossPhase * 0.4) * config.twistAmplitude * twistEnvelope * (0.8 + pointerEnergy * 0.2)

            const glowSweep = 0.5 + 0.5 * Math.sin(flowPhase + pointerInfluence.currentX * 0.5)
            const highlight = Math.pow(glowSweep, 4) * (0.6 + pointerEnergy * 0.4)
            const crossHighlight = Math.max(Math.cos(flowPhase * 0.5 - time * 0.6), 0)
            const mixRatio = Math.min(highlight + Math.pow(crossHighlight, 3) * 0.7, 1)

            const r = baseR + (glowR - baseR) * mixRatio
            const g = baseG + (glowG - baseG) * mixRatio
            const b = baseB + (glowB - baseB) * mixRatio

            for (let row = 0; row <= heightSegments; row++) {
                const idx = row * verticesPerRow + col
                const v = row / heightSegments
                const offset = (v - 0.5) * config.width

                const y = centerY + offset * Math.cos(twist)
                const z = waveZ + offset * Math.sin(twist)

                positions.setY(idx, y)
                positions.setZ(idx, z)
                colors.setXYZ(idx, r, g, b)
            }
        }

        positions.needsUpdate = true
        colors.needsUpdate = true
        geometry.computeVertexNormals()
    }

    const animateSilk = () => {
        animationFrameId = requestAnimationFrame(animateSilk)

        const now = typeof performance !== 'undefined' ? performance.now() : Date.now()
        const deltaMs = lastFrameTime === 0 ? 16 : now - lastFrameTime
        const deltaSeconds = Math.min(deltaMs, 48) / 1000
        lastFrameTime = now

        applyAutonomousBreeze(deltaSeconds)

        pointerInfluence.currentX += (pointerInfluence.targetX - pointerInfluence.currentX) * 0.04
        pointerInfluence.currentY += (pointerInfluence.targetY - pointerInfluence.currentY) * 0.05
        heroScrollProgress.current += (heroScrollProgress.target - heroScrollProgress.current) * 0.08

        const pointerEnergy = Math.min(Math.hypot(pointerInfluence.currentX, pointerInfluence.currentY), 1)
        const scrollEnergy = heroScrollProgress.current
        const energyMix = Math.min(pointerEnergy * 0.6 + scrollEnergy * 0.8, 1)

        targetConfig.speed = 0.11 + scrollEnergy * 0.08
        targetConfig.twistSpeed = 0.04 + pointerEnergy * 0.03
        targetConfig.twistAmplitude = 0.6 + scrollEnergy * 0.28 + pointerEnergy * 0.15
        targetConfig.flowFrequency = 0.32 + scrollEnergy * 0.24
        targetConfig.baseColor.copy(baseSilkColor).lerp(highlightSilkColor, energyMix * 0.6)
        targetConfig.glowColor.copy(baseGlowColor).lerp(accentGlowColor, 0.35 + energyMix * 0.4)

        config.speed += (targetConfig.speed - config.speed) * 0.05
        config.twistSpeed += (targetConfig.twistSpeed - config.twistSpeed) * 0.05
        config.twistAmplitude += (targetConfig.twistAmplitude - config.twistAmplitude) * 0.05
        config.flowFrequency += (targetConfig.flowFrequency - config.flowFrequency) * 0.05
        config.baseColor.lerp(targetConfig.baseColor, 0.05)
        config.glowColor.lerp(targetConfig.glowColor, 0.05)

        const flowTimeScale = 0.5 + config.speed * 2.4
        time += deltaSeconds * flowTimeScale
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
            ribbonMesh.position.y = silkBasePosition.y + pointerInfluence.currentY * 0.5
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
        lastFrameTime = 0

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
        if (renderer?.domElement && containerRef.value?.contains(renderer.domElement)) {
            containerRef.value.removeChild(renderer.domElement)
        }
        renderer = null
        scene = null
        camera = null
    }

    const supportsWebGL = () => {
        if (typeof document === 'undefined') {
            return false
        }
        const canvas = document.createElement('canvas')
        const gl = canvas.getContext('webgl') || canvas.getContext('experimental-webgl')
        return Boolean(gl)
    }

    const initSilkCanvas = () => {
        if (isReducedMotion.value || shouldUseStaticSilk.value || !containerRef.value || renderer) {
            return
        }

        const container = containerRef.value
        resetBreezeState()
        scene = new THREE.Scene()

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

        const geometry = new THREE.PlaneGeometry(config.length * 1.45, config.width * 1.1, config.segments, 30)
        const positionAttr = geometry.attributes.position as THREE.BufferAttribute | undefined
        if (!positionAttr) {
            console.error('PlaneGeometry is missing position attribute')
            return
        }
        const colorAttr = new THREE.BufferAttribute(new Float32Array(positionAttr.count * 3), 3)
        geometry.setAttribute('color', colorAttr)

        silkMaterial = new THREE.MeshPhysicalMaterial({
            color: 0xfefcf7,
            vertexColors: true,
            emissive: 0xfaf3e2,
            emissiveIntensity: 0.35,
            metalness: 0.18,
            roughness: 0.24,
            clearcoat: 0.96,
            clearcoatRoughness: 0.18,
            transmission: 0.18,
            thickness: 1.2,
            sheen: 1,
            sheenColor: new THREE.Color(0xfff5df),
            sheenRoughness: 0.6,
            iridescence: 0.28,
            iridescenceIOR: 1.2,
            iridescenceThicknessRange: [140, 360],
            envMapIntensity: 0.55,
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
                silkBaseRotation.y = silkVerticalSpin
                silkBaseRotation.z = Math.PI / 4.2
                silkBasePosition.set(0.05, -0.6, -1.75)
            } else if (isTablet) {
                silkBaseRotation.x = Math.PI / 3.1
                silkBaseRotation.y = silkVerticalSpin
                silkBaseRotation.z = Math.PI / 2.7
                silkBasePosition.set(0.45, -1.15, 0.35)
            } else {
                silkBaseRotation.x = Math.PI / 3.6
                silkBaseRotation.y = silkVerticalSpin
                silkBaseRotation.z = Math.PI / 2.35
                silkBasePosition.set(0.95, -1.65, 1.55)
            }

            ribbonMesh.rotation.set(silkBaseRotation.x, silkBaseRotation.y, silkBaseRotation.z)
            ribbonMesh.position.copy(silkBasePosition)
        }
        updateMeshPosition()
        scene.add(ribbonMesh)

        const ambientLight = new THREE.AmbientLight(0xfdf8ef, 0.85)
        scene.add(ambientLight)

        const mainLight = new THREE.DirectionalLight(0xfff1dc, 2.2)
        mainLight.position.set(12, 14, 8)
        scene.add(mainLight)

        const fillLight = new THREE.DirectionalLight(0xdfe9ff, 1.2)
        fillLight.position.set(-6, -8, 4)
        scene.add(fillLight)

        const hemiLight = new THREE.HemisphereLight(0xfffbf5, 0xe7ecff, 0.8)
        scene.add(hemiLight)

        const backLight = new THREE.SpotLight(0xf3e6ff, 2.5)
        backLight.position.set(0, 10, -6)
        backLight.lookAt(0, 0, 0)
        scene.add(backLight)

        resizeHandler = () => {
            if (!renderer || !camera || !containerRef.value) {
                return
            }
            const bounds = containerRef.value.getBoundingClientRect()
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

    return {
        heroScrollProgress,
        evaluateFallback: evaluateSilkFallback,
        syncMotionPreference,
        dispose: disposeSilk,
    }
}
