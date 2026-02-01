<template>
  <div class="font-demo">
    <h2>字体效果演示 (FontDemo)</h2>

    <section class="demo-row">
      <div class="card">
        <h3>Serif / 衬线</h3>
        <p class="serif">Playfair Display — 优雅的衬线字体示例</p>
      </div>

      <div class="card">
        <h3>Sans / 无衬线</h3>
        <p class="sans">Inter 系列 — 干净的无衬线示例</p>
      </div>
    </section>

    <section class="demo-row">
      <div class="card">
        <h3>Weight / 粗细</h3>
        <p class="w300">Light 300</p>
        <p class="w500">Regular 500</p>
        <p class="w800">Bold 800</p>
      </div>

      <div class="card">
        <h3>Letter-spacing / 字母间距</h3>
        <p class="tight">Tight spacing</p>
        <p class="normal">Normal spacing</p>
        <p class="wide">Wide spacing</p>
      </div>
    </section>

    <section class="demo-row">
      <div class="card">
        <h3>Text shadow / 阴影</h3>
        <p class="ts-soft">Soft shadow</p>
        <p class="ts-strong">Strong shadow</p>
      </div>

      <div class="card">
        <h3>Gradient text / 渐变文字</h3>
        <p class="grad">Gradient color text demo</p>
      </div>
    </section>

    <section class="demo-row">
      <div class="card">
        <h3>Stroke / 描边</h3>
        <p class="stroke">Outlined text</p>
      </div>

      <div class="card">
        <h3>3D / 立体变换</h3>
        <p class="three-d">3D transformed text</p>
      </div>
    </section>

    <section class="demo-row">
      <div class="card full">
        <h3>Animated gradient / 动态渐变</h3>
        <p class="anim-grad">Animated gradient text demo</p>
      </div>
    </section>

    <router-link to="/" class="back">← 返回首页</router-link>
  </div>
</template>

<script>
export default {
  name: 'FontDemo'
}
</script>

<style scoped>
@import url('https://fonts.googleapis.com/css2?family=Inter:wght@300;500;800&family=Playfair+Display:wght@400;700&display=swap');

.font-demo { padding: 28px; max-width: 1000px; margin: 0 auto; color: #222; }
.font-demo h2 { font-family: 'Playfair Display', serif; color: #A4B3B3; margin-bottom: 18px; }
.demo-row { display: flex; gap: 20px; margin-bottom: 20px; flex-wrap: wrap; }
.card { flex: 1 1 300px; background: rgba(255,255,255,0.7); padding: 16px; border-radius: 12px; border: 1px solid rgba(164,179,179,0.2); }
.card.full { flex-basis: 100%; }
.card h3 { margin: 0 0 8px 0; color: #6b7b7b; font-size: 0.95rem; }

.serif { font-family: 'Playfair Display', Georgia, serif; font-size: 1.25rem; }
.sans { font-family: 'Inter', system-ui, -apple-system, 'Segoe UI', Roboto, 'Helvetica Neue', Arial; font-size: 1.1rem; }

.w300 { font-weight: 300; }
.w500 { font-weight: 500; }
.w800 { font-weight: 800; }

.tight { letter-spacing: -0.5px; }
.normal { letter-spacing: 0.5px; }
.wide { letter-spacing: 3px; }

.ts-soft { text-shadow: 0 2px 6px rgba(0,0,0,0.12); }
.ts-strong { text-shadow: 0 6px 18px rgba(0,0,0,0.18); }

.grad { background: linear-gradient(90deg,#A4B3B3,#96a9c1); -webkit-background-clip: text; background-clip: text; -webkit-text-fill-color: transparent; font-weight:700; }

.stroke { color: transparent; -webkit-text-stroke: 1px #A4B3B3; font-weight:700; }

.three-d { transform: perspective(400px) rotateX(10deg) translateZ(10px); text-shadow: 0 10px 20px rgba(0,0,0,0.15); display:inline-block; }

.anim-grad { font-weight:800; font-size:1.6rem; background:linear-gradient(270deg,#A4B3B3,#c7d2d2,#9fb0b0); background-size:600% 600%; -webkit-background-clip:text; background-clip:text; -webkit-text-fill-color:transparent; animation: gradAnim 6s ease infinite; }
@keyframes gradAnim { 0%{background-position:0% 50%}50%{background-position:100% 50%}100%{background-position:0% 50%} }

.back { display:inline-block; margin-top: 20px; color: #667; }

</style>
