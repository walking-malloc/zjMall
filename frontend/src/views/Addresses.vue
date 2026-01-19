<template>
  <div class="addresses-page">
    <div class="addresses-container">
      <div class="page-header">
        <h2>收货地址</h2>
        <el-button type="primary" @click="showDialog = true">新增地址</el-button>
      </div>

      <div class="addresses-list" v-loading="loading">
        <el-empty v-if="!loading && addresses.length === 0" description="暂无收货地址" />
        
        <el-card 
          v-for="address in addresses" 
          :key="address.id"
          class="address-card"
          :class="{ 'default-address': address.isDefault === true || address.isDefault === 'true' || address.isDefault === 1 }"
        >
          <div class="address-content">
            <div class="address-info">
              <div class="address-header">
                <span class="receiver-name">{{ address.receiverName }}</span>
                <span class="receiver-phone">{{ address.receiverPhone }}</span>
                <el-tag 
                  v-if="address.isDefault === true || address.isDefault === 'true' || address.isDefault === 1" 
                  type="danger" 
                  size="small"
                >
                  默认
                </el-tag>
              </div>
              <div class="address-detail">
                {{ address.province }} {{ address.city }} {{ address.district }} {{ address.detail }}
              </div>
            </div>
            <div class="address-actions">
              <el-button 
                v-if="!(address.isDefault === true || address.isDefault === 'true' || address.isDefault === 1)" 
                text 
                @click="setDefault(address.id)"
              >
                设为默认
              </el-button>
              <el-button text @click="editAddress(address)">编辑</el-button>
              <el-button type="danger" text @click="deleteAddress(address.id)">删除</el-button>
            </div>
          </div>
        </el-card>
      </div>
    </div>

    <!-- 地址编辑对话框 -->
    <el-dialog
      v-model="showDialog"
      :title="editingAddress ? '编辑地址' : '新增地址'"
      width="600px"
    >
      <el-form :model="addressForm" label-width="100px">
        <el-form-item label="收货人">
          <el-input v-model="addressForm.receiver_name" placeholder="请输入收货人姓名" />
        </el-form-item>
        <el-form-item label="手机号">
          <el-input v-model="addressForm.receiver_phone" placeholder="请输入手机号" />
        </el-form-item>
        <el-form-item label="所在地区">
          <el-row :gutter="10">
            <el-col :span="8">
              <el-input v-model="addressForm.province" placeholder="省份" />
            </el-col>
            <el-col :span="8">
              <el-input v-model="addressForm.city" placeholder="城市" />
            </el-col>
            <el-col :span="8">
              <el-input v-model="addressForm.district" placeholder="区县" />
            </el-col>
          </el-row>
        </el-form-item>
        <el-form-item label="详细地址">
          <el-input 
            v-model="addressForm.detail" 
            type="textarea" 
            :rows="3"
            placeholder="请输入详细地址"
          />
        </el-form-item>
        <el-form-item label="邮政编码">
          <el-input v-model="addressForm.postal_code" placeholder="请输入邮政编码（可选）" />
        </el-form-item>
        <el-form-item>
          <el-checkbox v-model="addressForm.is_default">设为默认地址</el-checkbox>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showDialog = false">取消</el-button>
        <el-button type="primary" @click="saveAddress">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useUserStore } from '@/stores/user'
import { getAddressList, createAddress, updateAddress, deleteAddress as deleteAddressApi, setDefaultAddress } from '@/api/user'

const userStore = useUserStore()

const addresses = ref([])
const loading = ref(false)
const showDialog = ref(false)
const editingAddress = ref(null)

const addressForm = ref({
  receiver_name: '',
  receiver_phone: '',
  province: '',
  city: '',
  district: '',
  detail: '',
  postal_code: '',
  is_default: false
})

const loadAddresses = async () => {
  loading.value = true
  try {
    // 检查是否已登录（有 token）
    const token = localStorage.getItem('token')
    console.log('loadAddresses - Token 检查:', {
      hasToken: !!token,
      tokenLength: token?.length,
      tokenPreview: token ? token.substring(0, 30) + '...' : 'null'
    })
    
    if (!token) {
      ElMessage.error('请先登录')
      loading.value = false
      return
    }

    console.log('loadAddresses - 开始请求地址列表')
    // 后端会从 token 中解析 user_id，前端不需要传 user_id
    const res = await getAddressList()
    console.log('loadAddresses - 请求成功:', res.data)
    
    if (res.data.code === 0) {
      const addressList = res.data.data || []
      // 将默认地址置顶：默认地址排在前面，其他按创建时间倒序
      addresses.value = addressList.sort((a, b) => {
        // 如果一个是默认地址，另一个不是，默认地址排在前面
        if (a.isDefault && !b.isDefault) return -1
        if (!a.isDefault && b.isDefault) return 1
        // 如果都是默认或都不是默认，按创建时间倒序（最新的在前）
        return new Date(b.createdAt) - new Date(a.createdAt)
      })
      console.log('loadAddresses - 地址列表（已排序）:', addresses.value)
    } else {
      console.error('loadAddresses - 业务错误:', res.data)
      ElMessage.error(res.data.message || '获取地址列表失败')
    }
  } catch (error) {
    console.error('loadAddresses - 请求异常:', error)
    console.error('loadAddresses - 错误详情:', {
      message: error.message,
      response: error.response?.data,
      status: error.response?.status,
      headers: error.response?.headers
    })
    
    // 如果返回的是 HTML，说明请求被前端路由拦截了
    if (error.response?.data && typeof error.response.data === 'string' && error.response.data.includes('<!DOCTYPE')) {
      ElMessage.error('请求路径错误，请检查后端服务是否运行')
    } else if (error.response?.status === 401) {
      ElMessage.error('登录已过期，请重新登录')
      // 清除 token 并跳转到登录页
      localStorage.removeItem('token')
      userStore.logout()
    } else {
      ElMessage.error('获取地址列表失败: ' + (error.response?.data?.message || error.message))
    }
  } finally {
    loading.value = false
  }
}

