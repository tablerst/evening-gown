# Evening Gown · 淡雅紫金视觉系统

本项目是一个高端晚礼服店铺的宣传页面，采用 Vue 3 + TypeScript 开发，注重视觉设计与动效体验。

## 📚 开发文档

- **[DESIGN.md](./DESIGN.md)**: 原始视觉设计稿与 Design Tokens 定义。
- **[STYLE.md](./STYLE.md)**: 开发风格指南，包含 Sass 变量、色彩系统 (OKLCH) 与主题配置说明。

## 🛠️ 技术栈

- **Framework**: Vue 3, Pinia, Vue Router
- **Build Tool**: Vite
- **Styling**: Sass (SCSS), OKLCH Color Space
- **Motion**: GSAP (ScrollTrigger, SplitText implementation)
- **3D**: Three.js

## ✨ 新特性

### 1. 视觉交互
- **斜切遮罩 (Slanted Gradient Mask)**: 实现了基于 CSS Gradient 的斜切遮罩效果 (`SlantedBlock.vue`)，支持图片底色上叠加渐变遮罩，确保文字清晰可见，替代了简单的 `clip-path` 裁剪方案。
- **微动效 (Micro-interactions)**: 
  - 实现了自定义的文本分割工具 (`src/utils/textAnimation.ts`)，模拟 GSAP SplitText 效果。
  - 标题文字入场采用字符级交错动画 (Staggered Character Animation)。

## 🎨 主题系统

本项目内置了一套基于 OKLCH 的主题系统，支持 **Light** (默认) 和 **Dark** 模式。
详情请参考 [STYLE.md](./STYLE.md#32-主题系统-theming)。

## 🚀 快速开始

```bash
# 安装依赖
pnpm install

# 启动开发服务器
pnpm dev

# 构建生产版本
pnpm build
```
