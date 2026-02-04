<template>
  <div class="pay-page">
    <div class="pay-container">
      <h2>支付调试页</h2>

      <el-card class="form-card" shadow="never">
        <template #header>
          <span>创建支付单</span>
        </template>

        <el-form :model="form" label-width="100px">
          <el-form-item label="订单号">
            <el-input
              v-model="form.orderNo"
              placeholder="请输入订单号"
              clearable
              style="max-width: 400px"
            />
          </el-form-item>

          <el-form-item label="支付渠道">
            <el-radio-group v-model="form.payChannel">
              <el-radio label="alipay">支付宝（学习模式）</el-radio>
              <el-radio label="wechat">微信（学习模式）</el-radio>
              <el-radio label="balance">余额</el-radio>
            </el-radio-group>
          </el-form-item>

          <el-form-item>
            <el-button type="primary" :loading="creating" @click="handleCreatePayment">
              创建支付单
            </el-button>
          </el-form-item>
        </el-form>
      </el-card>

      <el-card v-if="payment" class="result-card" shadow="never">
        <template #header>
          <span>支付单信息</span>
        </template>

        <el-descriptions :column="1" border>
          <el-descriptions-item label="支付单号">
            {{ payment.paymentNo }}
          </el-descriptions-item>
          <el-descriptions-item label="订单号">
            {{ payment.orderNo }}
          </el-descriptions-item>
          <el-descriptions-item label="支付金额">
            ¥{{ Number(payment.amount || 0).toFixed(2) }}
          </el-descriptions-item>
          <el-descriptions-item label="支付渠道">
            {{ payment.payChannel }}
          </el-descriptions-item>
          <el-descriptions-item label="支付链接" v-if="payUrl">
            <el-link :href="payUrl" target="_blank" type="primary">
              {{ payUrl }}
            </el-link>
          </el-descriptions-item>
        </el-descriptions>

        <div v-if="qrCode" class="qrcode-box">
          <div class="qrcode-title">模拟扫码支付链接</div>
          <el-input v-model="qrCode" readonly type="textarea" :rows="3" />
        </div>

        <div v-if="Object.keys(payParams).length" class="params-box">
          <div class="params-title">支付参数（学习模式，用于调试）</div>
          <el-input
            v-model="payParamsJson"
            type="textarea"
            :rows="8"
            readonly
          />
        </div>
      </el-card>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { createPayment } from '@/api/payment'

const route = useRoute()

const form = ref({
  orderNo: '',
  payChannel: 'alipay'
})

const creating = ref(false)
const payment = ref(null)
const payUrl = ref('')
const qrCode = ref('')
const payParams = ref({})

const payParamsJson = computed(() => {
  if (!Object.keys(payParams.value).length) return ''
  try {
    return JSON.stringify(payParams.value, null, 2)
  } catch (e) {
    return ''
  }
})

const handleCreatePayment = async () => {
  if (!form.value.orderNo) {
    ElMessage.warning('请先输入订单号')
    return
  }

  creating.value = true
  try {
    const res = await createPayment({
      orderNo: form.value.orderNo,
      payChannel: form.value.payChannel,
      returnUrl: window.location.origin + '/orders'
    })

    if (res.data.code !== 0) {
      ElMessage.error(res.data.message || '创建支付单失败')
      return
    }

    payment.value = res.data.payment || res.data.data || null
    payUrl.value = res.data.pay_url || res.data.payUrl || ''
    qrCode.value = res.data.qr_code || res.data.qrCode || ''
    payParams.value = res.data.pay_params || res.data.payParams || {}

    ElMessage.success('支付单创建成功')
  } catch (error) {
    ElMessage.error('创建支付单失败: ' + (error.message || '未知错误'))
  } finally {
    creating.value = false
  }
}

onMounted(() => {
  const orderNoFromQuery = route.query.orderNo
  if (orderNoFromQuery) {
    form.value.orderNo = orderNoFromQuery
  }
})
</script>

<style scoped>
.pay-page {
  min-height: calc(100vh - 140px);
  padding: 20px;
  background: #f5f5f5;
}

.pay-container {
  max-width: 800px;
  margin: 0 auto;
}

.pay-container h2 {
  margin-bottom: 20px;
  font-size: 24px;
}

.form-card,
.result-card {
  margin-bottom: 20px;
}

.qrcode-box,
.params-box {
  margin-top: 20px;
}

.qrcode-title,
.params-title {
  font-weight: 500;
  margin-bottom: 8px;
}
</style>

