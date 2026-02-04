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

          <div class="order-footer">
            <div class="order-total">
              合计：<span class="total-price">¥{{ parseFloat(order.payAmount || 0).toFixed(2) }}</span>
            </div>
            <div class="order-actions">
              <el-button v-if="getStatusNumber(order.status) === 1" type="primary" @click="handlePay(order.orderNo)">
                立即付款
              </el-button>
              <el-button v-if="getStatusNumber(order.status) === 1" @click="handleCancel(order.orderNo)">
                取消订单
              </el-button>
              <el-button v-if="getStatusNumber(order.status) === 3" type="success" @click="handleConfirm(order.orderNo)">
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

    <!-- 订单详情弹窗 -->
    <el-dialog
      v-model="detailVisible"
      title="订单详情"
      width="800px"
      :close-on-click-modal="false"
    >
      <div v-loading="detailLoading" class="order-detail">
        <div v-if="orderDetail && orderDetail.order" class="detail-content">
          <!-- 订单基本信息 -->
          <el-card class="detail-section" shadow="never">
            <template #header>
              <span class="section-title">订单信息</span>
            </template>
            <el-descriptions :column="2" border>
              <el-descriptions-item label="订单号">{{ orderDetail.order.orderNo }}</el-descriptions-item>
              <el-descriptions-item label="订单状态">
                <span :class="'status-' + orderDetail.order.status">{{ getStatusText(orderDetail.order.status) }}</span>
              </el-descriptions-item>
              <el-descriptions-item label="创建时间">
                {{ formatTime(orderDetail.order.createdAt) }}
              </el-descriptions-item>
              <el-descriptions-item label="支付时间" v-if="orderDetail.order.paidAt">
                {{ formatTime(orderDetail.order.paidAt) }}
              </el-descriptions-item>
              <el-descriptions-item label="发货时间" v-if="orderDetail.order.shippedAt">
                {{ formatTime(orderDetail.order.shippedAt) }}
              </el-descriptions-item>
              <el-descriptions-item label="完成时间" v-if="orderDetail.order.completedAt">
                {{ formatTime(orderDetail.order.completedAt) }}
              </el-descriptions-item>
            </el-descriptions>
          </el-card>

          <!-- 收货信息 -->
          <el-card class="detail-section" shadow="never">
            <template #header>
              <span class="section-title">收货信息</span>
            </template>
            <el-descriptions :column="1" border>
              <el-descriptions-item label="收货人">{{ orderDetail.order.receiverName }}</el-descriptions-item>
              <el-descriptions-item label="联系电话">{{ orderDetail.order.receiverPhone }}</el-descriptions-item>
              <el-descriptions-item label="收货地址">{{ orderDetail.order.receiverAddress }}</el-descriptions-item>
              <el-descriptions-item label="买家留言" v-if="orderDetail.order.buyerRemark">
                {{ orderDetail.order.buyerRemark || '无' }}
              </el-descriptions-item>
            </el-descriptions>
          </el-card>

          <!-- 订单明细 -->
          <el-card class="detail-section" shadow="never">
            <template #header>
              <span class="section-title">订单明细</span>
            </template>
            <div class="detail-items">
              <div 
                v-for="item in orderDetail.items" 
                :key="item.id"
                class="detail-item"
              >
                <img :src="item.productImage || '/placeholder.png'" class="detail-item-image" />
                <div class="detail-item-info">
                  <h4>{{ item.productTitle }}</h4>
                  <p v-if="item.skuName" class="detail-sku-name">{{ item.skuName }}</p>
                </div>
                <div class="detail-item-right">
                  <div class="detail-item-quantity">x{{ item.quantity }}</div>
                  <div class="detail-item-price">¥{{ parseFloat(item.price || 0).toFixed(2) }}</div>
                  <div class="detail-item-subtotal">小计：¥{{ parseFloat(item.subtotalAmount || 0).toFixed(2) }}</div>
                </div>
              </div>
            </div>
          </el-card>

          <!-- 费用明细 -->
          <el-card class="detail-section" shadow="never">
            <template #header>
              <span class="section-title">费用明细</span>
            </template>
            <div class="price-detail">
              <div class="price-row">
                <span>商品总金额：</span>
                <span>¥{{ parseFloat(orderDetail.order.totalAmount || 0).toFixed(2) }}</span>
              </div>
              <div class="price-row" v-if="parseFloat(orderDetail.order.discountAmount || 0) > 0">
                <span>优惠金额：</span>
                <span class="discount">-¥{{ parseFloat(orderDetail.order.discountAmount || 0).toFixed(2) }}</span>
              </div>
              <div class="price-row" v-if="parseFloat(orderDetail.order.shippingAmount || 0) > 0">
                <span>运费：</span>
                <span>¥{{ parseFloat(orderDetail.order.shippingAmount || 0).toFixed(2) }}</span>
              </div>
              <div class="price-row total-row">
                <span>应付金额：</span>
                <span class="total-amount">¥{{ parseFloat(orderDetail.order.payAmount || 0).toFixed(2) }}</span>
              </div>
            </div>
          </el-card>
        </div>
      </div>
      <template #footer>
        <el-button @click="detailVisible = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getOrderList, cancelOrder, getOrderDetail } from '@/api/order'

