import { nextTick } from 'vue'
import type { Ref } from 'vue'
import * as THREE from 'three'
import { RoomEnvironment } from 'three/examples/jsm/environments/RoomEnvironment.js'

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
    let pmremGenerator: THREE.PMREMGenerator | null = null
    let environmentTexture: THREE.Texture | null = null
    let environmentRenderTarget: THREE.WebGLRenderTarget | null = null
    let fabricNormalTexture: THREE.Texture | null = null
    let animationFrameId: number | null = null
    let resizeHandler: (() => void) | null = null
    let time = 0
    let lastFrameTime = 0

    // 不考虑性能：提高横向细分，避免“像纸片”一样的生硬折线。
    // 这里的 heightSegments 对应 PlaneGeometry 的 heightSegments，必须与 updateRibbon 的遍历一致。
    const ribbonHeightSegments = 80

    const createFabricNormalMap = (size = 256) => {
        // 程序化“织物细纹”法线贴图：让高光更碎、更像真实丝绸纤维微表面。
        // 先做一个可平铺高度场，再用中心差分法生成法线。
        const height = new Float32Array(size * size)

        const hash = (x: number, y: number) => {
            const s = Math.sin(x * 127.1 + y * 311.7) * 43758.5453123
            return s - Math.floor(s)
        }

        const smoothstep = (t: number) => t * t * (3 - 2 * t)

        const valueNoise = (x: number, y: number) => {
            const x0 = Math.floor(x)
            const y0 = Math.floor(y)
            const x1 = x0 + 1
            const y1 = y0 + 1

            const sx = smoothstep(x - x0)
            const sy = smoothstep(y - y0)

            const n00 = hash(x0, y0)
            const n10 = hash(x1, y0)
            const n01 = hash(x0, y1)
            const n11 = hash(x1, y1)

            const ix0 = n00 + (n10 - n00) * sx
            const ix1 = n01 + (n11 - n01) * sx
            return ix0 + (ix1 - ix0) * sy
        }

        const fbm = (x: number, y: number) => {
            let amp = 0.6
            let freq = 1
            let sum = 0
            let norm = 0
            for (let i = 0; i < 5; i += 1) {
                sum += amp * (valueNoise(x * freq, y * freq) * 2 - 1)
                norm += amp
                amp *= 0.5
                freq *= 2
            }
            return sum / norm
        }

        for (let y = 0; y < size; y += 1) {
            for (let x = 0; x < size; x += 1) {
                const u = x / size
                const v = y / size

                // 横向纤维细纹（近似织物方向）+ 少量交错扰动
                const fiber = Math.sin(u * Math.PI * 2 * 42) * 0.18
                const cross = Math.sin(v * Math.PI * 2 * 9 + u * 6.0) * 0.12
                const noise = fbm(u * 6, v * 6) * 0.22
                height[y * size + x] = fiber + cross + noise
            }
        }

        const data = new Uint8Array(size * size * 4)
        const strength = 3.2
        const wrap = (n: number) => (n + size) % size

        for (let y = 0; y < size; y += 1) {
            for (let x = 0; x < size; x += 1) {
                const hL = height[y * size + wrap(x - 1)] ?? 0
                const hR = height[y * size + wrap(x + 1)] ?? 0
                const hD = height[wrap(y - 1) * size + x] ?? 0
                const hU = height[wrap(y + 1) * size + x] ?? 0

                const dx = (hR - hL) * strength
                const dy = (hU - hD) * strength

                const n = new THREE.Vector3(-dx, -dy, 1).normalize()
                const idx = (y * size + x) * 4
                data[idx] = Math.round((n.x * 0.5 + 0.5) * 255)
                data[idx + 1] = Math.round((n.y * 0.5 + 0.5) * 255)
                data[idx + 2] = Math.round((n.z * 0.5 + 0.5) * 255)
                data[idx + 3] = 255
            }
        }

        const tex = new THREE.DataTexture(data, size, size, THREE.RGBAFormat)
        tex.colorSpace = THREE.NoColorSpace
        tex.wrapS = THREE.RepeatWrapping
        tex.wrapT = THREE.RepeatWrapping
        tex.repeat.set(14, 5)
        tex.needsUpdate = true
        return tex
    }

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
        const heightSegments = ribbonHeightSegments
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
        const microTime = time * 0.9

        for (let col = 0; col < verticesPerRow; col += 1) {
            const ratio = col / widthSegments
            const flowPhase = ratio * (Math.PI * 2) * config.flowFrequency - travel
            const crossPhase = ratio * 10 - time * 0.65

            const breezePush = pointerInfluence.currentX * 0.75
            const lift = pointerInfluence.currentY * 0.9

            let waveZ = Math.sin(flowPhase) * (1 + scrollEnvelope * 0.8)
            waveZ += Math.sin(flowPhase * 1.5 - time * 0.4) * (0.3 + scrollEnvelope * 0.18) * (0.5 + pointerEnergy * 0.4)
            waveZ += Math.sin(crossPhase) * (0.2 + scrollEnvelope * 0.12)

            // 细密褶皱/湍流：提升“丝绸碎高光”的层次（不考虑性能）
            const microWrinkle =
                Math.sin(flowPhase * 6.5 + microTime * 1.8) * 0.08 +
                Math.sin(flowPhase * 10.5 - microTime * 1.2) * 0.05
            const turbulence =
                Math.sin(flowPhase * 2.2 + crossPhase * 1.7) * 0.06 +
                Math.sin(flowPhase * 3.8 - crossPhase * 1.1 + microTime) * 0.04

            waveZ += breezePush * 0.55
            waveZ += breathing * 0.18
            waveZ += (microWrinkle + turbulence) * (0.55 + scrollEnvelope * 0.5) * (0.5 + pointerEnergy * 0.5)

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

                // 边缘更容易抖动与起皱（更像薄面料）
                const edge = Math.pow(Math.abs(v - 0.5) * 2, 1.25)
                const edgeFlutter =
                    Math.sin(flowPhase * 5.2 + microTime * 2.6 + v * 9.0) * 0.09 +
                    Math.sin(flowPhase * 8.4 - microTime * 1.9 + v * 5.0) * 0.05
                const edgeLift = edge * edgeFlutter * (0.6 + pointerEnergy * 0.55) * (0.5 + scrollEnvelope * 0.7)

                const y = centerY + offset * Math.cos(twist)
                const z = waveZ + offset * Math.sin(twist) + edgeLift

                positions.setY(idx, y)
                positions.setZ(idx, z)
                colors.setXYZ(idx, r, g, b)
            }
        }

        positions.needsUpdate = true
        colors.needsUpdate = true
        geometry.computeVertexNormals()

        // 各向异性高光可能依赖 tangent；动态变形后重算（不考虑性能）。
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        const geoAny = geometry as any
        if (typeof geoAny.computeTangents === 'function') {
            try {
                geoAny.computeTangents()
            } catch {
                // computeTangents 可能因几何属性条件不满足而失败，忽略即可。
            }
        }
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

        if (scene) {
            scene.environment = null
        }

        silkMaterial?.dispose()
        silkMaterial = null
        ribbonMesh = null

        if (fabricNormalTexture) {
            fabricNormalTexture.dispose()
            fabricNormalTexture = null
        }

        if (environmentRenderTarget) {
            environmentRenderTarget.dispose()
            environmentRenderTarget = null
        }

        environmentTexture = null

        pmremGenerator?.dispose()
        pmremGenerator = null

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

        renderer = new THREE.WebGLRenderer({ antialias: true, alpha: true, powerPreference: 'high-performance' })
        renderer.setPixelRatio(Math.min(window.devicePixelRatio, 3))
        renderer.setSize(width, height)
        renderer.setClearAlpha(0)
        renderer.toneMapping = THREE.ACESFilmicToneMapping
        renderer.toneMappingExposure = 1.05
        renderer.outputColorSpace = THREE.SRGBColorSpace
        container.appendChild(renderer.domElement)

        // PMREM 环境：丝绸高光非常依赖环境反射，缺了会很“假”。
        pmremGenerator = new THREE.PMREMGenerator(renderer)
        pmremGenerator.compileEquirectangularShader()
        environmentRenderTarget = pmremGenerator.fromScene(new RoomEnvironment(), 0.04)
        environmentTexture = environmentRenderTarget.texture

        const geometry = new THREE.PlaneGeometry(
            config.length * 1.45,
            config.width * 1.1,
            config.segments,
            ribbonHeightSegments
        )
        const positionAttr = geometry.attributes.position as THREE.BufferAttribute | undefined
        if (!positionAttr) {
            console.error('PlaneGeometry is missing position attribute')
            return
        }
        const colorAttr = new THREE.BufferAttribute(new Float32Array(positionAttr.count * 3), 3)
        geometry.setAttribute('color', colorAttr)

        fabricNormalTexture = createFabricNormalMap(256)

        silkMaterial = new THREE.MeshPhysicalMaterial({
            // 丝绸为介电材质：metalness 应接近 0；高光主要来自 clearcoat + sheen + 各向异性。
            color: 0xfefcf7,
            vertexColors: true,

            emissive: 0xfaf3e2,
            emissiveIntensity: 0.18,

            metalness: 0.0,
            roughness: 0.18,
            ior: 1.45,

            // 表层强高光（类似薄清漆）
            clearcoat: 1.0,
            clearcoatRoughness: 0.08,

            // 织物 sheen：柔亮、带色的边缘高光
            sheen: 1.0,
            sheenColor: new THREE.Color(0xfff2df),
            sheenRoughness: 0.35,

            // 轻微薄膜彩色漂移：控制在小范围，避免塑料感
            iridescence: 0.12,
            iridescenceIOR: 1.25,
            iridescenceThicknessRange: [160, 360],

            // 各向异性高光（沿织物纹理方向）
            anisotropy: 0.85,
            anisotropyRotation: Math.PI / 2,

            // 丝绸很薄：少量透光即可
            transmission: 0.04,
            thickness: 0.08,
            attenuationDistance: 0.8,
            attenuationColor: new THREE.Color(0xfff7ea),

            // 微法线，让高光更“丝”
            normalMap: fabricNormalTexture,
            normalScale: new THREE.Vector2(0.22, 0.22),

            // 环境反射强度
            envMapIntensity: 1.15,

            // 额外高光控制（three 版本支持时生效）
            specularIntensity: 0.85,
            specularColor: new THREE.Color(0xffffff),

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

        // 绑定环境反射到 scene
        scene.environment = environmentTexture

        const ambientLight = new THREE.AmbientLight(0xfdf8ef, 0.55)
        scene.add(ambientLight)

        const mainLight = new THREE.DirectionalLight(0xfff1dc, 4.2)
        mainLight.position.set(12, 16, 10)
        scene.add(mainLight)

        const fillLight = new THREE.DirectionalLight(0xdfe9ff, 2.0)
        fillLight.position.set(-10, -10, 8)
        scene.add(fillLight)

        const hemiLight = new THREE.HemisphereLight(0xfffbf5, 0xe7ecff, 0.65)
        scene.add(hemiLight)

        const backLight = new THREE.SpotLight(0xf3e6ff, 6.5)
        backLight.position.set(0, 12, -10)
        backLight.angle = THREE.MathUtils.degToRad(32)
        backLight.penumbra = 0.75
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

        const lacksWebGL = !supportsWebGL()
        // 本版本以“画质优先”为目标：不再因为 CPU/网络做静态回退，仅在缺少 WebGL 时回退。
        const shouldFallback = lacksWebGL

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
