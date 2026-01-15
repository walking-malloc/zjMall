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
              <img :src="item.productImage" class="item-image" />
              <div class="item-info">
                <h4>{{ item.productTitle }}</h4>
                <p v-if="item.skuName" class="sku-name">{{ item.skuName }}</p>
                <p class="quantity">x{{ item.quantity }}</p>
              </div>
              <div class="item-price">¥{{ item.price.toFixed(2) }}</div>
            </div>
          </div>

          <div class="order-footer">
            <div class="order-total">
              共 {{ order.items.length }} 件商品，合计：<span class="total-price">¥{{ order.totalAmount.toFixed(2) }}</span>
            </div>
            <div class="order-actions">
              <el-button v-if="order.status === 'pending'" type="primary" @click="handlePay(order.id)">
                立即付款
              </el-button>
              <el-button v-if="order.status === 'shipped'" type="success" @click="handleConfirm(order.id)">
                确认收货
              </el-button>
              <el-button @click="handleViewDetail(order.id)">查看详情</el-button>
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
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'

const activeTab = ref('all')
const orders = ref([])
const loading = ref(false)
const currentPage = ref(1)
const pageSize = ref(10)
const total = ref(0)

const getStatusText = (status) => {
  const statusMap = {
    pending: '待付款',
    paid: '待发货',
    shipped: '待收货',
    completed: '已完成',
    cancelled: '已取消'
  }
  return statusMap[status] || '未知'
}

const handleTabChange = (tab) => {
  currentPage.value = 1
  loadOrders()
}

const loadOrders = async () => {
  loading.value = true
  try {
    // TODO: 调用订单API
    // const res = await getOrderList({ 
    //   status: activeTab.value === 'all' ? '' : activeTab.value,
    //   page: currentPage.value,
    //   page_size: pageSize.value
    // })
    // orders.value = res.data.data || []
    // total.value = res.data.total || 0
    
    // 模拟数据
    orders.value = []
    total.value = 0
  } catch (error) {
    ElMessage.error('获取订单列表失败')
  } finally {
    loading.value = false
  }
}

const handlePay = (orderId) => {
  ElMessage.info('支付功能开发中...')
}

const handleConfirm = (orderId) => {
  ElMessage.info('确认收货功能开发中...')
}

const handleViewDetail = (orderId) => {
  ElMessage.info('订单详情功能开发中...')
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

