<template>
  <div class="checkout-page">
    <div class="checkout-container">
      <h2>确认订单</h2>

      <!-- 收货地址 -->
      <el-card class="address-card" style="margin-bottom: 20px;">
        <template #header>
          <div class="card-header">
            <span>收货地址</span>
            <el-button text type="primary" @click="showAddressDialog = true">选择地址</el-button>
          </div>
        </template>
        <div v-if="selectedAddress">
          <p><strong>{{ selectedAddress.receiverName }}</strong> {{ selectedAddress.receiverPhone }}</p>
          <p>{{ selectedAddress.province }}{{ selectedAddress.city }}{{ selectedAddress.district }}{{ selectedAddress.detail }}</p>
        </div>
        <el-empty v-else description="请选择收货地址" :image-size="80" />
      </el-card>

      <!-- 商品信息 -->
      <el-card class="items-card" style="margin-bottom: 20px;">
        <template #header>
          <span>商品信息</span>
        </template>
        <div class="order-items">
          <div 
            v-for="item in checkoutItems" 
            :key="item.id"
            class="order-item"
          >
            <img :src="item.productImage || '/placeholder.png'" class="item-image" />
            <div class="item-info">
              <h4>{{ item.productTitle }}</h4>
              <p v-if="item.skuName" class="sku-name">{{ item.skuName }}</p>
            </div>
            <div class="item-price">¥{{ item.price.toFixed(2) }}</div>
            <div class="item-quantity">x{{ item.quantity }}</div>
            <div class="item-subtotal">¥{{ (item.price * item.quantity).toFixed(2) }}</div>
          </div>
        </div>
      </el-card>

      <!-- 订单备注 -->
      <el-card class="remark-card" style="margin-bottom: 20px;">
        <template #header>
          <span>订单备注</span>
        </template>
        <el-input 
          v-model="buyerRemark" 
          type="textarea" 
          :rows="3"
          placeholder="选填，可填写特殊要求"
          maxlength="200"
          show-word-limit
        />
      </el-card>

      <!-- 订单汇总 -->
      <el-card class="summary-card">
        <div class="summary-item">
          <span>商品总价：</span>
          <span>¥{{ totalAmount.toFixed(2) }}</span>
        </div>
        <div class="summary-item">
          <span>运费：</span>
          <span>¥{{ shippingAmount.toFixed(2) }}</span>
        </div>
        <div class="summary-item">
          <span>优惠：</span>
          <span class="discount">-¥{{ discountAmount.toFixed(2) }}</span>
        </div>
        <el-divider />
        <div class="summary-item total">
          <span>应付总额：</span>
          <span class="total-price">¥{{ payAmount.toFixed(2) }}</span>
        </div>
        <el-button 
          type="primary" 
          size="large" 
          :loading="creating"
          :disabled="!selectedAddress || checkoutItems.length === 0"
          @click="handleCreateOrder"
          style="width: 100%; margin-top: 20px;"
        >
          提交订单
        </el-button>
      </el-card>

      <!-- 地址选择对话框 -->
      <el-dialog v-model="showAddressDialog" title="选择收货地址" width="600px" @open="loadAddresses">
        <div v-loading="loadingAddresses">
          <el-radio-group v-model="selectedAddressId" @change="handleAddressChange">
            <div 
              v-for="address in addresses" 
              :key="address.id"
              class="address-option"
            >
              <el-radio :label="address.id">
                <div>
                  <p><strong>{{ address.receiverName }}</strong> {{ address.receiverPhone }}</p>
                  <p>{{ address.province }}{{ address.city }}{{ address.district }}{{ address.detail }}</p>
                </div>
              </el-radio>
            </div>
          </el-radio-group>
          <el-empty v-if="!loadingAddresses && addresses.length === 0" description="暂无收货地址">
            <el-button type="primary" @click="$router.push('/addresses')">去添加地址</el-button>
          </el-empty>
        </div>
        <template #footer>
          <el-button @click="showAddressDialog = false">取消</el-button>
          <el-button type="primary" @click="handleConfirmAddress">确定</el-button>
          <el-button type="primary" text @click="$router.push('/addresses')">管理地址</el-button>
        </template>
      </el-dialog>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { generateOrderToken, createOrder } from '@/api/order'
import { getAddressList } from '@/api/user'
import { useCartStore } from '@/stores/cart'

