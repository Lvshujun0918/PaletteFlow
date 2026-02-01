- [ ] 修改首页文字排版
- [ ] 修复左右面板背景的圆角显示bug
- [ ] 修复panelleft显示问题
- [ ] 修复panel right标题分散问题
- [ ] 去除fontdemo相关
- [ ] 复用毛玻璃的css参数（复制粘贴实在是太笨蛋了）

```css
  background: rgba(255, 255, 255, 0.5);
  backdrop-filter: blur(10px) saturate(180%);
  -webkit-backdrop-filter: blur(10px) saturate(180%);
  border-radius: 24px;
  box-shadow: 
    0 20px 60px rgba(0, 0, 0, 0.3),
    inset 0 1px 0 rgba(255, 255, 255, 0.3);
  border: 1px solid rgba(255, 255, 255, 0.2);
```
- [ ] 本来想加几个小功能的，再说吧
  - [ ] eg:多次交互
  - [ ] 更改生成的颜色数量
  - [ ] 修改特定颜色色值
  - [ ] 预览（？
  - [ ] 文字化解释