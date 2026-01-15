<template>
  <el-container class="app-layout">
    <Header />
    <el-main class="main-content">
      <router-view />
    </el-main>
    <Footer />
  </el-container>
</template>

<script setup>
import Header from './Header.vue'
import Footer from './Footer.vue'
</script>

<style scoped>
.app-layout {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}

.main-content {
  flex: 1;
  padding: 0;
  background: #f5f5f5;
}
</style>

