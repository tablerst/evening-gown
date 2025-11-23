# Evening Gown · 淡雅紫金视觉系统

> **Theme Code: Emerald Night / 暗夜森林**

本项目是一个高端晚礼服店铺的宣传页面，采用 Vue 3 + TypeScript 开发，深度贯彻 "Obsidian & Emerald" (黑曜石与祖母绿) 的视觉设计语言，注重沉浸式视觉体验与细腻的微动效。

## 📖 项目文档

详细的设计规范与开发指南请参考以下文档：

- **[🎨 UI Design Guide (DESIGN.md)](./DESIGN.md)**
  包含信息架构、Hero 首屏设计规范、排版体系及动效原则。
- **[💅 Style Guide (STYLE.md)](./STYLE.md)**
  包含色彩系统 (OKLCH)、Design Tokens 定义及主题配置说明。

## 💎 设计理念

> "黑色是舞台，不是主角；真正发光的是礼服与珠宝。"

本项目的设计核心在于**克制**与**仪式感**。通过“灰度介入”和“宝石色氛围”营造高级感，而非依赖高饱和度的色彩对撞。

- **核心关键词**：黑曜石 (Obsidian)、帝国祖母绿 (Imperial Emerald)、莫兰迪香槟 (Champagne)、古董金 (Antique Gold)。
- **视觉风格**：暗夜、丝绒感、流光丝绸、玻璃拟态。

## ✨ 核心特性

### 1. 沉浸式视觉 (Immersive Visuals)
- **Hero 首屏**: 模拟 "Obsidian Dreams" 场景，结合 Three.js 实现流动的丝绸光带效果，营造剧院暗厅般的沉浸感。
- **斜切遮罩 (Slanted Gradient Mask)**: 实现了基于 CSS Gradient 的斜切遮罩组件 (`SlantedBlock.vue`)，在图片底色上叠加渐变，确保文字清晰可读。

### 2. 细腻交互 (Micro-interactions)
- **文本动效**: 自定义文本分割工具 (`src/utils/textAnimation.ts`)，模拟 GSAP SplitText 效果。
- **材质感反馈**: 按钮与卡片的 Hover 效果模拟丝绸光带滑过或内发光，避免夸张的物理弹跳。

### 3. 主题系统 (Theming)
- 内置基于 **OKLCH** 色彩空间的主题系统。
- 定义了语义化的 CSS 变量（如 `--bg-obsidian`, `--brand-emerald`, `--text-champagne`），确保设计还原度。

## 🛠️ 技术栈

- **Core**: Vue 3, TypeScript, Vite
- **State**: Pinia
- **Styling**: Sass (SCSS), Tailwind CSS, OKLCH Color Space
- **Motion**: GSAP (ScrollTrigger, Custom SplitText)
- **3D / Visuals**: Three.js (Silk simulation)

## 🚀 快速开始

```bash
# 安装依赖
pnpm install

# 启动开发服务器
pnpm dev

# 构建生产版本
pnpm build
```