const router = useRouter()
const route = useRoute()
const cartStore = useCartStore()

// 从路由参数或购物车获取商品列表
const checkoutItems = ref([])
const selectedAddress = ref(null)
const selectedAddressId = ref('')
const addresses = ref([])
const showAddressDialog = ref(false)
const loadingAddresses = ref(false)
const buyerRemark = ref('')
const creating = ref(false)

// 金额计算
const totalAmount = computed(() => {
  return checkoutItems.value.reduce((sum, item) => {
    return sum + (item.price * item.quantity)
  }, 0)
})

const discountAmount = ref(0) // TODO: 优惠金额
const shippingAmount = ref(0) // TODO: 运费

const payAmount = computed(() => {
  return Math.max(0, totalAmount.value - discountAmount.value + shippingAmount.value)
})

// 加载地址列表
const loadAddresses = async () => {
  loadingAddresses.value = true
  try {
    console.log('开始加载地址列表...')
    const res = await getAddressList()
    console.log('地址列表API响应:', res)
    
    if (res.data && res.data.code === 0) {
      // 根据实际API返回的数据结构调整（proto 返回的是 data 字段）
      const addressList = res.data.data || res.data.addresses || []
      console.log('原始地址数据:', addressList)
      
      // 转换地址数据格式，兼容不同的字段命名
      addresses.value = addressList.map(addr => ({
        id: addr.id || addr.address_id,
        receiverName: addr.receiver_name || addr.receiverName,
        receiverPhone: addr.receiver_phone || addr.receiverPhone,
        province: addr.province || '',
        city: addr.city || '',
        district: addr.district || '',
        detail: addr.detail || '',
        postalCode: addr.postal_code || addr.postalCode,
        isDefault: addr.is_default !== undefined ? addr.is_default : (addr.isDefault !== undefined ? addr.isDefault : false)
      }))
      
      console.log('解析后的地址列表:', addresses.value)
      
      if (addresses.value.length === 0) {
        console.warn('地址列表为空')
        ElMessage.warning('您还没有收货地址，请先添加收货地址')
        return
      }
      
      // 如果有默认地址，自动选择
      const defaultAddress = addresses.value.find(addr => 
        addr.isDefault === true || addr.isDefault === 'true' || addr.isDefault === 1
      )
      if (defaultAddress) {
        selectedAddressId.value = defaultAddress.id
        selectedAddress.value = defaultAddress
        console.log('已选择默认地址:', defaultAddress)
      } else if (addresses.value.length > 0) {
        // 如果没有默认地址，选择第一个
        selectedAddressId.value = addresses.value[0].id
        selectedAddress.value = addresses.value[0]
        console.log('已选择第一个地址:', addresses.value[0])
      }
    } else {
      const errorMsg = res.data?.message || '获取地址列表失败'
      console.error('地址列表API返回错误:', res.data)
      ElMessage.error(errorMsg)
    }
  } catch (error) {
    console.error('加载地址列表异常:', error)
    console.error('错误详情:', {
      message: error.message,
      response: error.response?.data,
      status: error.response?.status,
      statusText: error.response?.statusText
    })
    
    if (error.response?.status === 401) {
      ElMessage.error('登录已过期，请重新登录')
      router.push('/login')
    } else if (error.response?.status === 403) {
      ElMessage.error('没有权限访问地址信息')
    } else {
      ElMessage.error('加载地址列表失败: ' + (error.message || '未知错误'))
    }
  } finally {
    loadingAddresses.value = false
  }
}

// 地址选择变化
const handleAddressChange = (addressId) => {
  const address = addresses.value.find(addr => addr.id === addressId)
  if (address) {
    selectedAddress.value = address
  }
}

// 确认选择地址
const handleConfirmAddress = () => {
  if (selectedAddressId.value) {
    handleAddressChange(selectedAddressId.value)
    showAddressDialog.value = false
  } else {
    ElMessage.warning('请选择一个收货地址')
  }
}