const router = useRouter()

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

// 枚举名称到数字的映射（protobuf枚举在JSON中可能被序列化为字符串）
const statusEnumMap = {
  'ORDER_STATUS_UNSPECIFIED': 0,
  'ORDER_STATUS_PENDING_PAY': 1,
  'ORDER_STATUS_PAID': 2,
  'ORDER_STATUS_SHIPPED': 3,
  'ORDER_STATUS_COMPLETED': 4,
  'ORDER_STATUS_CANCELLED': 5,
  'ORDER_STATUS_REFUNDING': 6,
  'ORDER_STATUS_REFUNDED': 7,
  'ORDER_STATUS_CLOSED': 8
}

// 将状态值转换为数字（处理枚举名称字符串、数字字符串或数字类型）
const getStatusNumber = (status) => {
  if (status === null || status === undefined) {
    return 0
  }
  
  // 如果是枚举名称字符串（如 "ORDER_STATUS_PENDING_PAY"）
  if (typeof status === 'string' && statusEnumMap[status] !== undefined) {
    return statusEnumMap[status]
  }
  
  // 如果是数字字符串或数字类型
  const num = typeof status === 'string' ? parseInt(status, 10) : Number(status)
  return isNaN(num) ? 0 : num
}

// 获取状态文本
const getStatusText = (status) => {
  const statusNum = getStatusNumber(status)
  
  // 调试日志（仅在开发环境）
  if (process.env.NODE_ENV === 'development') {
    console.log('订单状态转换:', {
      原始值: status,
      原始类型: typeof status,
      转换后数字: statusNum
    })
  }
  
  const map = {
    0: '未指定',
    1: '待付款',
    2: '待发货',
    3: '待收货',
    4: '已完成',
    5: '已取消',
    6: '退款中',
    7: '已退款',
    8: '已关闭'
  }
  
  // 如果状态值不在映射中，返回未知并记录日志
  if (map[statusNum] === undefined) {
    console.warn('未知的订单状态:', {
      原始值: status,
      原始类型: typeof status,
      转换后数字: statusNum
    })
    return '未知'
  }
  
  return map[statusNum]
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
  // 跳转到支付调试页，并携带订单号
  router.push({ name: 'Pay', query: { orderNo } })
}

const handleConfirm = (orderNo) => {
  ElMessage.info('确认收货功能开发中...')
}

const detailVisible = ref(false)
const orderDetail = ref(null)
const detailLoading = ref(false)

const handleViewDetail = async (orderNo) => {
  detailLoading.value = true
  detailVisible.value = true
  try {
    const res = await getOrderDetail(orderNo)
    if (res.data.code === 0) {
      orderDetail.value = res.data
    } else {
      ElMessage.error(res.data.message || '获取订单详情失败')
      detailVisible.value = false
    }
  } catch (error) {
    ElMessage.error('获取订单详情失败: ' + error.message)
    detailVisible.value = false
  } finally {
    detailLoading.value = false
  }
}

const formatTime = (timestamp) => {
  if (!timestamp) return '-'
  const date = new Date(timestamp)
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit'
  })
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

/* 订单详情弹窗样式 */
.order-detail {
  min-height: 200px;
}

.detail-content {
  padding: 0;
}

.detail-section {
  margin-bottom: 20px;
}

.detail-section:last-child {
  margin-bottom: 0;
}

.section-title {
  font-size: 16px;
  font-weight: bold;
}

.detail-items {
  padding: 10px 0;
}

.detail-item {
  display: flex;
  align-items: flex-start;
  padding: 15px 0;
  border-bottom: 1px solid #eee;
}

.detail-item:last-child {
  border-bottom: none;
}

.detail-item-image {
  width: 100px;
  height: 100px;
  object-fit: cover;
  border-radius: 4px;
  margin-right: 15px;
}

.detail-item-info {
  flex: 1;
  margin-right: 15px;
}

.detail-item-info h4 {
  margin: 0 0 8px 0;
  font-size: 15px;
  font-weight: 500;
}

.detail-sku-name {
  color: #999;
  font-size: 13px;
  margin: 0;
}

.detail-item-right {
  text-align: right;
  min-width: 150px;
}

.detail-item-quantity {
  color: #666;
  font-size: 14px;
  margin-bottom: 5px;
}

.detail-item-price {
  color: #666;
  font-size: 14px;
  margin-bottom: 5px;
}

.detail-item-subtotal {
  color: #f56c6c;
  font-size: 16px;
  font-weight: bold;
}

.price-detail {
  padding: 10px 0;
}

.price-row {
  display: flex;
  justify-content: space-between;
  padding: 8px 0;
  font-size: 14px;
}

.price-row .discount {
  color: #67c23a;
}

.total-row {
  border-top: 1px solid #eee;
  margin-top: 10px;
  padding-top: 15px;
  font-size: 16px;
  font-weight: bold;
}

.total-amount {
  color: #f56c6c;
  font-size: 18px;
}

.status-1 {
  color: #e6a23c;
}

.status-2 {
  color: #409eff;
}

.status-3 {
  color: #67c23a;
}

.status-4 {
  color: #909399;
}

.status-5,
.status-8 {
  color: #909399;
}

.status-6,
.status-7 {
  color: #f56c6c;
}
</style>

