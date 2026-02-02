<template>
  <div>
    <div v-if="$route && $route.path === '/'" class="app-container">
    <!-- 背景图片轮播 -->
    <div class="background-carousel">
      <div 
        v-for="(image, index) in backgroundImages" 
        :key="index"
        :class="['bg-image', { active: currentBgIndex === index }]"
        :style="{ backgroundImage: `url(${image})` }"
      ></div>
    </div>

    <!-- 主内容容器 -->
    <div class="main-content glass-surface">
      <!-- Logo -->
      <div class="logo-container glass-card">
        <img 
          :src="logoUrl" 
          alt="Logo" 
          class="logo"
          @error="handleLogoError"
        >
      </div>

      <!-- 标题 -->
      <h1 class="main-title">{{ appTitle }}</h1>

      <!-- 副标题/描述 -->
      <p class="subtitle">{{ appSubtitle }}</p>

      <!-- 进入主页面按钮 -->
      <div class="button-container">
        <GlassButton
          class="enter-button"
          @click="enterMainPage"
          @mouseenter="hoverButton = true"
          @mouseleave="hoverButton = false"
        >
          <span class="button-text">{{ buttonText }}</span>
          <span class="button-icon">→</span>
        </GlassButton>
      </div>

      <!-- 底部Slogan -->
      <div class="slogan-container">
        <p class="slogan">{{ slogan }}</p>
        <div class="slogan-decoration"></div>
      </div>
    </div>

    <!-- 加载指示器（可选） -->
    <div class="loading-indicator" v-if="isLoading">
      <div class="spinner"></div>
      <p>正在进入主页面...</p>
    </div>
    </div>

    <transition name="route-fade" mode="out-in" v-else>
      <router-view />
    </transition>
  </div>
</template>

<script>
import bg6 from './assets/bg6.png'
import bg1 from './assets/bg1.png'
import bg2 from './assets/bg2.png'
import bg3 from './assets/bg3.png'
import bg4 from './assets/bg4.png'
import bg5 from './assets/bg5.png'
import logo from './assets/logo.png'
import GlassButton from './components/GlassButton.vue'

export default {
  name: 'App',
  components: {
    GlassButton
  },
  data() {
    return {

      backgroundImages: [
        bg6,
        bg1,
        bg2,
        bg3,
        bg4,
        bg5
      ],
      currentBgIndex: 0,
      intervalId: null,
      
      appTitle: 'PaletteFlow',
      appSubtitle: '自然语言生成配色',
      slogan: '配色，易如反掌',
      
      logoUrl: logo,
      
      buttonText: '开始探索',
      hoverButton: false,
      isLoading: false
    }
  },
  computed: {
    
    nextBgIndex() {
      return (this.currentBgIndex + 1) % this.backgroundImages.length
    },
    // 标题渐变色随背景索引变化需要 transition配合）
    titleGradient() {
      const gradients = [
        'linear-gradient(135deg, #8b9ce0 0%, #9b7db5 100%)',  // bg6 - 柔和
        'linear-gradient(135deg, #f5a8cc 0%, #f59a93 100%)',  // bg1 - 柔和
        'linear-gradient(135deg, #7fc9f0 0%, #62e0e0 100%)',  // bg2 - 柔和
        'linear-gradient(135deg, #8ad9a3 0%, #79dfc9 100%)',  // bg3 - 柔和
        'linear-gradient(135deg, #f5a4b8 0%, #f5d9a3 100%)',  // bg4 - 柔和
        'linear-gradient(135deg, #7dd8d8 0%, #9b7db5 100%)'   // bg5 - 柔和
      ]
      return gradients[this.currentBgIndex]
    },
    // 按钮渐变色
    buttonGradient() {
      const gradients = [
        'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',  // bg6
        'linear-gradient(135deg, #f093fb 0%, #f5576c 100%)',  // bg1
        'linear-gradient(135deg, #4facfe 0%, #00f2fe 100%)',  // bg2
        'linear-gradient(135deg, #43e97b 0%, #38f9d7 100%)',  // bg3
        'linear-gradient(135deg, #fa709a 0%, #fee140 100%)',  // bg4
        'linear-gradient(135deg, #30cfd0 0%, #330867 100%)'   // bg5
      ]
      return gradients[this.currentBgIndex]
    }
  },
  methods: {
    // 进入主页面
    async enterMainPage() {
      this.isLoading = true
      
      // 模拟网络请求延迟
      await new Promise(resolve => setTimeout(resolve, 800))
      
      // 实际导航逻辑
      if (this.$router) {
        this.$router.push('/feature')
      } else {
        // 如果没有路由，可以跳转或其他处理
        window.location.href = '/home.html'
      }
      
      this.isLoading = false
    },
    
    // Logo加载失败处理
    handleLogoError(event) {
      console.warn('Logo加载失败，使用默认图标')
      event.target.src = 'data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iOTYiIGhlaWdodD0iOTYiIHZpZXdCb3g9IjAgMCAyNCAyNCIgZmlsbD0ibm9uZSIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj48cGF0aCBkPSJNNyAyMUg1QTUgNSAwIDAxNSA3aDNhNSA1IDAgMDE0IDEuNDg5VjNIMTNWOEg5LjcxMWE3IDcgMCAxMDYuMjcyIDYuNzI3TDIwIDEzbDItM0w5LjM5IDkuMzlBOSA5IDAgMDA1IDE5YTYgNiAwIDAwNiA2aDR2LTJoLTR6IiBmaWxsPSIjNjM2M0Y5Ii8+PC9zdmc+'
    },
    
    // 切换背景图片
    changeBackground() {
      this.currentBgIndex = (this.currentBgIndex + 1) % this.backgroundImages.length
    }
  },
  mounted() {
    // 开始背景图片轮播
    this.intervalId = setInterval(this.changeBackground, 3000)
    
    // 监听路由变化
    if (this.$route && this.$route.path !== '/') {
      this.enterMainPage()
    }
  },
  beforeUnmount() {
    // 清理定时器
    if (this.intervalId) {
      clearInterval(this.intervalId)
    }
  }
}
</script>

