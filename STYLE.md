<!-- Style.md -->

# Noir Luminescence – Style / 色彩与视觉样式规范

---

## 1. 设计关键词

- **主题名称**：Noir Luminescence（暗夜流光）  
- **核心意象**：  
  - 深夜剧院的黑幕  
  - 霓虹灯在紫蓝夜色里的散射  
  - 金属文字被冷光掠过后的渐变反光  
- **整体气质**：冷调、克制、有压黑的高定感，而非明亮甜美系。

---

## 2. 色彩系统

### 2.1 全局变量（CSS Custom Properties）

```css
:root {
  /* 1. 背景：极致的黑，带极微量的蓝，避免死黑 */
  --void-black: #030305;

  /* 2. 文本体系：冷月银分级 */
  --text-hero:   #ECEFFE;  /* 重要短句 / 首屏副标题 */
  --text-body:   #C9D2E5;  /* 正文主色 */
  --text-muted:  #8B94AA;  /* 次要信息 / 导航等 */
  --text-caption:#6B7280;  /* 脚注、弱化信息 */

  /* 兼容旧命名：moon-silver 作为正文别名 */
  --moon-silver: var(--text-body);
  
  /* 3. 装饰点缀：香槟金 (克制使用) */
  --champagne-gold: #D4AF37;
  
  /* 4. 氛围光：皇家紫 (深邃) */
  --royal-purple: #4B0082;
  
  /* 5. 氛围光：午夜蓝 (冷冽) */
  --midnight-blue: #191970;
}
