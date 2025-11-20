<script setup lang="ts">
import { onMounted, nextTick } from 'vue'
import gsap from 'gsap'
import { ScrollTrigger } from 'gsap/ScrollTrigger'
import ThemeToggle from './components/ThemeToggle.vue'
import SlantedBlock from './components/SlantedBlock.vue'
import CardGown from './components/CardGown.vue'
import HeroBackground3D from './components/HeroBackground3D.vue'
import { splitTextToSpans } from './utils/textAnimation'

gsap.registerPlugin(ScrollTrigger)
// ... existing gowns data ...

const gowns = [
  {
    id: 1,
    title: 'Starry Night Velvet',
    image:
      'https://images.unsplash.com/photo-1566174053879-31528523f8ae?q=80&w=800&auto=format&fit=crop',
    tags: ['年会', '晚宴'],
    price: '¥ 2,880',
    scene: 'Formal Evening',
  },
  {
    id: 2,
    title: 'Champagne Gold Silk',
    image:
      'https://images.unsplash.com/photo-1595777457583-95e059d581b8?q=80&w=800&auto=format&fit=crop',
    tags: ['婚礼', '敬酒'],
    price: '¥ 3,280',
    scene: 'Wedding Guest',
  },
  {
    id: 3,
    title: 'Midnight Blue Tulle',
    image:
      'https://images.unsplash.com/photo-1515372039744-b8f02a3ae446?q=80&w=800&auto=format&fit=crop',
    tags: ['舞会', '派对'],
    price: '¥ 1,980',
    scene: 'Prom Night',
  },
]

onMounted(() => {
  // Hero Animation
  const tl = gsap.timeline()

  // Text Split Animation
  const titleLine1 = document.querySelector('.hero-title-line-1') as HTMLElement
  const titleLine2 = document.querySelector('.hero-title-line-2') as HTMLElement

  if (titleLine1 && titleLine2) {
    const chars1 = splitTextToSpans(titleLine1, 'chars')
    const chars2 = splitTextToSpans(titleLine2, 'chars')

    tl.from(chars1, {
      opacity: 0,
      y: 80,
      rotateX: -90,
      stagger: 0.05,
      duration: 1,
      ease: 'back.out(1.7)',
    })
      .from(chars2, {
        opacity: 0,
        y: 80,
        rotateX: -90,
        stagger: 0.05,
        duration: 1,
        ease: 'back.out(1.7)',
      }, '-=0.8')
  } else {
    // Fallback if split fails or elements missing
    tl.from('.hero-title', { y: 50, opacity: 0, duration: 0.8, ease: 'power3.out' })
  }

  tl.from('.hero-subtitle', { y: 30, opacity: 0, duration: 0.8, ease: 'power3.out' }, '-=0.6')
    .from('.hero-actions', { y: 20, opacity: 0, duration: 0.6, ease: 'power3.out' }, '-=0.6')
    .from('.hero-visual', { x: 50, opacity: 0, duration: 1, ease: 'power3.out' }, '-=0.8')

  // Section Headers
  gsap.utils.toArray<HTMLElement>('.section-header').forEach((header) => {
    gsap.from(header, {
      scrollTrigger: {
        trigger: header,
        start: 'top 85%',
        toggleActions: 'play none none none', // once
      },
      y: 30,
      opacity: 0,
      duration: 0.8,
      ease: 'power3.out',
    })
  })

  // Cards Stagger (Selling Points & Gowns)
  gsap.utils.toArray<HTMLElement>('.grid-3').forEach((grid) => {
    gsap.from(grid.children, {
      scrollTrigger: {
        trigger: grid,
        start: 'top 85%',
      },
      y: 50,
      opacity: 0,
      duration: 0.6,
      stagger: 0.1,
      ease: 'power3.out',
    })
  })

  // Slanted Blocks
  gsap.utils.toArray<HTMLElement>('.slanted-block').forEach((block) => {
    gsap.from(block, {
      scrollTrigger: {
        trigger: block,
        start: 'top 85%',
      },
      opacity: 0,
      scale: 0.98,
      duration: 1,
      ease: 'power2.out',
    })
  })
})
</script>

