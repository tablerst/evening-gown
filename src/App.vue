<script setup lang="ts">
import { onMounted, onBeforeUnmount, ref } from 'vue'
import gsap from 'gsap'
import { ScrollTrigger } from 'gsap/ScrollTrigger'
import { createSilkRenderer } from './modules/silk/silkRenderer'
import type { NetworkInformationLike, NavigatorWithConnection } from './modules/silk/silkRenderer'

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
const isNavCompacted = ref(false)
const isReducedMotion = ref(false)
const shouldUseStaticSilk = ref(false)

let scrollHandler: (() => void) | null = null
let motionQuery: MediaQueryList | null = null
let motionChangeHandler: ((event: MediaQueryListEvent) => void) | null = null
let networkInfo: NetworkInformationLike | null = null
let networkChangeHandler: (() => void) | null = null

const {
    heroScrollProgress,
    evaluateFallback: evaluateSilkFallback,
    syncMotionPreference,
    dispose: disposeSilk,
} = createSilkRenderer({
    containerRef: silkContainer,
    isReducedMotion,
    shouldUseStaticSilk,
})

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
                        <article class="bento-card bento-card--xl bento-card--media project-item">
                            <div class="bento-card__media" aria-hidden="true">
                                <img src="https://images.unsplash.com/photo-1524504388940-b1c1722653e1?q=80&w=1980&auto=format&fit=crop"
                                    alt="Architected halo gown" loading="lazy" />
                            </div>
                            <div class="bento-card__meta">
                                <p class="bento-card__caption">LOOK 01 · STRUCTURE</p>
                                <h3 class="bento-card__title">Architected Halo</h3>
                                <p class="bento-card__caption">3D Corsetry · Hidden boning</p>
                            </div>
                        </article>

                        <article class="bento-card bento-card--tall bento-card--media project-item">
                            <div class="bento-card__media" aria-hidden="true">
                                <img src="https://images.unsplash.com/photo-1515372039744-b8f02a3ae446?q=80&w=800&auto=format&fit=crop"
                                    alt="Hand pleated silk drape" loading="lazy" />
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

                        <article class="bento-card bento-card--media project-item">
                            <div class="bento-card__media" aria-hidden="true">
                                <img src="https://images.unsplash.com/photo-1595777457583-95e059d581b8?q=80&w=800&auto=format&fit=crop"
                                    alt="Pearl embroidery detail" loading="lazy" />
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
