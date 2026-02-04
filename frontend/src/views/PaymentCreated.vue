<template>
  <div class="payment-created-page">
    <div class="payment-created-container">
      <el-card class="result-card" shadow="never">
        <el-result
          icon="success"
          title="支付单创建成功"
          sub-title="请在支付有效期内完成支付"
        >
          <template #extra>
            <div class="info-box">
              <div class="info-row">
                <span class="label">订单号：</span>
                <span class="value">{{ orderNo || '-' }}</span>
              </div>
              <div class="info-row">
                <span class="label">支付单号：</span>
                <span class="value">{{ paymentNo || '-' }}</span>
              </div>
              <div class="info-row">
                <span class="label">支付金额：</span>
                <span class="value amount">
                  ¥{{ Number(amount || 0).toFixed(2) }}
                </span>
              </div>
            </div>

            <div class="tip-text">
              <span v-if="payUrl">
                系统将自动跳转到支付页面（{{ countdown }} 秒后），
                如未跳转请点击下方按钮。
              </span>
              <span v-else>
                您可以前往订单列表查看并继续支付（{{ countdown }} 秒后自动跳转）。
              </span>
            </div>

            <div class="btn-group">
              <el-button
                v-if="payUrl"
                type="primary"
                @click="goPay"
              >
                立即前往支付
              </el-button>
              <el-button @click="goOrders">
                查看订单
              </el-button>
              <el-button text @click="goHome">
                返回首页
              </el-button>
            </div>
          </template>
        </el-result>
      </el-card>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onBeforeUnmount } from 'vue'
import { useRoute, useRouter } from 'vue-router'

const route = useRoute()
const router = useRouter()

const paymentNo = ref(route.query.paymentNo || '')
const orderNo = ref(route.query.orderNo || '')
const amount = ref(route.query.amount || 0)
const payUrl = ref(route.query.payUrl || '')

const countdown = ref(5)
let timer = null

const goPay = () => {
  if (!payUrl.value) return
  window.open(payUrl.value, '_blank')
}

const goOrders = () => {
  router.push('/orders')
}

const goHome = () => {
  router.push('/')
}

onMounted(() => {
  timer = setInterval(() => {
    if (countdown.value <= 1) {
      clearInterval(timer)
      timer = null
      if (payUrl.value) {
        goPay()
      } else {
        goOrders()
      }
    } else {
      countdown.value -= 1
    }
  }, 1000)
})

onBeforeUnmount(() => {
  if (timer) {
    clearInterval(timer)
  }
})
</script>

<style scoped>
.payment-created-page {
  min-height: calc(100vh - 140px);
  padding: 20px;
  background: #f5f5f5;
}

.payment-created-container {
  max-width: 600px;
  margin: 0 auto;
  padding-top: 40px;
}

.result-card {
  padding: 20px 10px;
}

.info-box {
  margin-bottom: 16px;
  padding: 12px 16px;
  background: #f8f8f8;
  border-radius: 8px;
}

.info-row {
  display: flex;
  justify-content: flex-start;
  margin-bottom: 6px;
  font-size: 14px;
}

.info-row:last-child {
  margin-bottom: 0;
}

.label {
  width: 80px;
  color: #909399;
}

.value {
  color: #303133;
}

.amount {
  font-weight: 600;
  color: #f56c6c;
  font-size: 18px;
}

.tip-text {
  margin: 12px 0 20px;
  font-size: 13px;
  color: #909399;
}

.btn-group {
  display: flex;
  gap: 10px;
  justify-content: center;
}
</style>

