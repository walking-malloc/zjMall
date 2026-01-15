import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export const useCartStore = defineStore('cart', () => {
  const items = ref([])

  // 从 localStorage 加载购物车数据
  const loadCart = () => {
    const saved = localStorage.getItem('cart')
    if (saved) {
      try {
        items.value = JSON.parse(saved)
      } catch (e) {
        console.error('加载购物车失败:', e)
        items.value = []
      }
    }
  }

  // 保存购物车到 localStorage
  const saveCart = () => {
    localStorage.setItem('cart', JSON.stringify(items.value))
  }

  // 总数量
  const totalCount = computed(() => {
    return items.value.reduce((sum, item) => sum + item.quantity, 0)
  })

  // 总金额
  const totalPrice = computed(() => {
    return items.value.reduce((sum, item) => {
      return sum + (item.price * item.quantity)
    }, 0)
  })

  // 添加商品到购物车
  const addItem = (product) => {
    const existingItem = items.value.find(
      item => item.productId === product.id && item.skuId === product.skuId
    )

    if (existingItem) {
      existingItem.quantity += product.quantity || 1
    } else {
      items.value.push({
        productId: product.id,
        productTitle: product.title,
        productImage: product.main_image,
        skuId: product.skuId || null,
        skuName: product.skuName || '',
        price: product.price,
        quantity: product.quantity || 1
      })
    }
    saveCart()
  }

  // 更新商品数量
  const updateQuantity = (index, quantity) => {
    if (quantity <= 0) {
      removeItem(index)
      return
    }
    items.value[index].quantity = quantity
    saveCart()
  }

  // 移除商品
  const removeItem = (index) => {
    items.value.splice(index, 1)
    saveCart()
  }

  // 清空购物车
  const clearCart = () => {
    items.value = []
    saveCart()
  }

  // 初始化时加载购物车
  loadCart()

  return {
    items,
    totalCount,
    totalPrice,
    addItem,
    updateQuantity,
    removeItem,
    clearCart,
    loadCart
  }
})

