<script setup lang="ts">
import { RouterView, RouterLink } from 'vue-router'
</script>

<template>
  <div id="app-root">
    <nav class="app-nav">
      <div class="nav-brand">Temporal Workflow Designer</div>
      <div class="nav-links">
        <RouterLink to="/" class="nav-link">Designer</RouterLink>
        <RouterLink to="/workflows" class="nav-link">Saved Workflows</RouterLink>
      </div>
    </nav>
    <RouterView />
  </div>
</template>

<style>
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

body {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
  background: #1a1a2e;
  color: #e0e0e0;
}

#app-root {
  height: 100vh;
  display: flex;
  flex-direction: column;
}

.app-nav {
  display: flex;
  align-items: center;
  padding: 0 20px;
  height: 48px;
  background: #16213e;
  border-bottom: 1px solid #0f3460;
}

.nav-brand {
  font-size: 16px;
  font-weight: 700;
  color: #e94560;
  margin-right: 32px;
}

.nav-links {
  display: flex;
  gap: 16px;
}

.nav-link {
  color: #a0a0b0;
  text-decoration: none;
  font-size: 14px;
  padding: 4px 8px;
  border-radius: 4px;
  transition: color 0.2s;
}

.nav-link:hover,
.nav-link.router-link-active {
  color: #ffffff;
}
</style>
