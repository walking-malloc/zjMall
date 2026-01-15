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
          :class="{ 'default-address': address.is_default }"
        >
          <div class="address-content">
            <div class="address-info">
              <div class="address-header">
                <span class="receiver-name">{{ address.receiver_name }}</span>
                <span class="receiver-phone">{{ address.receiver_phone }}</span>
                <el-tag v-if="address.is_default" type="danger" size="small">默认</el-tag>
              </div>
              <div class="address-detail">
                {{ address.province }} {{ address.city }} {{ address.district }} {{ address.detail }}
              </div>
            </div>
            <div class="address-actions">
              <el-button 
                v-if="!address.is_default" 
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
// import { listAddresses, createAddress, updateAddress, deleteAddress, setDefaultAddress } from '@/api/user'

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
    // TODO: 调用地址API
    // const res = await listAddresses(userStore.userInfo.id)
    // if (res.data.code === 0) {
    //   addresses.value = res.data.data || []
    // }
    
    // 模拟数据
    addresses.value = []
  } catch (error) {
    ElMessage.error('获取地址列表失败')
  } finally {
    loading.value = false
  }
}

const saveAddress = async () => {
  // TODO: 调用保存地址API
  ElMessage.success(editingAddress.value ? '地址更新成功' : '地址添加成功')
  showDialog.value = false
  resetForm()
  loadAddresses()
}

const editAddress = (address) => {
  editingAddress.value = address
  addressForm.value = { ...address }
  showDialog.value = true
}

const deleteAddress = async (id) => {
  try {
    await ElMessageBox.confirm('确定要删除这个地址吗？', '提示', {
      type: 'warning'
    })
    // TODO: 调用删除地址API
    // await deleteAddress(userStore.userInfo.id, id)
    ElMessage.success('删除成功')
    loadAddresses()
  } catch (error) {
    // 用户取消
  }
}

const setDefault = async (id) => {
  // TODO: 调用设置默认地址API
  // await setDefaultAddress(userStore.userInfo.id, id)
  ElMessage.success('设置成功')
  loadAddresses()
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
}

.default-address {
  border: 2px solid #409eff;
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

