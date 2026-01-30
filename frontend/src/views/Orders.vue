<template>
  <div class="orders-page">
    <div class="orders-container">
      <h2>我的订单</h2>
      
      <el-tabs v-model="activeTab" @tab-change="handleTabChange">
        <el-tab-pane label="全部" name="all"></el-tab-pane>
        <el-tab-pane label="待付款" name="pending"></el-tab-pane>
        <el-tab-pane label="待发货" name="paid"></el-tab-pane>
        <el-tab-pane label="待收货" name="shipped"></el-tab-pane>
        <el-tab-pane label="已完成" name="completed"></el-tab-pane>
      </el-tabs>

      <div class="orders-list" v-loading="loading">
        <el-empty v-if="!loading && orders.length === 0" description="暂无订单" />
        
        <el-card 
          v-for="order in orders" 
          :key="order.id" 
          class="order-card"
          style="margin-bottom: 20px;"
        >
          <template #header>
            <div class="order-header">
              <span>订单号：{{ order.orderNo }}</span>
              <span class="order-status">{{ getStatusText(order.status) }}</span>
            </div>
          </template>

          <div class="order-items">
            <div 
              v-for="item in order.items" 
              :key="item.id"
              class="order-item"
            >
              <img :src="item.productImage || '/placeholder.png'" class="item-image" />
              <div class="item-info">
                <h4>{{ item.productTitle }}</h4>
                <p v-if="item.skuName" class="sku-name">{{ item.skuName }}</p>
                <p class="quantity">x{{ item.quantity }}</p>
              </div>
              <div class="item-price">¥{{ parseFloat(item.price || 0).toFixed(2) }}</div>
            </div>
          </div>

          <div class="order-footer">
            <div class="order-total">
              共 {{ order.items ? order.items.length : 0 }} 件商品，合计：<span class="total-price">¥{{ parseFloat(order.totalAmount || 0).toFixed(2) }}</span>
            </div>
            <div class="order-actions">
              <el-button v-if="order.status === 1" type="primary" @click="handlePay(order.orderNo)">
                立即付款
              </el-button>
              <el-button v-if="order.status === 1" @click="handleCancel(order.orderNo)">
                取消订单
              </el-button>
              <el-button v-if="order.status === 3" type="success" @click="handleConfirm(order.orderNo)">
                确认收货
              </el-button>
              <el-button @click="handleViewDetail(order.orderNo)">查看详情</el-button>
            </div>
          </div>
        </el-card>
      </div>

      <el-pagination
        v-if="total > 0"
        v-model:current-page="currentPage"
        v-model:page-size="pageSize"
        :total="total"
        layout="total, prev, pager, next"
        @current-change="loadOrders"
        style="margin-top: 20px; justify-content: center;"
      />
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getOrderList, cancelOrder } from '@/api/order'

const activeTab = ref('all')
const orders = ref([])
const loading = ref(false)
const currentPage = ref(1)
const pageSize = ref(10)
const total = ref(0)

// 状态映射
const statusMap = {
  'all': 0,
  'pending': 1,  // ORDER_STATUS_PENDING_PAY
  'paid': 2,     // ORDER_STATUS_PAID
  'shipped': 3,  // ORDER_STATUS_SHIPPED
  'completed': 4 // ORDER_STATUS_COMPLETED
}

const getStatusText = (status) => {
  const map = {
    1: '待付款',
    2: '待发货',
    3: '待收货',
    4: '已完成',
    5: '已取消',
    6: '退款中',
    7: '已退款',
    8: '已关闭'
  }
  return map[status] || '未知'
}

const handleTabChange = (tab) => {
  currentPage.value = 1
  loadOrders()
}

const loadOrders = async () => {
  loading.value = true
  try {
    const status = statusMap[activeTab.value] || 0
    const res = await getOrderList({ 
      status,
      page: currentPage.value,
      pageSize: pageSize.value
    })
    
    if (res.data.code === 0) {
      orders.value = res.data.orders || []
      total.value = res.data.total || 0
    } else {
      ElMessage.error(res.data.message || '获取订单列表失败')
      orders.value = []
      total.value = 0
    }
  } catch (error) {
    ElMessage.error('获取订单列表失败: ' + error.message)
    orders.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

const handlePay = (orderNo) => {
  ElMessage.info('支付功能开发中...')
}

const handleConfirm = (orderNo) => {
  ElMessage.info('确认收货功能开发中...')
}

const handleViewDetail = async (orderNo) => {
  // TODO: 跳转到订单详情页或显示详情弹窗
  ElMessage.info('订单详情功能开发中...')
}

const handleCancel = async (orderNo) => {
  try {
    await ElMessageBox.confirm('确定要取消这个订单吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    
    const res = await cancelOrder(orderNo, '用户主动取消')
    if (res.data.code === 0) {
      ElMessage.success('订单已取消')
      loadOrders()
    } else {
      ElMessage.error(res.data.message || '取消订单失败')
    }
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('取消订单失败: ' + error.message)
    }
  }
}

onMounted(() => {
  loadOrders()
})
</script>

<style scoped>
.orders-page {
  min-height: calc(100vh - 140px);
  padding: 20px;
  background: #f5f5f5;
}

.orders-container {
  max-width: 1200px;
  margin: 0 auto;
  background: #fff;
  padding: 20px;
  border-radius: 4px;
}

.orders-container h2 {
  margin-bottom: 20px;
  font-size: 24px;
}

.order-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.order-status {
  color: #f56c6c;
  font-weight: bold;
}

.order-items {
  margin: 20px 0;
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
  margin: 0 0 5px 0;
}

.quantity {
  color: #999;
  font-size: 12px;
  margin: 0;
}

.item-price {
  color: #f56c6c;
  font-weight: bold;
  font-size: 16px;
  margin-right: 20px;
}

.order-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-top: 15px;
  border-top: 1px solid #eee;
}

.total-price {
  color: #f56c6c;
  font-size: 18px;
  font-weight: bold;
}

.order-actions {
  display: flex;
  gap: 10px;
}
</style>