<template>
  <div class="app-root">
    <!-- Header -->
    <header class="app-header">
      <div class="container">
        <h1 class="brand">Evening Gown</h1>
        <nav class="nav-links">
          <a href="#">系列</a>
          <a href="#">故事</a>
          <a href="#">预约</a>
        </nav>
        <ThemeToggle />
      </div>
    </header>

    <main>
      <!-- 1. Hero Section -->
      <section class="hero-section">
        <SlantedBlock direction="right" height="80vh" bg-color="var(--color-bg-page)"
          image="https://images.unsplash.com/photo-1566737236500-c8ac43014a67?q=80&w=1000&auto=format&fit=crop"
          slanted-mask mask-angle="45deg">
          <div class="container hero-container">
            <div class="hero-content">
              <h1 class="hero-title">
                <span class="hero-title-line-1 block">Elegance</span>
                <span class="hero-title-line-2 font-serif italic block text-brand-primary">Redefined</span>
              </h1>
              <p class="hero-subtitle">专为重要时刻打造的高定礼服系列。淡雅紫金，诠释不凡气质。</p>
              <div class="hero-actions">
                <button class="btn btn--primary btn--lg">预约试纱</button>
                <button class="btn btn--ghost">探索系列 &rarr;</button>
              </div>
              <div class="hero-meta">
                <span>高端定制</span>
                <span class="divider">|</span>
                <span>私人顾问</span>
              </div>
            </div>

            <div class="hero-visual-wrapper">
              <HeroBackground3D />
            </div>
          </div>
        </SlantedBlock>
      </section>

      <!-- 2. Selling Points -->
      <section class="section selling-points">
        <div class="container">
          <div class="grid-3">
            <div class="feature-card">
              <div class="icon-box">✨</div>
              <h3>独家设计</h3>
              <p>融合现代剪裁与经典美学，每一件都是独一无二的艺术品。</p>
            </div>
            <div class="feature-card">
              <div class="icon-box">🧵</div>
              <h3>顶级面料</h3>
              <p>严选进口真丝、蕾丝与施华洛世奇水晶，触感细腻。</p>
            </div>
            <div class="feature-card">
              <div class="icon-box">👑</div>
              <h3>私人定制</h3>
              <p>一对一量体裁衣，确保每一寸线条都完美贴合您的身形。</p>
            </div>
          </div>
        </div>
      </section>

      <!-- 3. Popular Gowns -->
      <section class="section popular-gowns">
        <div class="container">
          <div class="section-header">
            <h2>本季精选系列</h2>
            <div class="filters">
              <button class="chip active">全部</button>
              <button class="chip">婚礼</button>
              <button class="chip">年会</button>
            </div>
          </div>

          <div class="grid-3 gown-grid">
            <CardGown v-for="gown in gowns" :key="gown.id" v-bind="gown" />
          </div>

          <div class="text-center mt-8">
            <button class="btn btn--secondary">查看完整系列</button>
          </div>
        </div>
      </section>

      <!-- 4. Scene Story -->
      <section class="scene-story">
        <SlantedBlock direction="left" height="500px" bg-color="var(--color-neutral-0)" slanted-mask>
          <div class="container grid-2 h-full items-center relative z-10">
            <div class="story-content">
              <span class="overline">SCENE 01</span>
              <h2>璀璨晚宴</h2>
              <p>
                在灯光交错的晚宴现场，一袭流光溢彩的礼服让您成为全场焦点。精致的剪裁勾勒曼妙身姿，自信优雅。
              </p>
              <button class="btn btn--ghost">阅读故事</button>
            </div>
            <div class="story-image">
              <img src="https://images.unsplash.com/photo-1566737236500-c8ac43014a67?q=80&w=800&auto=format&fit=crop"
                alt="Party" class="rounded-lg shadow-lg" />
            </div>
          </div>
        </SlantedBlock>
      </section>

      <!-- 5. Customer Gallery -->
      <section class="section gallery-section">
        <div class="container">
          <div class="section-header text-center" style="justify-content: center; flex-direction: column">
            <h2>她们的高光时刻</h2>
            <p class="subtitle">来自真实客户的返图</p>
          </div>
          <div class="gallery-grid">
            <div class="gallery-item" v-for="i in 4" :key="i">
              <img :src="`https://images.unsplash.com/photo-${[
                  '1515934751635-c81c6bc9a2d8',
                  '1469334031218-e382a71b716b',
                  '1566737236500-c8ac43014a67',
                  '1595777457583-95e059d581b8',
                ][i - 1]
                }?q=80&w=400&h=500&auto=format&fit=crop`" alt="Customer" loading="lazy" />
              <div class="gallery-tag">
                <span>{{
                  ['上海 · 婚礼', '北京 · 年会', '深圳 · 晚宴', '杭州 · 旅拍'][i - 1]
                  }}</span>
              </div>
            </div>
          </div>
        </div>
      </section>

      <!-- 6. Process Section -->
      <section class="section process-section">
        <div class="container">
          <div class="section-header text-center" style="justify-content: center; flex-direction: column">
            <h2>定制流程</h2>
            <p class="subtitle">从量体到成衣的专属体验</p>
          </div>
          <div class="process-steps">
            <div class="step-item">
              <div class="step-icon">1</div>
              <h4>预约咨询</h4>
              <p>线上预约，专属顾问一对一沟通需求</p>
            </div>
            <div class="step-connector"></div>
            <div class="step-item">
              <div class="step-icon">2</div>
              <h4>量体试纱</h4>
              <p>到店精准量体，试穿多款样衣</p>
            </div>
            <div class="step-connector"></div>
            <div class="step-item">
              <div class="step-icon">3</div>
              <h4>精细调整</h4>
              <p>根据身形数据进行微调修改</p>
            </div>
            <div class="step-connector"></div>
            <div class="step-item">
              <div class="step-icon">4</div>
              <h4>完美交付</h4>
              <p>最终试穿确认，包装交付</p>
            </div>
          </div>
        </div>
      </section>

      <!-- 7. FAQ Section -->
      <section class="section faq-section">
        <div class="container container-narrow">
          <div class="section-header text-center" style="justify-content: center">
            <h2>常见问题</h2>
          </div>
          <div class="faq-list">
            <details class="faq-item">
              <summary>需要提前多久预约？</summary>
              <p>
                建议提前 3-7 天预约试纱，以便我们为您安排专属顾问和试衣间。如果是定制礼服，建议提前
                2-3 个月。
              </p>
            </details>
            <details class="faq-item">
              <summary>试纱是否收费？</summary>
              <p>
                首次试纱提供 3
                件免费试穿体验。如需更多款式试穿或专业造型服务，会收取一定的试纱费，该费用可在定单时抵扣。
              </p>
            </details>
            <details class="faq-item">
              <summary>可以租赁吗？</summary>
              <p>是的，我们提供高定礼服的租赁服务，租期通常为 3 天。同时也提供量身定制购买服务。</p>
            </details>
          </div>
        </div>
      </section>

      <!-- 8. CTA -->
      <section class="cta-section">
        <SlantedBlock direction="right" height="400px" bg-color="var(--color-purple-700)"
          overlay-color="rgba(0,0,0,0.3)">
          <div class="container h-full flex-center flex-col text-inverse">
            <h2>准备好闪耀全场了吗？</h2>
            <p class="mb-6">立即预约您的私人试纱体验</p>
            <button class="btn btn--primary btn--lg"
              style="background: var(--color-gold-400); color: var(--color-neutral-800)">
              立即预约
            </button>
          </div>
        </SlantedBlock>
      </section>
    </main>

    <footer class="app-footer">
      <div class="container">
        <div class="footer-content">
          <div class="brand-col">
            <h2 class="brand">Evening Gown</h2>
            <p>© 2025 All Rights Reserved.</p>
          </div>
          <div class="links-col">
            <h4>关于</h4>
            <a href="#">品牌故事</a>
            <a href="#">设计师</a>
          </div>
          <div class="links-col">
            <h4>服务</h4>
            <a href="#">预约试纱</a>
            <a href="#">定制流程</a>
          </div>
        </div>
      </div>
    </footer>
  </div>
</template>

<style scoped lang="scss">
@use '@/assets/styles/abstracts/variables' as vars;

.app-root {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  overflow-x: hidden;
}

.app-header {
  padding: vars.$space-4 0;
  background: rgba(255, 255, 255, 0.8);
  backdrop-filter: blur(10px);
  position: sticky;
  top: 0;
  z-index: 100;
  border-bottom: 1px solid var(--color-border-subtle);

  .container {
    display: flex;
    align-items: center;
    justify-content: space-between;
  }
}

.brand {
  font-family: vars.$font-family-serif;
  color: var(--color-brand-primary);
  font-size: vars.$font-size-xl;
  font-weight: 700;
}

.nav-links {
  display: none;
  gap: vars.$space-6;

  @media (min-width: 768px) {
    display: flex;
  }

  a {
    font-size: vars.$font-size-sm;
    font-weight: 500;
    color: var(--color-text-primary);

    &:hover {
      color: var(--color-brand-primary);
    }
  }
}

// Hero
.hero-section {
  position: relative;
  overflow: hidden;
  background: var(--color-bg-page);
}

.hero-container {
  display: grid;
  grid-template-columns: 1fr;
  gap: vars.$space-5;
  padding-top: vars.$space-6;
  padding-bottom: vars.$space-6;

  @media (min-width: 900px) {
    grid-template-columns: 1fr 1fr;
    align-items: center;
    height: 80vh;
    min-height: 600px;
    padding: 0;
  }
}

.hero-content {
  z-index: 10;
  padding: vars.$space-4;
}

.hero-title {
  font-family: vars.$font-family-serif;
  font-size: vars.$font-size-4xl;
  line-height: 1.1;
  margin-bottom: vars.$space-4;
  color: var(--color-brand-dark);

  .block {
    display: block;
  }

  .text-brand-primary {
    color: var(--color-brand-primary);
  }

  .italic {
    font-style: italic;
  }
}

.hero-subtitle {
  font-size: vars.$font-size-lg;
  color: var(--color-text-secondary);
  margin-bottom: vars.$space-6;
  max-width: 480px;
}

.hero-actions {
  display: flex;
  gap: vars.$space-4;
  margin-bottom: vars.$space-6;
}

.hero-meta {
  font-size: vars.$font-size-sm;
  color: var(--color-text-secondary);
  display: flex;
  gap: vars.$space-3;

  .divider {
    color: var(--color-border-subtle);
  }
}

.hero-visual {
  position: relative;
  height: 400px;
  width: 100%;

  @media (min-width: 900px) {
    height: 100%;
    position: absolute;
    right: 0;
    top: 0;
    width: 50%;
  }
}

// Sections
.section {
  padding: vars.$space-8 0;
}

.selling-points {
  background: var(--color-bg-section-alt);
}

.feature-card {
  text-align: center;
  padding: vars.$space-5;

  .icon-box {
    font-size: 2.5rem;
    margin-bottom: vars.$space-4;
  }

  h3 {
    margin-bottom: vars.$space-2;
    color: var(--color-brand-dark);
  }

  p {
    color: var(--color-text-secondary);
    font-size: vars.$font-size-sm;
  }
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: vars.$space-6;
  flex-wrap: wrap;
  gap: vars.$space-4;
}

.filters {
  display: flex;
  gap: vars.$space-2;

  .chip {
    padding: 6px 16px;
    border-radius: vars.$radius-pill;
    border: 1px solid var(--color-border-subtle);
    background: transparent;
    cursor: pointer;
    font-size: vars.$font-size-sm;
    color: var(--color-text-secondary);

    &.active,
    &:hover {
      background: var(--color-brand-primary);
      color: white;
      border-color: var(--color-brand-primary);
    }
  }
}

// Scene Story
.scene-story {
  margin: vars.$space-8 0;
}

.story-content {
  padding: vars.$space-6;

  .overline {
    font-size: vars.$font-size-xs;
    letter-spacing: 0.1em;
    color: var(--color-brand-accent);
    font-weight: 700;
    display: block;
    margin-bottom: vars.$space-2;
  }

  h2 {
    margin-bottom: vars.$space-4;
    color: var(--color-brand-dark);
  }

  p {
    margin-bottom: vars.$space-5;
    color: var(--color-text-secondary);
    line-height: 1.8;
  }
}

.story-image {
  padding: vars.$space-5;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;

  img {
    max-height: 400px;
    width: auto;
    object-fit: cover;
  }
}

// Utilities
.grid-3 {
  display: grid;
  grid-template-columns: 1fr;
  gap: vars.$space-5;

  @media (min-width: 768px) {
    grid-template-columns: repeat(3, 1fr);
  }
}

.grid-2 {
  display: grid;
  grid-template-columns: 1fr;
  gap: vars.$space-5;

  @media (min-width: 768px) {
    grid-template-columns: repeat(2, 1fr);
  }
}

.h-full {
  height: 100%;
}

.items-center {
  align-items: center;
}

.flex-center {
  display: flex;
  align-items: center;
  justify-content: center;
}

.flex-col {
  flex-direction: column;
}

.text-inverse {
  color: var(--color-text-inverse);
  text-align: center;
}

.mb-6 {
  margin-bottom: vars.$space-6;
}

.mt-8 {
  margin-top: vars.$space-8;
}

.rounded-lg {
  border-radius: vars.$radius-lg;
}

.shadow-lg {
  box-shadow: vars.$shadow-md;
}

// Footer
.app-footer {
  background: var(--color-neutral-800);
  color: var(--color-neutral-200);
  padding: vars.$space-8 0;

  .footer-content {
    display: grid;
    grid-template-columns: 1fr;
    gap: vars.$space-8;

    @media (min-width: 768px) {
      grid-template-columns: 2fr 1fr 1fr;
    }
  }

  h4 {
    color: white;
    margin-bottom: vars.$space-4;
  }

  a {
    display: block;
    margin-bottom: vars.$space-2;
    opacity: 0.7;

    &:hover {
      opacity: 1;
    }
  }
}

// Gallery
.gallery-section {
  background: var(--color-bg-section-alt);
}

.gallery-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: vars.$space-4;

  @media (min-width: 768px) {
    grid-template-columns: repeat(4, 1fr);
  }
}

.gallery-item {
  position: relative;
  border-radius: vars.$radius-md;
  overflow: hidden;
  aspect-ratio: 3/4;

  img {
    width: 100%;
    height: 100%;
    object-fit: cover;
    transition: transform 0.4s ease;
  }

  &:hover img {
    transform: scale(1.05);
  }

  .gallery-tag {
    position: absolute;
    bottom: vars.$space-3;
    left: vars.$space-3;
    background: rgba(255, 255, 255, 0.9);
    padding: 4px 12px;
    border-radius: vars.$radius-pill;
    font-size: vars.$font-size-xs;
    font-weight: 500;
    color: var(--color-text-primary);
  }
}

// Process
.process-section {
  background: var(--color-bg-page);
}

.process-steps {
  display: flex;
  flex-direction: column;
  gap: vars.$space-6;
  align-items: center;

  @media (min-width: 768px) {
    flex-direction: row;
    justify-content: space-between;
    align-items: flex-start;
  }
}

.step-item {
  text-align: center;
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;

  .step-icon {
    width: 48px;
    height: 48px;
    border-radius: 50%;
    background: var(--color-brand-primary);
    color: white;
    display: flex;
    align-items: center;
    justify-content: center;
    font-weight: bold;
    margin-bottom: vars.$space-3;
    font-family: vars.$font-family-serif;
    font-size: vars.$font-size-lg;
  }

  h4 {
    margin-bottom: vars.$space-2;
    color: var(--color-brand-dark);
  }

  p {
    font-size: vars.$font-size-sm;
    color: var(--color-text-secondary);
    max-width: 200px;
  }
}

.step-connector {
  display: none;

  @media (min-width: 768px) {
    display: block;
    height: 1px;
    background: var(--color-gold-400);
    flex: 0.5;
    margin-top: 24px; // Half of icon height
  }
}

// FAQ
.faq-section {
  background: var(--color-bg-section-alt);
}

.container-narrow {
  max-width: 840px;
  margin: 0 auto;
}

.faq-list {
  display: flex;
  flex-direction: column;
  gap: vars.$space-4;
}

.faq-item {
  border-bottom: 1px solid var(--color-border-subtle);
  padding-bottom: vars.$space-4;

  summary {
    font-weight: 600;
    cursor: pointer;
    list-style: none;
    position: relative;
    padding-right: 24px;
    color: var(--color-text-primary);

    &::-webkit-details-marker {
      display: none;
    }

    &::after {
      content: '+';
      position: absolute;
      right: 0;
      color: var(--color-brand-primary);
      font-weight: 300;
      font-size: 1.5rem;
      line-height: 1;
    }
  }

  &[open] summary::after {
    content: '-';
  }

  p {
    margin-top: vars.$space-3;
    color: var(--color-text-secondary);
    line-height: 1.6;
    font-size: vars.$font-size-sm;
  }
}

.subtitle {
  color: var(--color-text-secondary);
  margin-top: vars.$space-2;
}
</style>