// 创建订单
const handleCreateOrder = async () => {
  // 验证地址
  if (!selectedAddress.value || !selectedAddress.value.id) {
    ElMessage.warning('请选择收货地址')
    // 如果地址列表为空，重新加载
    if (addresses.value.length === 0) {
      await loadAddresses()
    }
    showAddressDialog.value = true
    return
  }

  if (checkoutItems.value.length === 0) {
    ElMessage.warning('没有可结算的商品')
    return
  }

  creating.value = true

  try {
    console.log('开始创建订单，当前选中的地址:', selectedAddress.value)
    
    // 检查 Token 是否存在
    const authToken = localStorage.getItem('token')
    if (!authToken || authToken.trim() === '') {
      ElMessage.error('登录已过期，请重新登录')
      router.push('/login')
      return
    }
    
    // 1. 生成Token
    let tokenRes
    try {
      tokenRes = await generateOrderToken()
    } catch (error) {
      console.error('生成订单Token失败:', error)
      if (error.response?.status === 401) {
        ElMessage.error('登录已过期，请重新登录')
        localStorage.removeItem('token')
        router.push('/login')
        return
      }
      ElMessage.error('生成订单Token失败: ' + (error.message || '未知错误'))
      return
    }
    
    if (tokenRes.data.code !== 0) {
      ElMessage.error(tokenRes.data.message || '生成Token失败')
      // 如果是未登录错误，跳转到登录页
      if (tokenRes.data.code === 401 || tokenRes.data.message?.includes('未登录')) {
        localStorage.removeItem('token')
        router.push('/login')
      }
      return
    }
    const token = tokenRes.data.token

    // 2. 创建订单
    // 验证商品数据
    const orderItems = checkoutItems.value.map(item => {
      if (!item.productId || !item.skuId) {
        console.error('商品数据不完整:', item)
        throw new Error(`商品 ${item.productTitle || item.id} 数据不完整`)
      }
      return {
        cart_item_id: item.id, // 购物车项ID，用于订单创建成功后删除购物车项
        product_id: item.productId,
        sku_id: item.skuId,
        quantity: item.quantity
      }
    })
    
    // 再次确认地址ID存在
    const addressId = selectedAddress.value.id
    if (!addressId) {
      ElMessage.error('地址ID无效，请重新选择地址')
      await loadAddresses()
      showAddressDialog.value = true
      return
    }
    
    console.log('创建订单请求数据:', {
      items: orderItems,
      addressId: addressId,
      token: token
    })
    
    let orderRes
    try {
      orderRes = await createOrder({
        items: orderItems,
        addressId: addressId,
        buyerRemark: buyerRemark.value,
        token: token
      })
    } catch (error) {
      console.error('创建订单请求失败:', error)
      if (error.response?.status === 401) {
        ElMessage.error('登录已过期，请重新登录')
        localStorage.removeItem('token')
        router.push('/login')
        return
      }
      throw error
    }

    if (orderRes.data.code === 0) {
      console.log('订单创建成功，响应数据:', orderRes.data)
      
      // 兼容两种字段命名：snake_case 和 camelCase
      const orderNo = orderRes.data.order_no || orderRes.data.orderNo || ''
      const payAmount = orderRes.data.pay_amount || orderRes.data.payAmount || '0.00'
      
      ElMessage.success('订单创建成功！')
      
      // 注意：购物车项删除已由后端在订单创建成功后自动处理
      // 这里只需要刷新购物车数据，确保前端显示与后端一致
      await cartStore.loadCart()

      // 跳转到订单详情或订单列表
      ElMessageBox.confirm(
        `订单创建成功！订单号：${orderNo}，支付金额：¥${payAmount}`,
        '提示',
        {
          confirmButtonText: '查看订单',
          cancelButtonText: '继续购物',
          type: 'success'
        }
      ).then(() => {
        router.push('/orders')
      }).catch(() => {
        router.push('/product/products')
      })
    } else {
      ElMessage.error(orderRes.data.message || '订单创建失败')
    }
  } catch (error) {
    console.error('创建订单失败:', error)
    console.error('错误详情:', {
      message: error.message,
      response: error.response?.data,
      status: error.response?.status,
      statusText: error.response?.statusText
    })
    
    // 如果是 401 错误，已经在上面处理了，这里不再重复处理
    if (error.response?.status === 401) {
      // 已经在上面处理了
      return
    }
    
    ElMessage.error('订单创建失败: ' + (error.message || '未知错误'))
  } finally {
    creating.value = false
  }
}

  // 初始化