const saveAddress = async () => {
  try {
    if (editingAddress.value) {
      // 更新地址
      const res = await updateAddress(editingAddress.value.id, addressForm.value)
      if (res.data.code === 0) {
        ElMessage.success('地址更新成功')
        showDialog.value = false
        resetForm()
        loadAddresses()
      } else {
        ElMessage.error(res.data.message || '更新失败')
      }
    } else {
      // 创建地址
      const res = await createAddress(addressForm.value)
      if (res.data.code === 0) {
        ElMessage.success('地址添加成功')
        showDialog.value = false
        resetForm()
        loadAddresses()
      } else {
        ElMessage.error(res.data.message || '添加失败')
      }
    }
  } catch (error) {
    ElMessage.error('操作失败: ' + (error.response?.data?.message || error.message))
  }
}

const editAddress = (address) => {
  editingAddress.value = address
  // 转换字段名：后端返回的是驼峰命名，表单需要下划线命名
  addressForm.value = {
    receiver_name: address.receiverName,
    receiver_phone: address.receiverPhone,
    province: address.province,
    city: address.city,
    district: address.district,
    detail: address.detail,
    postal_code: address.postalCode,
    is_default: address.isDefault
  }
  showDialog.value = true
}

const deleteAddress = async (id) => {
  try {
    await ElMessageBox.confirm('确定要删除这个地址吗？', '提示', {
      type: 'warning'
    })
    const res = await deleteAddressApi(id)
    if (res.data.code === 0) {
      ElMessage.success('删除成功')
      loadAddresses()
    } else {
      ElMessage.error(res.data.message || '删除失败')
    }
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败: ' + (error.response?.data?.message || error.message))
    }
  }
}

const setDefault = async (id) => {
  try {
    const res = await setDefaultAddress(id)
    if (res.data.code === 0) {
      ElMessage.success('设置成功')
      loadAddresses()
    } else {
      ElMessage.error(res.data.message || '设置失败')
    }
  } catch (error) {
    ElMessage.error('设置失败: ' + (error.response?.data?.message || error.message))
  }
}

const resetForm = () => {
  editingAddress.value = null
  addressForm.value = {
    receiver_name: '',
    receiver_phone: '',
    province: '',
    city: '',
    district: '',
    detail: '',
    postal_code: '',
    is_default: false
  }
}

onMounted(() => {
  loadAddresses()
})
</script>

<style scoped>
.addresses-page {
  min-height: calc(100vh - 140px);
  padding: 20px;
  background: #f5f5f5;
}

.addresses-container {
  max-width: 1000px;
  margin: 0 auto;
  background: #fff;
  padding: 20px;
  border-radius: 4px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.page-header h2 {
  margin: 0;
  font-size: 24px;
}

.addresses-list {
  min-height: 200px;
}

.address-card {
  margin-bottom: 15px;
  transition: all 0.3s ease;
}

.address-card:hover {
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.1);
}

.default-address {
  border: 2px solid #409eff !important;
  background: linear-gradient(135deg, #e6f4ff 0%, #f0f9ff 100%) !important;
  box-shadow: 0 2px 8px rgba(64, 158, 255, 0.2) !important;
  position: relative;
}

.default-address :deep(.el-card__body) {
  background: transparent !important;
}

.default-address::before {
  content: '';
  position: absolute;
  left: 0;
  top: 0;
  bottom: 0;
  width: 4px;
  background: #409eff;
  border-radius: 4px 0 0 4px;
  z-index: 1;
}

.address-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.address-info {
  flex: 1;
}

.address-header {
  display: flex;
  align-items: center;
  gap: 15px;
  margin-bottom: 10px;
}

.receiver-name {
  font-weight: bold;
  font-size: 16px;
}

.receiver-phone {
  color: #666;
}

.address-detail {
  color: #666;
  line-height: 1.6;
}

.address-actions {
  display: flex;
  gap: 10px;
}
</style>

