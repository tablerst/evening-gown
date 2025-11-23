<!-- Design.md -->

# Noir Luminescence – Web UI Design 规范  
高定晚礼服官网「Midnight Refraction | NOIR & ÉCLAT」

---

## 1. 设计原则

1. **暗夜舞台感**  
   - 整体以极深的黑色为背景，所有内容被视为在「黑幕」上被灯光照亮的元素。  
   - 图片、文字、光效均从黑色中“浮现”，不做大面积浅色铺底。

2. **克制的奢华感**  
   - 多使用细线条、小字号、大留白来体现高级感。  
   - 金色仅作为点缀而非主色，用于边框、分隔线、小符号。

3. **冷调霓虹 + 高定时装融合**  
   - 标题与氛围用蓝紫霓虹渐变，正文偏冷银色，整体气质偏冷。  
   - 模特图片与礼服质感保持时装大片风格：低饱和、高对比。

4. **内容层级清晰**  
   - Hero 首屏用于品牌记忆与主视觉，不承担过多信息。  
   - 首屏以下逐步展开：系列标题 → 单品展示 → 品牌触点（社交、联系）。

---

## 2. 页面结构与布局

### 2.1 全局布局

- 页面宽度：流式布局，最大内容宽度建议控制在 `1200–1440px`。  
- 背景：全局使用深黑色（`var(--void-black)`），局部通过渐变和光晕叠加层增加空间感。  
- 字体：
  - 英文标题：`Cinzel`（衬线、粗体、带古典雕塑感）。
  - 英文斜体 / 引言：`Playfair Display`。
  - 其它正文与导航：`Montserrat`。
- 滚动方向：纵向单页式，首屏为满屏 Hero，以下为模块化 section。

### 2.2 导航栏（Nav）

**位置与行为**

- 固定在页面顶部：`fixed top-0 w-full z-50`，全站共享。  
- 高度适中（约 `64px`），左右内边距 `px-8`，内容水平三段布局：

  1. 左侧：品牌 Logo
  2. 中间：主导航菜单（桌面端可见）
  3. 右侧：主行动按钮 `Private View`

**视觉风格**

- 背景：`nav-glass`  
  - 半透明近黑背景：`rgba(3, 3, 5, 0.85)`  
  - 背景模糊：`backdrop-filter: blur(12px)`  
  - 底部一条极细深蓝紫描边：`border-bottom: 1px solid rgba(25, 25, 112, 0.3)`

- Logo：
  - 文本：`NOIR & ÉCLAT`，大写，字间距较大（`tracking-[0.2em]`），白色。
  - 左侧小符号：金色五角星 `✦`，用于建立品牌记忆点。

- 导航菜单项：
  - 文本：全大写、字距拉大（`tracking-[0.2em]`），使用 `text-muted` 颜色。
  - hover：过渡到纯白，突出当前悬停项。
  - 间距：`space-x-12`，保证足够呼吸感。

- 右侧按钮 `Private View`：
  - 黑色背景 + 金色细描边 + 金色文字。
  - hover 状态：填充金色、文字变黑，表现为“反转”式高亮。
  - 字体：极小字号 + 大字距，强调精致与克制。

---

### 2.3 首屏 Hero 区

**区域设定**

- 高度：`h-screen`，占满首屏视窗。  
- 水平与垂直居中：`flex items-center justify-center`。  
- 背景层更新为「动态丝绸引擎」：

  1. **WebGL 动态丝绸**（`#silk-canvas`）  
    - 使用 Three.js 渲染 400 段的丝绸平面，持续扭转与流淌。  
    - 基础色 `baseColor` 取自 `--void-black` 的蓝黑调，光晕 `glowColor` 在 `--royal-purple` 与寒冷青蓝之间缓慢过渡，符合 Noir Luminescence 主题。  
    - 添加 Exponential Fog 与 ACES Filmic tone mapping，确保丝绸远端融入黑幕。

  2. **渐变遮罩层**  
    - 左侧偏午夜蓝 (`midnight-blue`)，右侧偏皇家紫 (`royal-purple`)，中间加黑色过渡，形成冷调光影。  
    - 顶部另加自上而下的黑色渐变，使导航区域更加清晰。

  3. **AI Mood Weaver 面板**  
    - 右上角悬浮毛玻璃面板（宽 280-320px），半透明深灰背景 + 细描边。  
    - 包含简约标题、情绪输入文本域、`GENERATE` 按钮以及来自 Gemini 的诗性描述。  
    - 按钮在调用时显示微型 spinner 与 `WEAVING` 文字，成功后在 `ai-response` 区淡入描述。

