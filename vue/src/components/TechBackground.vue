<template>
  <div class="tech-bg" aria-hidden="true">
    <div class="tech-bg__aurora" />
    <div class="tech-bg__grid" />
    <div class="tech-bg__vignette" />
    <div class="tech-bg__orbs">
      <span v-for="n in 6" :key="n" class="tech-bg__orb" :style="orbStyle(n)" />
    </div>
    <div class="tech-bg__noise" />
  </div>
</template>

<script setup>
const orbStyle = (n) => {
  const left = [12, 68, 40, 82, 24, 56][n - 1] ?? 50
  const top = [18, 22, 70, 58, 48, 84][n - 1] ?? 50
  const size = [240, 180, 320, 220, 140, 260][n - 1] ?? 220
  const delay = [0, -6, -12, -18, -9, -15][n - 1] ?? 0
  const dur = [18, 22, 26, 20, 16, 24][n - 1] ?? 20
  return {
    left: `${left}%`,
    top: `${top}%`,
    width: `${size}px`,
    height: `${size}px`,
    animationDelay: `${delay}s`,
    animationDuration: `${dur}s`,
  }
}
</script>

<style scoped>
.tech-bg {
  position: absolute;
  inset: 0;
  overflow: hidden;
  pointer-events: none;
  background:
    radial-gradient(1200px 700px at 20% 20%, rgba(56, 189, 248, 0.22), transparent 60%),
    radial-gradient(900px 600px at 85% 30%, rgba(167, 139, 250, 0.22), transparent 55%),
    radial-gradient(900px 700px at 45% 85%, rgba(34, 197, 94, 0.14), transparent 60%),
    linear-gradient(180deg, #050a16 0%, #070b1c 55%, #040814 100%);
}

.tech-bg__aurora {
  position: absolute;
  inset: -30%;
  background: conic-gradient(
    from 200deg,
    rgba(56, 189, 248, 0.0) 0deg,
    rgba(56, 189, 248, 0.16) 70deg,
    rgba(167, 139, 250, 0.20) 160deg,
    rgba(34, 197, 94, 0.12) 250deg,
    rgba(56, 189, 248, 0.0) 360deg
  );
  filter: blur(28px) saturate(1.3);
  opacity: 0.85;
  animation: auroraSpin 22s linear infinite;
}

.tech-bg__grid {
  position: absolute;
  inset: 0;
  background-image:
    linear-gradient(to right, rgba(148, 163, 184, 0.12) 1px, transparent 1px),
    linear-gradient(to bottom, rgba(148, 163, 184, 0.12) 1px, transparent 1px);
  background-size: 48px 48px;
  mask-image: radial-gradient(ellipse 70% 60% at 50% 35%, #000 58%, transparent 100%);
  opacity: 0.75;
  transform: perspective(900px) rotateX(58deg) translateY(-12%);
  transform-origin: 50% 0%;
  animation: gridFloat 10s ease-in-out infinite;
}

.tech-bg__vignette {
  position: absolute;
  inset: 0;
  background:
    radial-gradient(circle at 50% 40%, transparent 0%, rgba(0, 0, 0, 0.35) 55%, rgba(0, 0, 0, 0.75) 100%);
  pointer-events: none;
}

.tech-bg__orbs {
  position: absolute;
  inset: 0;
  pointer-events: none;
}

.tech-bg__orb {
  position: absolute;
  border-radius: 999px;
  background: radial-gradient(circle at 30% 30%, rgba(56, 189, 248, 0.45), rgba(167, 139, 250, 0.22) 45%, rgba(0, 0, 0, 0) 70%);
  filter: blur(1px);
  transform: translate(-50%, -50%);
  opacity: 0.75;
  animation: orbDrift 20s ease-in-out infinite;
  mix-blend-mode: screen;
}

.tech-bg__noise {
  position: absolute;
  inset: 0;
  background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='140' height='140'%3E%3Cfilter id='n'%3E%3CfeTurbulence type='fractalNoise' baseFrequency='.8' numOctaves='3' stitchTiles='stitch'/%3E%3C/filter%3E%3Crect width='140' height='140' filter='url(%23n)' opacity='.18'/%3E%3C/svg%3E");
  opacity: 0.08;
  mix-blend-mode: overlay;
  pointer-events: none;
}

@keyframes auroraSpin {
  0% {
    transform: rotate(0deg);
  }
  100% {
    transform: rotate(360deg);
  }
}

@keyframes gridFloat {
  0%,
  100% {
    opacity: 0.65;
    transform: perspective(900px) rotateX(58deg) translateY(-14%);
  }
  50% {
    opacity: 0.8;
    transform: perspective(900px) rotateX(58deg) translateY(-10%);
  }
}

@keyframes orbDrift {
  0%,
  100% {
    transform: translate(-50%, -50%) translate3d(-14px, 6px, 0) scale(1);
  }
  50% {
    transform: translate(-50%, -50%) translate3d(18px, -10px, 0) scale(1.08);
  }
}

@media (prefers-reduced-motion: reduce) {
  .tech-bg__aurora,
  .tech-bg__grid,
  .tech-bg__orb {
    animation: none !important;
  }
}
</style>
