<template>
  <div class="cart-page">
    <div class="cart-container">
      <h2>购物车</h2>
      
      <el-card v-if="cartStore.items.length === 0" class="empty-cart">
        <el-empty description="购物车是空的">
          <el-button type="primary" @click="$router.push('/product/products')">
            去购物
          </el-button>
        </el-empty>
      </el-card>

      <el-row :gutter="20" v-else>
        <el-col :span="18">
          <el-card>
            <template #header>
              <div class="cart-header">
                <el-checkbox v-model="selectAll" @change="handleSelectAll">全选</el-checkbox>
                <span>商品信息</span>
                <span>单价</span>
                <span>数量</span>
                <span>小计</span>
                <span>操作</span>
              </div>
            </template>

            <div class="cart-items">
              <div 
                v-for="item in cartStore.items" 
                :key="item.id"
                class="cart-item"
              >
                <el-checkbox 
                  v-model="item.selected" 
                  @change="updateSelected"
                />
                <img :src="item.productImage || '/placeholder.png'" class="product-image" />
                <div class="product-info">
                  <h4>{{ item.productTitle }}</h4>
                  <p v-if="item.skuName" class="sku-name">{{ item.skuName }}</p>
                </div>
                <div class="price">¥{{ item.price.toFixed(2) }}</div>
                <div class="quantity">
                  <el-input-number
                  v-model="item.quantity"
                    :min="1"
                    :max="999"
                  @change="(val) => handleQuantityChange(item, val)"
                  />
                </div>
                <div class="subtotal">¥{{ (item.price * item.quantity).toFixed(2) }}</div>
                <el-button 
                  type="danger" 
                  text 
                  @click="handleRemove(item)"
                >
                  删除
                </el-button>
              </div>
            </div>
          </el-card>
        </el-col>

        <el-col :span="6">
          <el-card class="cart-summary">
            <template #header>
              <h3>结算</h3>
            </template>
            <div class="summary-item">
              <span>已选商品：</span>
              <span>{{ selectedCount }} 件</span>
            </div>
            <div class="summary-item">
              <span>合计：</span>
              <span class="total-price">¥{{ selectedTotal.toFixed(2) }}</span>
            </div>
            <el-button 
              type="primary" 
              size="large" 
              :disabled="selectedCount === 0"
              @click="handleCheckout"
              style="width: 100%; margin-top: 20px;"
            >
              去结算
            </el-button>
          </el-card>
        </el-col>
      </el-row>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useCartStore } from '@/stores/cart'
import { ElMessage } from 'element-plus'

const router = useRouter()
const cartStore = useCartStore()

const selectAll = ref(false)

const selectedItems = computed(() => {
  return cartStore.items.filter(item => item.selected)
})

const selectedCount = computed(() => {
  return selectedItems.value.reduce((sum, item) => sum + item.quantity, 0)
})

const selectedTotal = computed(() => {
  return selectedItems.value.reduce((sum, item) => {
    return sum + (item.price * item.quantity)
  }, 0)
})

const handleSelectAll = (val) => {
  cartStore.items.forEach(item => {
    item.selected = val
  })
}

const updateSelected = () => {
  selectAll.value = cartStore.items.every(item => item.selected)
}

const handleQuantityChange = async (item, quantity) => {
  try {
    await cartStore.updateQuantity(item, quantity)
  } catch (error) {
    console.error('更新数量失败:', error)
    ElMessage.error('更新数量失败，请稍后重试')
  }
}

const handleRemove = async (item) => {
  try {
    await cartStore.removeItem(item)
    ElMessage.success('已删除该商品')
  } catch (error) {
    console.error('删除商品失败:', error)
    ElMessage.error('删除商品失败，请稍后重试')
  }
}

const handleCheckout = async () => {
  if (!selectedItems.value.length) {
    ElMessage.warning('请选择要结算的商品')
    return
  }

  try {
    const itemIds = selectedItems.value.map(item => item.id)
    const res = await cartStore.previewCheckout(itemIds)
    if (res && res.code === 0) {
      console.log('结算预览结果:', res.data)
      ElMessage.success('已完成结算预览，后续可接入订单创建流程')
    } else {
      ElMessage.error(res?.message || '结算预览失败')
    }
  } catch (error) {
    console.error('结算预览失败:', error)
    ElMessage.error('结算预览失败，请稍后重试')
  }
}

onMounted(async () => {
  await cartStore.loadCart()
  await cartStore.refreshSummary()
})
</script>

<style scoped>
.cart-page {
  min-height: calc(100vh - 140px);
  padding: 20px;
  background: #f5f5f5;
}

.cart-container {
  max-width: 1200px;
  margin: 0 auto;
}

.cart-container h2 {
  margin-bottom: 20px;
  font-size: 24px;
}

.empty-cart {
  text-align: center;
  padding: 60px 0;
}

.cart-header {
  display: grid;
  grid-template-columns: 60px 2fr 100px 120px 100px 80px;
  gap: 20px;
  align-items: center;
  font-weight: bold;
}

.cart-items {
  min-height: 200px;
}

.cart-item {
  display: grid;
  grid-template-columns: 60px 80px 2fr 100px 120px 100px 80px;
  gap: 20px;
  align-items: center;
  padding: 20px 0;
  border-bottom: 1px solid #eee;
}

.cart-item:last-child {
  border-bottom: none;
}

.product-image {
  width: 80px;
  height: 80px;
  object-fit: cover;
  border-radius: 4px;
}

.product-info h4 {
  margin: 0 0 5px 0;
  font-size: 14px;
}

.sku-name {
  color: #999;
  font-size: 12px;
  margin: 0;
}

.price {
  color: #f56c6c;
  font-weight: bold;
}

.subtotal {
  color: #f56c6c;
  font-weight: bold;
  font-size: 16px;
}

.cart-summary {
  position: sticky;
  top: 20px;
}

.summary-item {
  display: flex;
  justify-content: space-between;
  margin-bottom: 15px;
  font-size: 14px;
}

.total-price {
  color: #f56c6c;
  font-size: 24px;
  font-weight: bold;
}
</style>

