# **晚礼服官网色彩与主题规范 (Style Guide) \- Ver 2.0**

主题代号：White Phantom / 白色幻影  
美学基石：高调摄影、Linear 界面美学、珠光质感

本 Style 文档定义了全新的“白中之白”设计语言 Token。所有的颜色不再是单一的色块，而是光与材质的映射。

## **1\. 色彩体系 (Color Palette)**

我们区分“结构白”与“材质白”，通过微小的冷暖倾向来构建空间深度。

### **1.1 基础色板 (Foundation)**

| 角色 | 色彩命名 | Hex / RGBA | 描述与用途 |
| :---- | :---- | :---- | :---- |
| **画布底色** | **Atelier Mist (工坊迷雾)** | \#F9F9F9 | 极轻微的暖灰白，用于全局 Body 背景，防止纯白刺眼。 |
| **表面层** | **Pure Silk (纯丝)** | \#FFFFFF | 绝对纯白。用于卡片、图片容器，创造“悬浮”在迷雾上的感觉。 |
| **主文本** | **Charcoal Ink (炭墨)** | \#1A1A1A | 取代纯黑。深邃、柔和，像宣纸上的墨迹。 |
| **次文本** | **Stone Grey (岩石灰)** | \#595959 | 用于元数据、次级说明，保持冷调。 |

### **1.2 装饰与功能色 (Accents)**

| 角色 | 色彩命名 | Hex / RGBA | 描述与用途 |
| :---- | :---- | :---- | :---- |
| **边界线** | **Platinum Hairline (铂金丝)** | \#E5E5E5 | 用于 Linear 风格的 1px 结构性边框。 |
| **高光色** | **Champagne (香槟金)** | \#D4AF37 | **克制使用 (1%)**。仅用于 Hover 高光、价格点或焦点状态。 |
| **玻璃态** | **Frosted Vapor (磨砂蒸汽)** | rgba(255, 255, 255, 0.65) | 配合 Blur 使用，营造雾化玻璃效果。 |
| **阴影色** | **Ambient Shadow (环境影)** | rgba(200, 200, 200, 0.2) | 极淡的冷色阴影，而非黑色。 |

## **2\. 字体与排版 (Typography)**

采用“配对张力”策略：感性的衬线体标题 vs 理性的无衬线体 UI。

### **2.1 字体家族**

* **Display (标题/叙事)**: *Ogg*, *Playfair Display*, 或 *Bon Vivant*。  
  * 特征：高对比度笔触，具有书法感，优雅且锋利。  
* **Interface (UI/正文)**: *Inter*, *Switzer*, 或 *Neue Montreal*。  
  * 特征：中性、现代、高可读性，符合 Linear 风格的工程感。

### **2.2 排版层级**

* **Hero Title**: 72-96px | Serif | Italic (斜体) | Line-height 1.1  
* **Section Header**: 32-48px | Serif | Regular | Letter-spacing \-0.02em  
* **UI Label**: 12px | Sans | Uppercase (全大写) | Letter-spacing \+1.5px (宽字间距)  
* **Body Text**: 16px | Sans | Light/Regular | Line-height 1.6

## **3\. 材质与特效 (Materials & Shaders)**

### **3.1 CSS 变量定义 (Design Tokens)**

:root {  
  /\* 空间基调 \*/  
  \--bg-atelier: \#F9F9F9;  
  \--bg-card:    \#FFFFFF;  
    
  /\* 文字 \*/  
  \--text-primary:   \#1A1A1A;  
  \--text-secondary: \#595959;  
  \--text-accent:    \#D4AF37;

  /\* 边框 \- Linear 核心 \*/  
  \--border-linear: linear-gradient(135deg, rgba(0,0,0,0.1) 0%, rgba(0,0,0,0) 100%);  
  \--border-light:  \#E5E5E5;

  /\* 玻璃拟态 \*/  
  \--glass-bg:      rgba(255, 255, 255, 0.65);  
  \--glass-border:  rgba(255, 255, 255, 0.4);  
  \--glass-blur:    blur(16px);  
  \--glass-shadow:  0 8px 32px 0 rgba(31, 38, 135, 0.05);

  /\* 珠光着色器参数 (WebGL Reference) \*/  
  \--silk-sheen:    \#FFF8E7; /\* 略带暖色的高光 \*/  
  \--silk-roughness: 0.4;  
  \--silk-transmission: 0.2;  
}

### **3.2 阴影与光晕 (Light & Shadow)**

* 悬浮态 (Levitation):  
  box-shadow: 0 20px 40px \-10px rgba(200, 200, 200, 0.3);  
  创造一种物体漂浮在灯箱上的感觉，底部有漫反射光。  
* 内发光 (Inner Radiance):  
  用于激活状态的按钮或卡片。  
  box-shadow: inset 0 0 20px rgba(255, 255, 255, 0.8), 0 0 10px rgba(212, 175, 55, 0.2);

### **3.3 图标风格**

* **Stroke**: 1px 固定线宽。  
* **Style**: 几何线性，无填充。  
* **Color**: 默认炭墨色，Hover 时填充极淡的香槟金背景 rgba(212, 175, 55, 0.1)。

## **4\. 图像处理规范**

* **高调摄影 (High-Key)**: 所有产品图需在浅灰或白色背景下拍摄，直方图向右偏移。  
* **去底 (Cutout)**: 对于部分产品展示，移除背景，仅保留淡淡的接触阴影，使其与网页的“工坊迷雾”背景无缝融合。  
* **纹理叠加**: 在纯色块区域，可叠加极低透明度的“噪点 (Noise)”或“纸张纹理”，避免数码味道过重。