# 晚礼服官网色彩与主题规范（Style Guide）

> 主题代号：**Emerald Night / 暗夜森林**

本 Style 文档用于定义整站的配色方案、色彩角色与使用比例，为 UI 与前端实现提供统一的设计 Token。

---

## 1. 品牌调性与色彩原则

- 核心场景：夜晚、剧院、精品店橱窗、深色天鹅绒。
- 关键词：**黑曜石、帝国祖母绿、莫兰迪香槟、铂金银、古董金**。
- 设计原则：
  - 黑色是舞台，不是主角；真正发光的是礼服与珠宝。
  - 金色只做“首饰”，不做“墙纸”。
  - 通过“灰度介入”和“宝石色氛围”获得高级感，而不是靠高饱和对撞。

---

## 2. 基础色板（Design Tokens）

### 2.1 CSS 变量命名（推荐）

```css
:root {
  /* 基底色 - 材质感黑 */
  --bg-obsidian: #050505;        /* Obsidian Void，全局背景 */
  --bg-surface:  #08140F;        /* Deepest Forest，卡片 / 面板 */

  /* 品牌主色 - 帝国祖母绿 */
  --brand-emerald:        #043D2C;
  --brand-emerald-hover:  #056E41;

  /* 文字与中性色 - 莫兰迪体系 */
  --text-champagne: #E8D6B3;     /* 一级文字 / 重点信息 */
  --text-platinum:  #C0C0C0;     /* 正文 / 次级信息 */

  /* 点缀色 - 金银珠宝 */
  --accent-gold:   #D4AF37;      /* 古董金：边框 / Icon / Logo */
  --accent-silver: #E0E0E0;      /* 星光银：图标 / 高光 */

  /* 材质滤镜 */
  --glass-blur:      blur(20px);
  --velvet-gradient: radial-gradient(circle,
                                     var(--brand-emerald-hover),
                                     var(--bg-obsidian));
}
