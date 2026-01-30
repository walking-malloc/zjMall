import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import {
  addCartItem,
  updateCartItemQuantity,
  removeCartItem,
  removeCartItems,
  clearCart as clearCartApi,
  getCart,
  getCartSummary,
  checkoutPreview,
  refreshCart as refreshCartApi
} from '@/api/cart'

export const useCartStore = defineStore('cart', () => {
  const items = ref([])
  const summary = ref({
    totalItems: 0,
    totalQuantity: 0,
    totalPrice: 0,
    hasInvalidItems: false
  })

  const transformItemFromApi = (apiItem) => {
    const price = parseFloat(apiItem.price || apiItem.current_price || '0') || 0
    // 确保 product_id 和 sku_id 存在
    const productId = apiItem.product_id || apiItem.productId || ''
    const skuId = apiItem.sku_id || apiItem.skuId || ''
    
    if (!productId || !skuId) {
      console.error('⚠️ 购物车项数据不完整:', {
        id: apiItem.id,
        product_id: apiItem.product_id,
        productId: apiItem.productId,
        sku_id: apiItem.sku_id,
        skuId: apiItem.skuId,
        fullItem: apiItem
      })
    }
    
    return {
      id: apiItem.id,
      productId: productId, // 确保不为 undefined
      productTitle: apiItem.product_title || apiItem.productTitle || '',
      productImage: apiItem.product_image || apiItem.productImage || '',
      skuId: skuId, // 确保不为 undefined
      skuName: apiItem.sku_name || apiItem.skuName || '',
      price,
      quantity: apiItem.quantity || 0,
      stock: apiItem.stock || 0,
      isValid: apiItem.is_valid !== false,
      invalidReason: apiItem.invalid_reason || apiItem.invalidReason || '',
      // 默认勾选有效商品
      selected: apiItem.is_valid !== false
    }
  }

  // 从后端加载购物车
  const loadCart = async () => {
    try {
      const res = await getCart()
      if (res.data && res.data.code === 0) {
        const apiItems = res.data.items || []
        console.log('购物车原始数据:', apiItems)
        items.value = apiItems.map(transformItemFromApi)
        console.log('转换后的购物车数据:', items.value.map(item => ({
          id: item.id,
          productId: item.productId,
          skuId: item.skuId,
          productTitle: item.productTitle
        })))

        if (res.data.summary) {
          const s = res.data.summary
          summary.value = {
            totalItems: s.total_items || 0,
            totalQuantity: s.total_quantity || 0,
            totalPrice: parseFloat(s.total_price || '0') || 0,
            hasInvalidItems: !!s.has_invalid_items
          }
        }
      } else {
        console.error('加载购物车失败:', res.data)
      }
    } catch (e) {
      console.error('加载购物车失败:', e)
    }
  }

  // 刷新统计信息
  const refreshSummary = async () => {
    try {
      const res = await getCartSummary()
      if (res.data && res.data.code === 0 && res.data.data) {
        const s = res.data.data
        summary.value = {
          totalItems: s.total_items || 0,
          totalQuantity: s.total_quantity || 0,
          totalPrice: parseFloat(s.total_price || '0') || 0,
          hasInvalidItems: !!s.has_invalid_items
        }
      }
    } catch (e) {
      console.error('加载购物车统计信息失败:', e)
    }
  }

  // 总数量（前端计算，主要用于展示）
  const totalCount = computed(() => {
    return items.value.reduce((sum, item) => sum + item.quantity, 0)
  })

  // 总金额（前端计算，主要用于展示）
  const totalPrice = computed(() => {
    return items.value.reduce((sum, item) => {
      return sum + item.price * item.quantity
    }, 0)
  })

  // 添加商品到购物车（调用后端）
  const addItem = async (payload) => {
    // payload: { id, skuId, quantity }
    const res = await addCartItem({
      product_id: String(payload.id),
      sku_id: payload.skuId ? String(payload.skuId) : '',
      quantity: payload.quantity || 1
    })
    
    // 检查业务错误码
    if (res.data && res.data.code !== 0) {
      const errorMessage = res.data.message || '添加到购物车失败'
      ElMessage.error(errorMessage)
      throw new Error(errorMessage)
    }
    
    ElMessage.success('已添加到购物车')
    await loadCart()
    await refreshSummary()
  }

  // 更新商品数量（调用后端）
  const updateQuantity = async (item, quantity) => {
    if (quantity <= 0) {
      await removeItem(item)
      return
    }
    await updateCartItemQuantity(item.id, quantity)
    await loadCart()
    await refreshSummary()
  }

  // 移除单个商品
  const removeItem = async (item) => {
    await removeCartItem(item.id)
    await loadCart()
    await refreshSummary()
  }

  // 批量删除
  const removeSelectedItems = async () => {
    const ids = items.value.filter(i => i.selected).map(i => i.id)
    if (!ids.length) return
    await removeCartItems(ids)
    await loadCart()
    await refreshSummary()
  }

  // 清空购物车
  const clearCart = async () => {
    await clearCartApi()
    items.value = []
    summary.value = {
      totalItems: 0,
      totalQuantity: 0,
      totalPrice: 0,
      hasInvalidItems: false
    }
  }

  // 结算预览
  const previewCheckout = async (itemIds) => {
    const res = await checkoutPreview({
      item_ids: itemIds
    })
    return res.data
  }

  // 刷新购物车（调用后端 refresh 接口，实时同步价格和库存）
  const refreshCart = async () => {
    const res = await refreshCartApi()
    if (res.data && res.data.code === 0) {
      const apiItems = res.data.items || []
      items.value = apiItems.map(transformItemFromApi)

      if (res.data.summary) {
        const s = res.data.summary
        summary.value = {
          totalItems: s.total_items || 0,
          totalQuantity: s.total_quantity || 0,
          totalPrice: parseFloat(s.total_price || '0') || 0,
          hasInvalidItems: !!s.has_invalid_items
        }
      }
    }
    return res.data
  }

  return {
    items,
    summary,
    totalCount,
    totalPrice,
    loadCart,
    refreshSummary,
    addItem,
    updateQuantity,
    removeItem,
    removeSelectedItems,
    clearCart,
    previewCheckout,
    refreshCart,
  }
})