**内容布局**

- 垂直结构（自上而下）：
  1. 副标题：`The 2025 Midnight Series`
  2. 主标题：`OBSIDIAN / DREAMS`
  3. 向下引导 CTA：竖线 + 文案 `DISCOVER THE ESSENCE`

- 水平结构：
  - 文案区仍居中，但需与右上角的 AI 面板保持 24px 以上安全距；在超宽屏上可将标题组整体向左偏移 5vw，露出更多丝绸动态。

**文字样式**

- 副标题：
  - 字体：`Playfair Display` 斜体。
  - 颜色：`text-hero-sub`（较亮的冷银）。
  - 字距：`tracking-widest`，显得轻盈而高雅。  
  - 数字年份 `2025` 使用金色强调。

- 主标题（两行）：
  - `.skew-title` 样式：  
    - 大号字号（桌面端可达 `8xl–9xl`）。  
    - `skewX(-15deg)` 形成斜切效果。  
    - 复杂竖向渐变（白 → 银 → 淡蓝紫 → 皇家紫 → 透明黑），并做 `background-clip: text` 填充。  
    - 带有冷蓝外发光 `drop-shadow`，营造霓虹金属字效果。
  - 行间距：紧凑（`line-height: 0.9`），增强整体性。

- CTA 区域：
  - 一条竖直金色渐变细线（`h-16 w-[1px]`），下方放置小写字间距很大的说明文本。  
  - 文案采用 `.eyebrow` 样式：极小字号 + 超大字距，颜色为 `text-muted`，hover 变白。

**角落信息**

- 左下：`Vol. II`  
  - 使用 `font-serif italic` 与 `text-muted`，当作系列编号，弱信息。

- 右下：滚动提示  
  - 竖向金色细线（`gold-line-vertical`）+ 小字号竖排 `SCROLL`（`text-caption`）。  
  - 帮助用户理解还有内容可看。

- 右上：`Mood Weaver` 面板  
  - 贴近导航下缘，与 nav 形成 24px 间距；当滚动离开首屏，面板淡出隐藏。  
  - 移动端隐藏，改由按钮触发 modal（后续可扩展）。

---

### 2.4 Collection 区（作品展示区）

**整体容器**

- 上下内边距：`py-32`，左右：`px-6 md:px-20`。  
- 背景仍为 `var(--void-black)`，保持连续性。  
- z-index 保持在页内容层（高于背景图）。

#### 2.4.1 区域标题行

- 左侧：

  - 一条金色竖线贴左边：高于标题文字，制造“装订线”效果。  
  - 上方 Tag：使用 `.eyebrow` 显示 `Masterpieces`。  
  - 主标题 `The Collection`：衬线字体、较大字号（`text-4xl / 5xl`），白色。

- 右侧：

  - 一段简短说明文案：`text-body-copy`，小号、右对齐。  
  - 文中某句可用金色斜体突出。

#### 2.4.2 内容网格布局

布局采用 `md:grid-cols-12` 的不对称错位网格：

1. **作品 1（Moonlit Velvet）**
   - 列跨度：`md:col-span-7`，居左，占据比较大的面积。  
   - 图片比例：`aspect-[3/4]`，竖图。  
   - 默认状态：
     - 带轻微 `grayscale-[30%]`，整体偏冷，配合全局滤镜：`saturate(0.8) contrast(1.1) brightness(0.9)`。  
   - Hover：
     - 图片放大 `scale(1.05)`，恢复色彩（提升饱和度、亮度）。  
     - 边框从 `border-white/5` 增强到 `border-white/20`。  
   - 文本区域：
     - 标题：衬线斜体大字，白色，hover 变为金色。  
     - 副信息（材质/工时）：小号全大写，使用 `text-caption`。

2. **作品 2（Nebula Gown）**
   - 列跨度：`md:col-span-4 md:col-start-9`，且整体下移（通过 `md:mt-32`），与作品 1 形成错层。  
   - 图片同样为 `aspect-[3/4]`，但无灰度处理，更偏紫色体系。  
   - 标题 hover 变为淡蓝紫色，与主视觉色系呼应。