onMounted(async () => {
  // 先加载购物车数据，确保数据是最新的
  await cartStore.loadCart()
  
  // 从路由参数获取商品（从购物车跳转过来）
  if (route.query.items) {
    try {
      const itemIds = JSON.parse(route.query.items)
      checkoutItems.value = cartStore.items.filter(item => itemIds.includes(item.id))
      
      // 过滤掉明确标记为无效的商品
      const invalidItems = checkoutItems.value.filter(item => item.isValid === false)
      if (invalidItems.length > 0) {
        const invalidNames = invalidItems.map(item => item.productTitle).join('、')
        ElMessage.warning(`以下商品已失效：${invalidNames}`)
        checkoutItems.value = checkoutItems.value.filter(item => item.isValid !== false)
      }
    } catch (e) {
      console.error('解析商品参数失败:', e)
      ElMessage.error('解析商品参数失败')
    }
  } else {
    // 如果没有参数，使用购物车中选中的商品
    checkoutItems.value = cartStore.items.filter(item => item.selected)
    
    // 过滤掉明确标记为无效的商品
    checkoutItems.value = checkoutItems.value.filter(item => item.isValid !== false)
  }

  // 验证商品数据完整性
  const incompleteItems = checkoutItems.value.filter(item => !item.productId || !item.skuId)
  if (incompleteItems.length > 0) {
    console.error('商品数据不完整:', incompleteItems)
    console.error('完整购物车数据:', cartStore.items)
    ElMessage.error(`部分商品数据不完整（${incompleteItems.length}个商品），请刷新购物车后重试`)
    // 重新加载购物车
    await cartStore.loadCart()
    // 重新筛选商品
    if (route.query.items) {
      try {
        const itemIds = JSON.parse(route.query.items)
        checkoutItems.value = cartStore.items.filter(item => itemIds.includes(item.id))
      } catch (e) {
        console.error('解析商品参数失败:', e)
      }
    } else {
      checkoutItems.value = cartStore.items.filter(item => item.selected)
    }
    // 再次检查
    const stillIncomplete = checkoutItems.value.filter(item => !item.productId || !item.skuId)
    if (stillIncomplete.length > 0) {
      console.error('重新加载后仍有不完整数据:', stillIncomplete)
      ElMessage.error('商品数据异常，请重新添加商品到购物车')
      router.push('/cart')
      return
    }
  }

  // 如果没有商品，跳转回购物车
  if (checkoutItems.value.length === 0) {
    ElMessage.warning('没有可结算的商品')
    router.push('/cart')
    return
  }

  // 调试：打印商品数据
  console.log('结算商品数据:', checkoutItems.value.map(item => ({
    id: item.id,
    productId: item.productId,
    skuId: item.skuId,
    productTitle: item.productTitle,
    quantity: item.quantity
  })))

  // 加载地址列表
  await loadAddresses()
})
</script>

<style scoped>
.checkout-page {
  min-height: calc(100vh - 140px);
  padding: 20px;
  background: #f5f5f5;
}

.checkout-container {
  max-width: 1000px;
  margin: 0 auto;
}

.checkout-container h2 {
  margin-bottom: 20px;
  font-size: 24px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.order-items {
  min-height: 100px;
}

.order-item {
  display: flex;
  align-items: center;
  padding: 15px 0;
  border-bottom: 1px solid #eee;
}

.order-item:last-child {
  border-bottom: none;
}

.item-image {
  width: 80px;
  height: 80px;
  object-fit: cover;
  border-radius: 4px;
  margin-right: 15px;
}

.item-info {
  flex: 1;
}

.item-info h4 {
  margin: 0 0 5px 0;
  font-size: 14px;
}

.sku-name {
  color: #999;
  font-size: 12px;
  margin: 0;
}

.item-price {
  width: 100px;
  text-align: right;
  color: #666;
}

.item-quantity {
  width: 80px;
  text-align: center;
  color: #666;
}

.item-subtotal {
  width: 120px;
  text-align: right;
  color: #f56c6c;
  font-weight: bold;
}

.summary-item {
  display: flex;
  justify-content: space-between;
  margin-bottom: 10px;
  font-size: 14px;
}

.summary-item.total {
  font-size: 18px;
  font-weight: bold;
}

.discount {
  color: #67c23a;
}

.total-price {
  color: #f56c6c;
  font-size: 24px;
}

.address-option {
  padding: 10px;
  margin-bottom: 10px;
  border: 1px solid #eee;
  border-radius: 4px;
}

.address-option:hover {
  background: #f5f5f5;
}
</style>

