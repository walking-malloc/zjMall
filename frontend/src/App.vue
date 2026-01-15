<template>
  <router-view />
</template>

<script setup>
import { onMounted } from 'vue'
import { useCartStore } from '@/stores/cart'

const cartStore = useCartStore()

onMounted(() => {
  // 初始化购物车
  cartStore.loadCart()
})
</script>

<style>
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

body {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
  background-color: #f5f5f5;
}

#app {
  min-height: 100vh;
}
</style>