<style scoped>
@import url('https://fonts.font.im/css?family=Playfair+Display:600,700');
.app-container {
  position: relative;
  width: 100vw;
  height: 100vh;
  overflow: hidden;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, sans-serif;
  display: flex;
  justify-content: center;
  align-items: center;
}

/* 背景图片轮播样式 */
.background-carousel {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  z-index: 1;
}

.bg-image {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background-size: cover;
  background-position: center;
  background-repeat: no-repeat;
  opacity: 0;
  transition: opacity 2s ease-in-out;
  filter: blur(2px) brightness(0.7);
}

.bg-image.active {
  opacity: 1;
}

/* 主内容容器 - 半透明Docker效果 */
.main-content {
  display: flex;
  flex-direction: column;
  position: relative;
  z-index: 2;
  max-width: 800px;
  width: 90%;
  height: 60vh;
  padding: 60px 60px 10px 60px;
  text-align: center;
  animation: float 6s ease-in-out infinite;
  justify-content: space-between;
}

@keyframes float {
  0%, 100% { transform: translateY(0px); }
  50% { transform: translateY(-10px); }
}

/* Logo样式 */
.logo-container {
  animation: fadeInDown 1s ease;
  width: 110px;
  height: 110px;
  flex-shrink: 0;
  aspect-ratio: 1 / 1;
  display: flex;
  justify-content: center;
  align-items: center;
  margin-left: auto;
  margin-right: auto;
  border-radius: 30px;
  transition: all 0.4s cubic-bezier(0.4, 0, 0.2, 1);
}

.logo-container:hover {
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.12);
}

.logo {
  width: 96px;
  height: 96px;
  object-fit: contain;
  filter: drop-shadow(0 4px 12px rgba(0, 0, 0, 0.2));
  transition: transform 0.5s ease, filter 0.5s ease;
}

.logo:hover {
  transform: scale(1.1) rotate(5deg);
  filter: drop-shadow(0 6px 20px rgba(0, 0, 0, 0.3));
}

/* 标题样式 */
.main-title {
  font-size: 3.5rem;
  font-weight: 600;
  color: #333333;
  font-family: 'Playfair Display', Georgia, 'Times New Roman', serif;
  text-shadow: 0 2px 4px rgba(0, 0, 0, 0.1), 0 -1px 2px rgba(255, 255, 255, 0.3);
  animation: fadeInUp 1s ease 0.2s both;
  letter-spacing: -0.5px;
  transition: color 0.4s ease, text-shadow 0.4s ease;
}

.subtitle {
  font-size: 1.2rem;
  color: #a0a0a0;
  animation: fadeInUp 1s ease 0.4s both;
  font-weight: 300;
  letter-spacing: 2px;
  flex-direction: row;
}

/* 按钮样式 */
.button-container {
  animation: fadeInUp 1s ease 0.6s both;
}

.enter-button {
  position: relative;
  padding: 1.2rem 3.5rem;
  font-size: 1.1rem;
  gap: 1rem;
  overflow: hidden;
}

.button-icon {
  font-size: 1.4rem;
  transition: transform 0.3s ease;
}

.enter-button:hover .button-icon {
  transform: translateX(5px);
}

/* 底部Slogan */
.slogan-container {
  animation: fadeIn 1.5s ease 0.8s both;
}

.slogan {
  font-size: 1.1rem;
  color: #888;
  font-weight: 300;
  letter-spacing: 3px;
  text-transform: uppercase;
  margin: 0;
  position: relative;
  display: inline-block;
  padding: 0 1rem;
}
/* Slogan装饰线 */
.slogan-decoration {
  position: absolute;
  bottom: -10px;
  left: 50%;
  transform: translateX(-50%);
  width: 60px;
  height: 2px;
  background: linear-gradient(90deg, transparent, #667eea, transparent);
}

/* 加载指示器 */
.loading-indicator {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background: rgba(255, 255, 255, 0.85);
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  z-index: 1000;
  backdrop-filter: blur(10px);
}

.spinner {
  width: 50px;
  height: 50px;
  border: 3px solid #f3f3f3;
  border-top: 3px solid #667eea;
  border-radius: 50%;
  animation: spin 1s linear infinite;
  margin-bottom: 1rem;
}

@keyframes spin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}

/* 动画 */
@keyframes fadeInDown {
  from {
    opacity: 0;
    transform: translateY(-30px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

@keyframes fadeInUp {
  from {
    opacity: 0;
    transform: translateY(30px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

@keyframes fadeIn {
  from {
    opacity: 0;
  }
  to {
    opacity: 1;
  }
}
</style>
