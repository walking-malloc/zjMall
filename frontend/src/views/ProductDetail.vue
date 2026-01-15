<template>
  <div class="product-detail-page">
    <el-container>
      <el-header>
        <div class="header-content">
          <div class="logo" @click="$router.push('/')">
            <h1>zjMall</h1>
          </div>
          <div class="nav-menu">
            <el-button type="text" @click="$router.push('/')">首页</el-button>
            <el-button type="text" @click="$router.push('/product/products')">商品列表</el-button>
          </div>
        </div>
      </el-header>

      <el-main>
        <div class="detail-container" v-loading="loading">
          <el-row :gutter="40" v-if="product">
            <el-col :span="12">
              <el-image
                :src="product.main_image || '/placeholder.png'"
                fit="contain"
                style="width: 100%; height: 500px;"
              />
            </el-col>
            <el-col :span="12">
              <h1>{{ product.title }}</h1>
              <p class="subtitle">{{ product.subtitle }}</p>
              <div class="price-section">
                <span class="price">¥{{ product.price || '0.00' }}</span>
              </div>
              <div class="actions">
                <el-button type="primary" size="large" @click="handleAddToCart">
                  加入购物车
                </el-button>
                <el-button type="danger" size="large" @click="handleBuyNow">
                  立即购买
                </el-button>
              </div>
            </el-col>
          </el-row>
          
          <el-empty v-if="!loading && !product" description="商品不存在" />
        </div>
      </el-main>
    </el-container>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { getProductDetail } from '@/api/product'
import { ElMessage } from 'element-plus'

const route = useRoute()
const router = useRouter()

const product = ref(null)
const loading = ref(false)

const loadProduct = async () => {
  loading.value = true
  try {
    const res = await getProductDetail(route.params.id)
    if (res.data.code === 0) {
      product.value = res.data.data
    } else {
      ElMessage.error('获取商品详情失败')
    }
  } catch (error) {
    ElMessage.error('获取商品详情失败')
  } finally {
    loading.value = false
  }
}

const handleAddToCart = () => {
  ElMessage.info('购物车功能开发中...')
}

const handleBuyNow = () => {
  ElMessage.info('立即购买功能开发中...')
}

onMounted(() => {
  loadProduct()
})
</script>

<style scoped>
.product-detail-page {
  min-height: 100vh;
}

.header-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
  height: 100%;
}

.logo h1 {
  margin: 0;
  color: #409eff;
  cursor: pointer;
}

.detail-container {
  max-width: 1200px;
  margin: 0 auto;
  padding: 20px;
}

.subtitle {
  color: #999;
  font-size: 16px;
  margin: 20px 0;
}

.price-section {
  margin: 30px 0;
}

.price {
  color: #f56c6c;
  font-size: 36px;
  font-weight: bold;
}

.actions {
  margin-top: 40px;
  display: flex;
  gap: 20px;
}
</style>

