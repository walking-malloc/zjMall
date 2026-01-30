<template>
  <div class="create-order-test">
    <el-card>
      <template #header>
        <h2>创建订单测试页面</h2>
      </template>

      <el-form :model="form" label-width="120px" style="max-width: 600px;">
        <el-form-item label="商品ID">
          <el-input v-model="form.productId" placeholder="请输入商品ID" />
        </el-form-item>

        <el-form-item label="SKU ID">
          <el-input v-model="form.skuId" placeholder="请输入SKU ID" />
        </el-form-item>

        <el-form-item label="购买数量">
          <el-input-number v-model="form.quantity" :min="1" :max="999" />
        </el-form-item>

        <el-form-item label="收货地址ID">
          <el-input v-model="form.addressId" placeholder="请输入地址ID" />
        </el-form-item>

        <el-form-item label="买家留言">
          <el-input 
            v-model="form.buyerRemark" 
            type="textarea" 
            :rows="3"
            placeholder="选填"
          />
        </el-form-item>

        <el-form-item>
          <el-button type="primary" @click="handleGenerateToken" :loading="tokenLoading">
            1. 生成Token
          </el-button>
          <el-button type="success" @click="handleCreateOrder" :loading="orderLoading" :disabled="!token">
            2. 创建订单
          </el-button>
          <el-button @click="handleReset">重置</el-button>
        </el-form-item>

        <el-form-item v-if="token" label="Token">
          <el-input :value="token" readonly>
            <template #append>
              <el-button @click="copyToken">复制</el-button>
            </template>
          </el-input>
          <div style="color: #999; font-size: 12px; margin-top: 5px;">
            有效期：{{ tokenExpireSeconds }}秒
          </div>
        </el-form-item>
      </el-form>

      <el-divider />

      <div v-if="orderResult">
        <h3>创建结果：</h3>
        <el-alert
          :type="orderResult.code === 0 ? 'success' : 'error'"
          :title="orderResult.message"
          :description="orderResult.code === 0 ? `订单号：${orderResult.orderNo}，支付金额：¥${orderResult.payAmount}` : ''"
          show-icon
          :closable="false"
        />
      </div>

      <el-divider />

      <div>
        <h3>调试信息：</h3>
        <el-collapse>
          <el-collapse-item title="请求日志" name="logs">
            <pre style="background: #f5f5f5; padding: 10px; border-radius: 4px; overflow-x: auto;">{{ logs }}</pre>
          </el-collapse-item>
        </el-collapse>
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import { generateOrderToken, createOrder } from '@/api/order'

const form = ref({
  productId: '',
  skuId: '',
  quantity: 1,
  addressId: '',
  buyerRemark: ''
})

const token = ref('')
const tokenExpireSeconds = ref(0)
const tokenLoading = ref(false)
const orderLoading = ref(false)
const orderResult = ref(null)
const logs = ref('')

const addLog = (message) => {
  const timestamp = new Date().toLocaleTimeString()
  logs.value += `[${timestamp}] ${message}\n`
}

const handleGenerateToken = async () => {
  tokenLoading.value = true
  addLog('开始生成Token...')
  
  try {
    const res = await generateOrderToken()
    addLog(`Token生成成功: ${JSON.stringify(res.data)}`)
    
    if (res.data.code === 0) {
      token.value = res.data.token
      tokenExpireSeconds.value = res.data.expire_seconds
      ElMessage.success('Token生成成功')
      addLog(`Token: ${token.value}`)
    } else {
      ElMessage.error(res.data.message || 'Token生成失败')
      addLog(`Token生成失败: ${res.data.message}`)
    }
  } catch (error) {
    addLog(`Token生成异常: ${error.message}`)
    ElMessage.error('Token生成失败: ' + error.message)
  } finally {
    tokenLoading.value = false
  }
}

const handleCreateOrder = async () => {
  if (!token.value) {
    ElMessage.warning('请先生成Token')
    return
  }

  if (!form.value.productId || !form.value.skuId || !form.value.addressId) {
    ElMessage.warning('请填写完整的订单信息')
    return
  }

  orderLoading.value = true
  orderResult.value = null
  addLog('开始创建订单...')
  addLog(`请求参数: ${JSON.stringify({
    items: [{
      product_id: form.value.productId,
      sku_id: form.value.skuId,
      quantity: form.value.quantity
    }],
    address_id: form.value.addressId,
    buyer_remark: form.value.buyerRemark,
    token: token.value
  }, null, 2)}`)

  try {
    const res = await createOrder({
      items: [{
        product_id: form.value.productId,
        sku_id: form.value.skuId,
        quantity: form.value.quantity
      }],
      addressId: form.value.addressId,
      buyerRemark: form.value.buyerRemark,
      token: token.value
    })

    addLog(`订单创建响应: ${JSON.stringify(res.data, null, 2)}`)
    orderResult.value = res.data

    if (res.data.code === 0) {
      ElMessage.success('订单创建成功！')
      addLog(`订单创建成功: 订单号=${res.data.order_no}, 支付金额=${res.data.pay_amount}`)
      // 清空Token，因为已经使用过了
      token.value = ''
    } else {
      ElMessage.error(res.data.message || '订单创建失败')
      addLog(`订单创建失败: ${res.data.message}`)
    }
  } catch (error) {
    addLog(`订单创建异常: ${error.message}`)
    ElMessage.error('订单创建失败: ' + error.message)
    orderResult.value = {
      code: 1,
      message: error.message
    }
  } finally {
    orderLoading.value = false
  }
}

const handleReset = () => {
  form.value = {
    productId: '',
    skuId: '',
    quantity: 1,
    addressId: '',
    buyerRemark: ''
  }
  token.value = ''
  tokenExpireSeconds.value = 0
  orderResult.value = null
  logs.value = ''
  ElMessage.info('已重置')
}

const copyToken = () => {
  navigator.clipboard.writeText(token.value).then(() => {
    ElMessage.success('Token已复制到剪贴板')
  }).catch(() => {
    ElMessage.error('复制失败')
  })
}
</script>

<style scoped>
.create-order-test {
  padding: 20px;
  min-height: calc(100vh - 140px);
  background: #f5f5f5;
}

.create-order-test h2 {
  margin: 0;
}

.create-order-test h3 {
  margin-top: 20px;
  margin-bottom: 10px;
}

pre {
  font-size: 12px;
  line-height: 1.5;
}
</style>