3. **作品 3（The Royal Silhouette / ETHEREAL）**
   - 列跨度：`md:col-span-10 md:col-start-2`，居中宽幅展示。  
   - 图片比例：`aspect-[16/9]`，横图，主体偏上（`object-top`）。  
   - 覆盖层：
     - 默认半透明黑蒙层（`bg-black/50`），hover 变浅。  
   - 中央叠加文字：
     - 使用 `.skew-title project-text-reveal`，混合模式 `mix-blend-overlay`，仿佛刻在图片上的字样。
   - 下方信息行：
     - 左侧：作品名 `The Royal Silhouette`。  
     - 右侧：金色圆角 Capsule 按钮 `INQUIRE`，边框为金色，hover 填充金色、文字变黑。

---

### 2.5 自定义光标与交互

**光标结构**

- 核心点： `.cursor-dot`  
  - 小圆点（6px），纯金色填充 + 金色光晕。  
- 外圈：`.cursor-outline`  
  - 直径 40px 的细描边圆环，银色边 + 内部淡紫色径向渐变。

**行为规范**

- 默认跟随：  
  - `cursor-dot` 直接跟随鼠标位置。  
  - `cursor-outline` 通过 `animate` 轻微延迟到位，形成拖尾感。

- 交互态（针对链接 / 按钮 / 卡片等可交互元素）：
  - 外圈尺寸放大至 50px。  
  - 描边颜色变为金色；内部紫色光晕增强。  

> 注意：移动端或触控设备应隐藏自定义光标，回退到系统默认。

---

### 2.6 动效与滚动

1. **首屏入场动画（GSAP Timeline）**
   - 副标题渐显 → 主标题两行从下方模糊上移 → CTA 与角落信息淡入。  
   - 主要控制节奏在 1.5–2.5s 内，避免过长阻挡信息。

2. **背景视差**
   - `#hero-bg` 在滚动时缓慢下移与缩放（`yPercent: 15, scale: 1.05`），营造空间层次。

3. **作品卡片入场**
   - 每个 `.project-item` 在进入视窗时自下而上淡入，带轻微延迟，按顺序出现。

4. **横幅大字拉伸动画**
   - `.project-text-reveal` 在进入视口时：
     - 透明度从 0 → 1  
     - 字间距从正常 → 0.15em  
   - 呈现“从画面里被拉亮、拉开”的效果。

5. **动态丝绸参数缓动**
  - 实时参数 `speed / twistSpeed / twistAmplitude / flowFrequency` 通过 `targetConfig` → `config` 的缓动插值（0.05-0.08）实现丝滑过渡。  
  - 颜色使用 `THREE.Color.lerp`，确保 base 与 glow 色彩逐帧柔和切换。  
  - AI 面板在生成期间按钮切换 `GENERATE ↔ WEAVING`，请求完成后 `ai-response` 以 1s 淡入。

---

### 2.7 页脚（Footer）

- 背景：纯黑 `bg-black`，顶部有一条轻微白色分隔线（`border-t border-white/5`）。  
- 中心添加一片紫色径向光晕背景（radial-gradient），不影响文字可读性。  
- 内容结构：
  1. 上方：`By Appointment Only`，使用金色、全大写、小字号 + 大字距。  
  2. 中部：`YOUR LEGACY`，使用 `.skew-title` 做大号标题。  
  3. 下方：社交渠道横向排列（Instagram / WeChat / Email），小号大写、`text-muted`，中间用金色小圆点分隔。  
  4. 最底部：版权信息，使用 `text-caption`。

---

## 3. 响应式与断点建议

- **桌面端（≥ 1024px）**：  
  - 按当前设计正常展示。  
  - 不对主标题和网格布局做大改动，仅控制最大宽度。

- **平板（768px–1023px）**：  
  - 保留 Hero 大标题，但可适度减小字号。  
  - Collection 区从 12 列网格回落为简单上下排列，但可保持错层顶部间距。

- **移动端（< 768px）**：  
  - 标题 skew 角度略减、字号降级，避免挤压。  
  - 所有网格项改为单列纵向排列，去掉大幅错位间距。  
  - 导航菜单折叠，视需求增加 Hamburger 菜单（当前实现中未包含，可后续扩展）。

---
